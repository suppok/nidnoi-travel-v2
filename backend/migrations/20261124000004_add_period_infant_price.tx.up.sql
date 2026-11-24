-- Infant price per departure (child under 2, no seat), next to the child with-bed / no-bed prices.
ALTER TABLE tour_program_period
    ADD COLUMN infant_price INTEGER NULL;
