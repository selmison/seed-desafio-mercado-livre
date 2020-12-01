CREATE TABLE users (
  id uuid NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  password VARCHAR(255),
  created_at timestamp
);