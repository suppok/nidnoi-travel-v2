-- Opt-in: email the wholesale automatically when sales save a LINE booking on the entry page.
ALTER TABLE wholesale
    ADD COLUMN auto_email BOOLEAN NOT NULL DEFAULT FALSE;
