-- Links a tour to the airline master. Optional: some tours have no flight.
-- The free-text tour_program.airline column is no longer used by the API; it stays until a later
-- migration drops it, so code deployed before this change keeps working.
ALTER TABLE tour_program ADD COLUMN airline_id UUID NULL REFERENCES airline (id);

CREATE INDEX tour_program_airline_id_idx ON tour_program (airline_id);
