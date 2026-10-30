-- Removes the seeded wholesales (created by "system"); their contacts go with them (ON DELETE CASCADE).
DELETE FROM wholesale WHERE create_by = '00000000-0000-0000-0000-000000000000';
