-- The sales person who owns the booking, picked in the LIFF form (defaults to the submitter).
-- Nullable: bookings made before this column have none.
ALTER TABLE line_booking ADD COLUMN sales_user_id UUID NULL REFERENCES app_user (id);

CREATE INDEX line_booking_sales_user_id_idx ON line_booking (sales_user_id) WHERE sales_user_id IS NOT NULL;
