-- Contact email addresses of the wholesale (lowercase, display order, max 5 enforced by the API).
ALTER TABLE wholesale
    ADD COLUMN emails TEXT[] NOT NULL DEFAULT '{}';
