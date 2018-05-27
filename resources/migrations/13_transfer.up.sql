-- transfer Table
CREATE TABLE transfer (
    id UUID PRIMARY KEY NOT NULL,
    from_account UUID REFERENCES bank_account(id),
    to_account UUID REFERENCES bank_account(id),
    amount NUMERIC(11, 2) NOT NULL,
    date timestamp NOT NULL
);
