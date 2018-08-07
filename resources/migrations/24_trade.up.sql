-- transfer Table
CREATE TYPE tstatus AS ENUM ('open', 'close');

CREATE TABLE trade (
    id UUID PRIMARY KEY NOT NULL,
    number INTEGER,
    stock_id UUID REFERENCES stock(id),
    wallet_id UUID REFERENCES wallet(id),
    opened_at TIMESTAMP NOT NULL,
    buys NUMERIC (11,2) NOT NULL,
    buy_amount NUMERIC (11,2),
    closed_at TIMESTAMP,
    sells NUMERIC (11,2),
    sell_amount NUMERIC (11,2),
    amount NUMERIC (11,2),
    dividend NUMERIC (11,2),
    capital NUMERIC (11,2),
    status tstatus,
    net NUMERIC (11,2)
);

CREATE TABLE trade_operation (
    trade_id UUID REFERENCES trade(id),
    operation_id UUID REFERENCES operation(id),
    PRIMARY KEY (trade_id, operation_id)
)
