-- Staff phone number (digits only, empty = none), shown to customers in sales tool texts.
ALTER TABLE app_user ADD COLUMN phone TEXT NOT NULL DEFAULT '';
