CREATE TABLE products (
  id uuid NOT NULL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  price NUMERIC(9,2) NOT NULL,
  amount INTEGER NOT NULL,
  description TEXT,
  category_id uuid REFERENCES categories (id),
  created_at timestamp
);
