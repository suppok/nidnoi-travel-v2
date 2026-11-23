-- Seeded rows use the nil UUID as create_by / update_by.
INSERT INTO app_config (id, code, value, model_type, description, status, create_date, update_date, create_by, update_by) VALUES
('6f1b0f3e-5c1a-4d8e-9b0a-1c2d3e4f5a60', 'APP_SETTINGS', '{"maxFileSize": 10}'::jsonb, 'APP_SETTINGS', 'General app settings',
 TRUE, now(), now(), '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000')
ON CONFLICT (code) DO NOTHING;
