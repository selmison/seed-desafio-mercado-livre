CREATE TABLE features (
  id uuid NOT NULL PRIMARY KEY,
  type_id uuid REFERENCES types_of_features (id),
  name VARCHAR(255) NOT NULL,
  details VARCHAR(255) NOT NULL
);
