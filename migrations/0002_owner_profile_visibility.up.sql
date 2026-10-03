-- Per-field visibility of the owner's public profile (GET /owners/{id}).
-- The nickname is always public; gender is hidden and the avatar is shown unless the owner changes it.
ALTER TABLE owners ADD COLUMN IF NOT EXISTS is_gender_public BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE owners ADD COLUMN IF NOT EXISTS is_avatar_public BOOLEAN NOT NULL DEFAULT true;
