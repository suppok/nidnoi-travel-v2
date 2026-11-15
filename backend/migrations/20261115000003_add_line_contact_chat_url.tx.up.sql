-- The customer's chat in LINE OA Manager (https://chat.line.biz/…), pasted by sales: its chat ID is
-- not the Messaging API user ID, so it can't be derived. '' = not known yet.
ALTER TABLE line_contact
    ADD COLUMN chat_url TEXT NOT NULL DEFAULT '';
