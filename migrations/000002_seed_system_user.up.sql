-- Seed a system user used as the default owner before auth is implemented.
-- UUID is a fixed well-known constant: all-zeros with a trailing 1.
INSERT INTO users (id, email, role, created_at)
VALUES ('00000000-0000-0000-0000-000000000001', 'system@custos.internal', 'system', NOW())
ON CONFLICT (id) DO NOTHING;
