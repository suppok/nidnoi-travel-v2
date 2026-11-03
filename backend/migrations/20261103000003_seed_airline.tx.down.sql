-- Removes the seeded airlines (created by "system"). Fails if a tour still references one (FK).
DELETE FROM airline WHERE create_by = '00000000-0000-0000-0000-000000000000';
