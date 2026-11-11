-- Sales review of the finance documents: approve (and push them to the customer's LINE chat when
-- one is linked) or send them back to finance with a note.
ALTER TABLE line_booking
    ADD COLUMN documents_approved_at   TIMESTAMPTZ NULL,
    ADD COLUMN documents_approved_by   UUID        NULL REFERENCES app_user (id),
    -- When the documents card was pushed to the customer, and to which chat (NULL = not pushed).
    ADD COLUMN customer_sent_at        TIMESTAMPTZ NULL,
    ADD COLUMN customer_sent_contact_id UUID       NULL REFERENCES line_contact (id),
    -- The latest fix request from sales to finance.
    ADD COLUMN fix_note                TEXT        NOT NULL DEFAULT '',
    ADD COLUMN fix_requested_at        TIMESTAMPTZ NULL;
