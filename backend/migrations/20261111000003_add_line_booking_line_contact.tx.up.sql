-- The customer's LINE OA chat (line_contact), picked by staff, so a receipt can be pushed to it later.
-- Nullable: older bookings and bookings whose customer never chatted with the OA have none.
ALTER TABLE line_booking ADD COLUMN line_contact_id UUID NULL REFERENCES line_contact (id);

CREATE INDEX line_booking_line_contact_id_idx ON line_booking (line_contact_id) WHERE line_contact_id IS NOT NULL;
