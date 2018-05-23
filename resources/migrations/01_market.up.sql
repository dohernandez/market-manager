-- market Table
CREATE TABLE market (
    id UUID PRIMARY KEY NOT NULL,
    name text NOT NULL,
    display_name text NOT NULL
);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO market (id, name, display_name) VALUES
(uuid_generate_v4(), 'stock', 'Stock'),
(uuid_generate_v4(), 'cryptocurrency', 'Cryptocurrency');
