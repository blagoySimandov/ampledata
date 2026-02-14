DROP INDEX IF EXISTS idx_users_stripe_subscription_id;
ALTER TABLE users DROP COLUMN IF EXISTS subscription_tier;
ALTER TABLE users DROP COLUMN IF EXISTS stripe_subscription_id;
ALTER TABLE users DROP COLUMN IF EXISTS tokens_included;
ALTER TABLE users DROP COLUMN IF EXISTS current_period_start;
ALTER TABLE users DROP COLUMN IF EXISTS current_period_end;
ALTER TABLE users ADD COLUMN tokens_purchased BIGINT NOT NULL DEFAULT 0;
