-- Add new column.
ALTER TABLE stock ADD type UUID REFERENCES stock_info(id);
ALTER TABLE stock ADD sector UUID REFERENCES stock_info(id);
ALTER TABLE stock ADD industry UUID REFERENCES stock_info(id);
