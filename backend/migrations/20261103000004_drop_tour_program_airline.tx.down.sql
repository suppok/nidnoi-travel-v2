-- Restores the column empty; the old free-text values are not recovered.
ALTER TABLE tour_program ADD COLUMN airline TEXT NOT NULL DEFAULT '';
