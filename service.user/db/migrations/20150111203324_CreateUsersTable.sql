
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  email TEXT NOT NULL,
  auth_token TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX email_index ON users (email);
CREATE UNIQUE INDEX auth_token_index ON users (auth_token);

CREATE TRIGGER set_updated_at BEFORE UPDATE ON users FOR EACH ROW
  EXECUTE PROCEDURE set_updated_at();

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back
DROP TRIGGER set_updated_at ON users;
DROP TABLE users;
