CREATE TABLE types_of_features (
  id uuid NOT NULL PRIMARY KEY,
  product_id uuid REFERENCES products (id),
  type VARCHAR(255) NOT NULL
);
