-- banking Table
CREATE TYPE eaccountnotype AS ENUM ('iban', 'hash');

CREATE TABLE bank_account (
    id UUID PRIMARY KEY NOT NULL,
    name VARCHAR(180) NOT NULL,
    account_no TEXT NOT NULL,
    alias VARCHAR(30) NOT NULL,
    account_no_type eaccountnotype,
    UNIQUE (alias)
);

CREATE INDEX bank_account_alias_idx ON bank_account (alias);
