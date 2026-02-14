ALTER TABLE users DROP COLUMN IF EXISTS tokens_purchased;
ALTER TABLE users ADD COLUMN subscription_tier VARCHAR(50);
ALTER TABLE users ADD COLUMN stripe_subscription_id VARCHAR(255);
ALTER TABLE users ADD COLUMN tokens_included BIGINT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN current_period_start TIMESTAMPTZ;
ALTER TABLE users ADD COLUMN current_period_end TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_users_stripe_subscription_id ON users (stripe_subscription_id);
