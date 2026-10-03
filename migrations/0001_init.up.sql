-- Tailverse database schema (PostgreSQL + PostGIS).
-- Based on database_schema.sql from the project notes, extended with what the REST API needs:
--   owners.email/password_hash (auth), pets.approx_address, services.title/category.

CREATE EXTENSION IF NOT EXISTS postgis;

-- Owner (human account)
CREATE TABLE IF NOT EXISTS owners (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             TEXT NOT NULL,
    password_hash     TEXT NOT NULL,
    nickname          TEXT NOT NULL,
    gender            TEXT,               -- nullable, user can hide
    avatar_url        TEXT,
    is_profile_public BOOLEAN NOT NULL DEFAULT true,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_owners_email ON owners (lower(email));

-- Pet (the actual "user" of the social network)
CREATE TABLE IF NOT EXISTS pets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    species         TEXT NOT NULL,        -- 'dog','cat',... auto-derived from breed if possible
    breed           TEXT,
    birth_date      DATE,
    approx_address  TEXT,                 -- district/street, never the exact address
    approx_location geography(Point,4326),
    avatar_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_pets_owner ON pets(owner_id);

-- Walk spots (parks, dog areas, etc.)
CREATE TABLE IF NOT EXISTS walk_spots (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    location   geography(Point,4326) NOT NULL,
    tags       TEXT[] NOT NULL DEFAULT '{}',   -- e.g. {'fenced','water','off-leash'}
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_walk_spots_geo ON walk_spots USING GIST(location);

-- Walk announcements ("Иду гулять")
CREATE TABLE IF NOT EXISTS walk_announcements (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pet_id       UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    spot_id      UUID REFERENCES walk_spots(id),      -- nullable if free-form point
    custom_point geography(Point,4326),               -- used if spot_id is null
    starts_at    TIMESTAMPTZ NOT NULL,
    duration_min INT NOT NULL CHECK (duration_min > 0),
    status       TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'finished')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT walk_announcements_place CHECK ((spot_id IS NULL) <> (custom_point IS NULL))
);
CREATE INDEX IF NOT EXISTS idx_announcements_time ON walk_announcements(starts_at);
CREATE INDEX IF NOT EXISTS idx_announcements_spot ON walk_announcements(spot_id);
CREATE INDEX IF NOT EXISTS idx_announcements_point ON walk_announcements USING GIST(custom_point);

-- Who joined an announcement
CREATE TABLE IF NOT EXISTS announcement_participants (
    announcement_id UUID NOT NULL REFERENCES walk_announcements(id) ON DELETE CASCADE,
    pet_id          UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (announcement_id, pet_id)
);

-- Feed posts (reviews/photos after a walk)
CREATE TABLE IF NOT EXISTS posts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pet_id     UUID NOT NULL REFERENCES pets(id) ON DELETE CASCADE,
    spot_id    UUID REFERENCES walk_spots(id),   -- nullable, TBD scoping decision
    text       TEXT NOT NULL DEFAULT '',
    photo_urls TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_posts_spot ON posts(spot_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_created ON posts(created_at DESC, id DESC);

-- Services (grooming, paid walking, etc.)
CREATE TABLE IF NOT EXISTS services (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_owner_id UUID NOT NULL REFERENCES owners(id) ON DELETE CASCADE,
    title             TEXT NOT NULL,
    category          TEXT NOT NULL,        -- 'grooming','dog_walking',...
    description       TEXT NOT NULL DEFAULT '',
    location          geography(Point,4326) NOT NULL,
    price             NUMERIC(12, 2) CHECK (price >= 0),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_services_geo ON services USING GIST(location);
CREATE INDEX IF NOT EXISTS idx_services_category ON services(category);
