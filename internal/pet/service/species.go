package pet_service

import "strings"

// breedKeywords maps a lower-case fragment of a breed name to the species.
// It covers popular breeds only: when nothing matches the owner has to send "species" explicitly.
var breedKeywords = []struct {
	keyword string
	species string
}{
	// dogs
	{"retriever", "dog"}, {"ретривер", "dog"},
	{"terrier", "dog"}, {"терьер", "dog"},
	{"shepherd", "dog"}, {"овчарка", "dog"},
	{"spaniel", "dog"}, {"спаниель", "dog"},
	{"poodle", "dog"}, {"пудель", "dog"},
	{"bulldog", "dog"}, {"бульдог", "dog"},
	{"husky", "dog"}, {"хаски", "dog"},
	{"labrador", "dog"}, {"лабрадор", "dog"},
	{"corgi", "dog"}, {"корги", "dog"},
	{"beagle", "dog"}, {"бигль", "dog"},
	{"dachshund", "dog"}, {"такса", "dog"},
	{"chihuahua", "dog"}, {"чихуахуа", "dog"},
	{"pug", "dog"}, {"мопс", "dog"},
	{"shiba", "dog"}, {"сиба", "dog"},
	{"spitz", "dog"}, {"шпиц", "dog"},
	{"collie", "dog"}, {"колли", "dog"},
	{"doberman", "dog"}, {"доберман", "dog"},
	{"rottweiler", "dog"}, {"ротвейлер", "dog"},
	{"malamute", "dog"}, {"маламут", "dog"},
	{"samoyed", "dog"}, {"самоед", "dog"},
	{"schnauzer", "dog"}, {"шнауцер", "dog"},
	{"boxer", "dog"}, {"боксёр", "dog"}, {"боксер", "dog"},
	{"hound", "dog"}, {"гончая", "dog"},
	{"mastiff", "dog"}, {"мастиф", "dog"},
	{"dog", "dog"}, {"собака", "dog"},
	// cats
	{"maine coon", "cat"}, {"мейн-кун", "cat"}, {"мейн кун", "cat"},
	{"british shorthair", "cat"}, {"британская", "cat"},
	{"scottish fold", "cat"}, {"шотландская", "cat"},
	{"sphynx", "cat"}, {"сфинкс", "cat"},
	{"siamese", "cat"}, {"сиамская", "cat"},
	{"persian", "cat"}, {"персидская", "cat"},
	{"bengal", "cat"}, {"бенгальская", "cat"},
	{"ragdoll", "cat"}, {"рэгдолл", "cat"},
	{"russian blue", "cat"}, {"русская голубая", "cat"},
	{"cat", "cat"}, {"кошка", "cat"}, {"кот", "cat"},
}

// SpeciesByBreed returns the species for a known breed, or "" when it cannot be derived.
func SpeciesByBreed(breed string) string {
	breed = strings.ToLower(strings.TrimSpace(breed))
	if breed == "" {
		return ""
	}
	for _, entry := range breedKeywords {
		if strings.Contains(breed, entry.keyword) {
			return entry.species
		}
	}
	return ""
}
