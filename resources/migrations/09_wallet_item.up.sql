-- wallet_item Table
CREATE TABLE wallet_item (
    id UUID PRIMARY KEY NOT NULL,
    wallet_id UUID REFERENCES wallet(id),
    stock_id UUID REFERENCES stock(id),
    amount INTEGER,
    invested NUMERIC(11, 2) NOT NULL,
    dividend NUMERIC(7, 4),
    buys NUMERIC(11, 2) NOT NULL,
    sells NUMERIC(11, 2)
);
