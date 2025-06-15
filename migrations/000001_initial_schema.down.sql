-- Drop tables in reverse order due to foreign key constraints
DROP TABLE IF EXISTS user_tokens;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS facilities;

-- Drop the UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";