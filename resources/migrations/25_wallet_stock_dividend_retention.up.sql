-- wallet_stock_dividend_retention Table
CREATE TABLE wallet_stock_dividend_retention (
    wallet_id UUID REFERENCES wallet(id),
    stock_id UUID REFERENCES stock(id),
    retention NUMERIC (11,4),
    date TIMESTAMP,
    PRIMARY KEY (wallet_id, stock_id, date)
);
