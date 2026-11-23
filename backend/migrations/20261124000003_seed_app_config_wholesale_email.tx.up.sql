-- Reply-To and Cc of the booking email to wholesales: this inbox gets every reply and a copy.
-- Seeded rows use the nil UUID as create_by / update_by.
INSERT INTO app_config (id, code, value, model_type, description, status, create_date, update_date, create_by, update_by) VALUES
('2b7c4e1a-8d3f-4a6b-9c5e-0f1a2b3c4d5e', 'WHOLESALE_EMAIL',
 '{"replyTo": "supanat.pokturng@gmail.com", "cc": ["supanat.pokturng@gmail.com"]}'::jsonb, 'WHOLESALE_EMAIL',
 'Reply-To and Cc of the booking email to wholesales',
 TRUE, now(), now(), '00000000-0000-0000-0000-000000000000', '00000000-0000-0000-0000-000000000000')
ON CONFLICT (code) DO NOTHING;
