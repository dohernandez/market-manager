-- Account Table
CREATE TYPE doperation AS ENUM ('buy', 'sell', 'connectivity', 'dividend', 'interest');

CREATE TABLE account (
    id UUID PRIMARY KEY NOT NULL,
    date timestamp NOT NULL,
    stock_id UUID,
    operation doperation,
    amount INTEGER,
    price NUMERIC(7, 2) NOT NULL,
    price_change NUMERIC(7, 4),
    price_change_commission NUMERIC(7, 2),
    value NUMERIC(7, 2) NOT NULL,
    commission NUMERIC(7, 2)
);

CREATE INDEX account_operation_idx ON account (operation);
