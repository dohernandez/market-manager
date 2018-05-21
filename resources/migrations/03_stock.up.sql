-- Stock Table
CREATE TABLE stock  (
    id UUID PRIMARY KEY NOT NULL,
    market_id UUID REFERENCES market(id),
    exchange_id UUID REFERENCES exchange(id),
    name VARCHAR(120) NOT NULL,
    symbol VARCHAR(10) NOT NULL,
    value NUMERIC(7, 3) DEFAULT 0
);
