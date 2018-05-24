-- wallet_item Table
CREATE TABLE wallet_item (
    id UUID PRIMARY KEY NOT NULL,
    wallet_id UUID REFERENCES wallet(id),
    stock_id UUID REFERENCES stock(id),
    amount INTEGER,
    invested NUMERIC(7, 2) NOT NULL,
    dividend NUMERIC(7, 4),
    buys NUMERIC(7, 2) NOT NULL,
    sells NUMERIC(7, 2)
);
