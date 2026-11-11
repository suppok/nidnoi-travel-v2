UPDATE line_booking SET state = 20 WHERE state = 25;
DROP INDEX IF EXISTS line_booking_finance_user_idx;
ALTER TABLE line_booking DROP COLUMN IF EXISTS confirmed_at, DROP COLUMN IF EXISTS finance_user_id;
DROP TABLE IF EXISTS sales_finance;
