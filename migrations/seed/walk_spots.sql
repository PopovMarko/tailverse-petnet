-- Sample walk spots for local development (the API has no endpoint to create spots yet).
-- Coordinates are in Moscow and Kryvyi Rih; replace them with places near you. Safe to run repeatedly.
-- Note ST_MakePoint takes (lng, lat).
INSERT INTO walk_spots (id, name, location, tags) VALUES
    ('11111111-1111-4111-8111-111111111111', 'Парк Горького — собачья площадка', ST_SetSRID(ST_MakePoint(37.6010, 55.7298), 4326)::geography, '{fenced,water}'),
    ('22222222-2222-4222-8222-222222222222', 'Нескучный сад',                    ST_SetSRID(ST_MakePoint(37.5890, 55.7210), 4326)::geography, '{off-leash}'),
    ('33333333-3333-4333-8333-333333333333', 'Сокольники — площадка для выгула', ST_SetSRID(ST_MakePoint(37.6720, 55.7920), 4326)::geography, '{fenced}'),
    ('44444444-4444-4444-8444-444444444444', 'Измайловский парк',                ST_SetSRID(ST_MakePoint(37.7700, 55.7720), 4326)::geography, '{water,off-leash}'),
    ('55555555-5555-4555-8555-555555555555', 'Кривой Рог — площадка для выгула', ST_SetSRID(ST_MakePoint(33.390509, 47.905549), 4326)::geography, '{}')
ON CONFLICT (id) DO NOTHING;
