DROP INDEX IF EXISTS tour_program_period_fire_sale_idx;
ALTER TABLE tour_program_period
    DROP CONSTRAINT IF EXISTS tour_program_period_fire_sale_discount_check,
    DROP CONSTRAINT IF EXISTS tour_program_period_fire_sale_order_check,
    DROP COLUMN IF EXISTS fire_sale_order,
    DROP COLUMN IF EXISTS fire_sale;
