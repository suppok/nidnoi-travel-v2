-- Drops the free-text airline column, replaced by tour_program.airline_id (airline master).
-- Run only after the code that stopped using it is deployed: older code still reads and writes it.
ALTER TABLE tour_program DROP COLUMN airline;
