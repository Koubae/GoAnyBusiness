CREATE EXTENSION IF NOT EXISTS pgcrypto;   -- gen_random_uuid(), digest(), etc.
CREATE EXTENSION IF NOT EXISTS citext;     -- case-insensitive text (perfect for email)