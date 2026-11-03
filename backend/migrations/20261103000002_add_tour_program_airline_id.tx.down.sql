DROP INDEX IF EXISTS tour_program_airline_id_idx;
ALTER TABLE tour_program DROP COLUMN IF EXISTS airline_id;
