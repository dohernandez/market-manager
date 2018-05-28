-- Add new column.
ALTER TABLE wallet_item ADD capital NUMERIC(11, 2) DEFAULT 0;
ALTER TABLE wallet_item ADD capital_rate NUMERIC(6, 4) DEFAULT 0;
