-- transfer Table
CREATE TABLE transfer (
    id UUID PRIMARY KEY NOT NULL,
    from UUID REFERENCES bank_account(id),
    to UUID REFERENCES bank_account(id),
    amount NUMERIC(7, 2) NOT NULL,
    date timestamp NOT NULL
);
