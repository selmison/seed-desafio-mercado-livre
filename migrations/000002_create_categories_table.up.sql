CREATE TABLE categories (
  id uuid NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE
);