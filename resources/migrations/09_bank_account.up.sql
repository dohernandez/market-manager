-- bank_account Table
CREATE TABLE bank_account (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR(180) NOT NULL,
    iban VARCHAR(32) NOT NULL,
    alias VARCHAR(30) NOT NULL,
    UNIQUE (alias)
);

CREATE INDEX bank_account_alias_idx ON bank_account (alias);
