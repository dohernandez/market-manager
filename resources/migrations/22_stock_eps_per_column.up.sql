-- Add new column.
ALTER TABLE stock ADD eps NUMERIC(7, 2) DEFAULT 0;
ALTER TABLE stock ADD per NUMERIC(7, 2) DEFAULT 0;
