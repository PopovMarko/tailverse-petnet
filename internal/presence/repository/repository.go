package presence_repository

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	core_domain "github.com/PopovMarko/tailverse-petnet/internal/core/domain"
	"github.com/redis/go-redis/v9"
)

// Redis layout:
//
//	spot:{spot_id}:present  ZSET  member = pet_id, score = check-in time (unix ms)
//	pet:{pet_id}:location   STRING spot_id, with the presence TTL — where the pet is right now
//
// A member expires once its score is older than the TTL; expired members are trimmed on every read.
// The key-level EXPIRE only cleans up spots nobody has visited for a whole TTL.
type PresenceRepository struct {
	client *redis.Client
}

func NewPresenceRepository(client *redis.Client) *PresenceRepository {
	return &PresenceRepository{client: client}
}

func spotKey(spotId string) string { return "spot:" + spotId + ":present" }
func petKey(petId string) string   { return "pet:" + petId + ":location" }

// CheckIn marks the pet as present at spotId and returns the spot it was at before ("" when none).
func (r *PresenceRepository) CheckIn(ctx context.Context, spotId, petId string, at time.Time, ttl time.Duration) (string, error) {
	previousSpotId, err := r.client.Get(ctx, petKey(petId)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("get pet location: %w", err)
	}

	_, err = r.client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		if previousSpotId != "" && previousSpotId != spotId {
			pipe.ZRem(ctx, spotKey(previousSpotId), petId)
		}
		pipe.ZAdd(ctx, spotKey(spotId), redis.Z{Score: float64(at.UnixMilli()), Member: petId})
		pipe.Expire(ctx, spotKey(spotId), ttl)
		pipe.Set(ctx, petKey(petId), spotId, ttl)
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("check in: %w", err)
	}

	if previousSpotId == spotId {
		return "", nil
	}
	return previousSpotId, nil
}

// CheckOut removes the pet from the spot. It returns false when the pet was not there.
func (r *PresenceRepository) CheckOut(ctx context.Context, spotId, petId string) (bool, error) {
	removed, err := r.client.ZRem(ctx, spotKey(spotId), petId).Result()
	if err != nil {
		return false, fmt.Errorf("check out: %w", err)
	}

	// Delete the location only if it still points at this spot.
	location, err := r.client.Get(ctx, petKey(petId)).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return false, fmt.Errorf("get pet location: %w", err)
	}
	if location == spotId {
		if err := r.client.Del(ctx, petKey(petId)).Err(); err != nil {
			return false, fmt.Errorf("delete pet location: %w", err)
		}
	}
	return removed > 0, nil
}

// CountPresent returns the number of pets present at every spot in one round trip.
func (r *PresenceRepository) CountPresent(ctx context.Context, spotIds []string, now time.Time, ttl time.Duration) (map[string]int, error) {
	if len(spotIds) == 0 {
		return map[string]int{}, nil
	}

	cutoff := expiredCutoff(now, ttl)
	counts := make([]*redis.IntCmd, len(spotIds))
	_, err := r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for i, spotId := range spotIds {
			pipe.ZRemRangeByScore(ctx, spotKey(spotId), "-inf", cutoff)
			counts[i] = pipe.ZCard(ctx, spotKey(spotId))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("count present: %w", err)
	}

	result := make(map[string]int, len(spotIds))
	for i, spotId := range spotIds {
		result[spotId] = int(counts[i].Val())
	}
	return result, nil
}

func (r *PresenceRepository) ListPresent(ctx context.Context, spotId string, now time.Time, ttl time.Duration) ([]core_domain.PresenceEntry, error) {
	var members *redis.ZSliceCmd
	_, err := r.client.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.ZRemRangeByScore(ctx, spotKey(spotId), "-inf", expiredCutoff(now, ttl))
		members = pipe.ZRangeWithScores(ctx, spotKey(spotId), 0, -1)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list present: %w", err)
	}

	entries := make([]core_domain.PresenceEntry, 0, len(members.Val()))
	for _, member := range members.Val() {
		petId, ok := member.Member.(string)
		if !ok {
			continue
		}
		checkedInAt := time.UnixMilli(int64(member.Score)).UTC()
		entries = append(entries, core_domain.PresenceEntry{
			SpotId:      spotId,
			PetId:       petId,
			CheckedInAt: checkedInAt,
			ExpiresAt:   checkedInAt.Add(ttl),
		})
	}
	return entries, nil
}

// expiredCutoff is the inclusive upper score bound of check-ins older than ttl.
func expiredCutoff(now time.Time, ttl time.Duration) string {
	return strconv.FormatInt(now.Add(-ttl).UnixMilli(), 10)
}
