-- wallet_bank_account Table
CREATE TABLE wallet_bank_account (
    wallet_id UUID REFERENCES wallet(id),
    bank_account_id UUID REFERENCES bank_account(id),
    PRIMARY KEY(wallet_id, bank_account_id)
);
