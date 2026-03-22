-- +goose up
ALTER TABLE users
ADD COLUMN password TEXT NOT NULL DEFAULT 'unset';


-- +goose down
ALTER TABLE users
DROP COLUMN password;