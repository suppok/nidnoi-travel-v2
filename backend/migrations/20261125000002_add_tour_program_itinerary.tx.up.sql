-- Short day-by-day timeline of the places a tour visits: [{"day": 1, "places": ["สนามบินสุวรรณภูมิ", ...]}, ...].
ALTER TABLE tour_program
    ADD COLUMN itinerary JSONB NOT NULL DEFAULT '[]';
