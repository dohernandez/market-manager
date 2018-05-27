-- wallet Table
CREATE TABLE wallet (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR(120) NOT NULL,
    url TEXT NOT NULL,
    invested NUMERIC(11, 2) DEFAULT 0,
    capital NUMERIC(11, 2) DEFAULT 0,
    funds NUMERIC(11, 2) DEFAULT 0
);


