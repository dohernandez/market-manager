-- Stock Dividend Table
CREATE TYPE dstatus AS ENUM ('projected', 'announced', 'payed');

CREATE TABLE stock_dividend  (
    stock_id UUID REFERENCES stock(id),
    ex_date timestamp NOT NULL,
    payment_date timestamp,
    record_date timestamp,
    status dstatus DEFAULT 'projected',
    amount NUMERIC(7, 4) DEFAULT 0,
    change_from_prev NUMERIC(7, 2) DEFAULT 0,
    change_from_prev_year NUMERIC(7, 2) DEFAULT 0,
    prior_12_months_yield NUMERIC(7, 2) DEFAULT 0,
    UNIQUE (stock_id, ex_date)
);

CREATE INDEX stock_dividend_stock_id_date_idx ON stock_dividend (stock_id, ex_date);
