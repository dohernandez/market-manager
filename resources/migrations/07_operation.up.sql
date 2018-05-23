-- operation Table
CREATE TYPE eaction AS ENUM ('buy', 'sell', 'connectivity', 'dividend', 'interest');

CREATE TABLE operation (
    id UUID PRIMARY KEY NOT NULL,
    date timestamp NOT NULL,
    stock_id UUID,
    action eaction,
    amount INTEGER,
    price NUMERIC(7, 2) NOT NULL,
    price_change NUMERIC(7, 4),
    price_change_commission NUMERIC(7, 2),
    value NUMERIC(7, 2) NOT NULL,
    commission NUMERIC(7, 2)
);

CREATE INDEX operation_type_idx ON operation (action);
