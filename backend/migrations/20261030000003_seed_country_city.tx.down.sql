-- Removes the seeded rows (created by "system"); rows added by staff are kept.
DELETE FROM city WHERE create_by = '00000000-0000-0000-0000-000000000000';
DELETE FROM country WHERE create_by = '00000000-0000-0000-0000-000000000000';
