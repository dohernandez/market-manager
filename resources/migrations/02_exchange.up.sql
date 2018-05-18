-- Exchange Table
CREATE TABLE exchange (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR(120) NOT NULL,
    symbol VARCHAR(10) NOT NULL
);

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

INSERT INTO exchange (id, name, symbol) VALUES
(uuid_generate_v4(), 'NASDAQ COMPOSITE', 'NASDAQ'),
(uuid_generate_v4(), 'New York Stock Exchange', 'NYSE'),
(uuid_generate_v4(), 'Bolsas y Mercados Españoles', 'BME'),
(uuid_generate_v4(), 'Frankfurter Wertpapierbörse', 'FRA'),
(uuid_generate_v4(), 'Borsa Italiana S.p. A', 'BIT');
