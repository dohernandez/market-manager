-- Add new column.
ALTER TABLE stock ADD hv_20_day NUMERIC(5, 2) DEFAULT 0;
ALTER TABLE stock ADD hv_52_week NUMERIC(5, 2) DEFAULT 0;
ALTER TABLE stock ADD price_volatility_update timestamp;
