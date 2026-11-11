-- Finance documents for a confirmed LINE booking (PDF URLs under R2 bookings/documents/):
-- deposit -> AI (ใบรับเงินมัดจำ) + IV (ใบแจ้งหนี้); paid in full -> RE (ใบเสร็จรับเงิน).
-- documents_at is set when finance first submits them; NULL on a confirmed booking = waiting for finance.
ALTER TABLE line_booking
    ADD COLUMN ai_file      TEXT        NOT NULL DEFAULT '',
    ADD COLUMN iv_file      TEXT        NOT NULL DEFAULT '',
    ADD COLUMN re_file      TEXT        NOT NULL DEFAULT '',
    ADD COLUMN documents_at TIMESTAMPTZ NULL,
    ADD COLUMN documents_by UUID        NULL REFERENCES app_user (id);
CREATE INDEX line_booking_documents_at_idx ON line_booking (sales_user_id, documents_at DESC) WHERE documents_at IS NOT NULL;
