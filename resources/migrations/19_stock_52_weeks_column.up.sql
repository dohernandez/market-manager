-- Add new column.
ALTER TABLE stock ADD high_52_week NUMERIC(11, 2) DEFAULT 0;
ALTER TABLE stock ADD low_52_week NUMERIC(11, 2) DEFAULT 0;
ALTER TABLE stock ADD high_low_52_Week_update timestamp;
