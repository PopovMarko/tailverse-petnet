-- Tailverse database schema (PostgreSQL + PostGIS)

-- Owner (human account)
CREATE TABLE IF NOT EXISTS owners (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    nickname     TEXT NOT NULL,
    gender       TEXT,               -- nullable, user can hide
    avatar_url   TEXT,
    is_profile_public BOOLEAN DEFAULT true,
    created_at   TIMESTAMPTZ DEFAULT now()
);

-- Pet (the actual "user" of the social network)
CREATE TABLE IF NOT EXISTS pets (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id     UUID NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
    name         TEXT NOT NULL,
    species      TEXT NOT NULL,       -- 'dog','cat',... auto-derived from breed if possible
    breed        TEXT,
    birth_date   DATE,
    approx_location geography(Point,4326),  -- district/street-level, not exact
    avatar_url   TEXT,
    created_at   TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_pets_owner ON pets(owner_id);

-- Walk spots (parks, dog areas, etc.)
CREATE TABLE IF NOT EXISTS walk_spots (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         TEXT NOT NULL,
    location     geography(Point,4326) NOT NULL,
    tags         TEXT[],              -- e.g. {'fenced','water','off-leash'}
    created_at   TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_walk_spots_geo ON walk_spots USING GIST(location);

-- Walk announcements ("Иду гулять")
CREATE TABLE IF NOT EXISTS walk_announcements (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pet_id       UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    spot_id      UUID REFERENCES walk_spots(id),      -- nullable if free-form point
    custom_point geography(Point,4326),                -- used if spot_id is null
    starts_at    TIMESTAMPTZ NOT NULL,
    duration_min INT NOT NULL,
    status       TEXT NOT NULL DEFAULT 'active',       -- active/cancelled/finished
    created_at   TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_announcements_time ON walk_announcements(starts_at);

-- Who joined an announcement
CREATE TABLE IF NOT EXISTS announcement_participants (
    announcement_id UUID NOT NULL REFERENCES walk_announcements(id) ON DELETE CASCADE,
    pet_id          UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    joined_at       TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (announcement_id, pet_id)
);

-- Feed posts (reviews/photos after a walk)
CREATE TABLE IF NOT EXISTS posts (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pet_id       UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    spot_id      UUID REFERENCES walk_spots(id),   -- nullable, TBD scoping decision
    text         TEXT,
    photo_urls   TEXT[],
    created_at   TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_posts_spot ON posts(spot_id);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC);

-- Services (grooming, paid walking, etc.)
CREATE TABLE IF NOT EXISTS services (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_owner_id UUID NOT NULL REFERENCES owners(id),
    type         TEXT NOT NULL,        -- 'grooming','walking',...
    description  TEXT,
    location     geography(Point,4326),
    price        NUMERIC,
    created_at   TIMESTAMPTZ DEFAULT now()
);
