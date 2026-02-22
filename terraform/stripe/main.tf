# ─────────────────────────────────────────────
# Billing Meter
# ─────────────────────────────────────────────
# NOTE: stripe_billing_meter support depends on your provider version.
# If "stripe_billing_meter" is not yet available in lukasaron/stripe,
# create the meter once via the Stripe CLI:
#
#   stripe billing_meters create \
#     --display-name="Enrichment Credits" \
#     --event-name="enrichment_credits" \
#     --default-aggregation[formula]=sum \
#     --customer-mapping[event-payload-key]=stripe_customer_id \
#     --customer-mapping[type]=by_id \
#     --value-settings[event-payload-key]=value
#
# Then import it: terraform import stripe_billing_meter.enrichment_credits <meter_id>

resource "stripe_billing_meter" "enrichment_credits" {
  display_name = "Enrichment Credits"
  event_name   = var.enrichment_meter_event_name

  default_aggregation {
    formula = "sum"
  }

  customer_mapping {
    event_payload_key = "stripe_customer_id"
    type              = "by_id"
  }

  value_settings {
    event_payload_key = "value"
  }
}

# ─────────────────────────────────────────────
# Starter Tier  ($29/mo, 1000 credits, $0.025 overage)
# ─────────────────────────────────────────────
resource "stripe_product" "starter_base" {
  name        = "AmpleData Starter"
  description = "Includes 1000 enrichment credits per month"

  metadata = {
    ampledata_tier         = "starter"
    ampledata_product_type = "base"
  }
}

resource "stripe_product" "starter_metered" {
  name        = "AmpleData Starter — Overage"
  description = "Usage-based billing for credits beyond your included allowance"

  metadata = {
    ampledata_tier         = "starter"
    ampledata_product_type = "metered"
  }
}

resource "stripe_price" "starter_base" {
  product     = stripe_product.starter_base.id
  currency    = "usd"
  unit_amount = 2900

  recurring {
    interval = "month"
  }

  metadata = {
    ampledata_price_type = "base"
    ampledata_tier       = "starter"
  }
}

resource "stripe_price" "starter_metered" {
  product             = stripe_product.starter_metered.id
  currency            = "usd"
  unit_amount_decimal = "2.5"

  recurring {
    interval   = "month"
    usage_type = "metered"
    meter      = stripe_billing_meter.enrichment_credits.id
  }

  metadata = {
    ampledata_price_type = "metered"
    ampledata_tier       = "starter"
  }
}

# ─────────────────────────────────────────────
# Pro Tier  ($99/mo, 5000 credits, $0.018 overage)
# ─────────────────────────────────────────────
resource "stripe_product" "pro_base" {
  name        = "AmpleData Pro"
  description = "Includes 5000 enrichment credits per month"

  metadata = {
    ampledata_tier         = "pro"
    ampledata_product_type = "base"
  }
}

resource "stripe_product" "pro_metered" {
  name        = "AmpleData Pro — Overage"
  description = "Usage-based billing for credits beyond your included allowance"

  metadata = {
    ampledata_tier         = "pro"
    ampledata_product_type = "metered"
  }
}

resource "stripe_price" "pro_base" {
  product     = stripe_product.pro_base.id
  currency    = "usd"
  unit_amount = 9900

  recurring {
    interval = "month"
  }

  metadata = {
    ampledata_price_type = "base"
    ampledata_tier       = "pro"
  }
}

resource "stripe_price" "pro_metered" {
  product             = stripe_product.pro_metered.id
  currency            = "usd"
  unit_amount_decimal = "1.8"

  recurring {
    interval   = "month"
    usage_type = "metered"
    meter      = stripe_billing_meter.enrichment_credits.id
  }

  metadata = {
    ampledata_price_type = "metered"
    ampledata_tier       = "pro"
  }
}

# ─────────────────────────────────────────────
# Enterprise Tier  ($299/mo, 25000 credits, $0.01 overage)
# ─────────────────────────────────────────────
resource "stripe_product" "enterprise_base" {
  name        = "AmpleData Enterprise"
  description = "Includes 25000 enrichment credits per month"

  metadata = {
    ampledata_tier         = "enterprise"
    ampledata_product_type = "base"
  }
}

resource "stripe_product" "enterprise_metered" {
  name        = "AmpleData Enterprise — Overage"
  description = "Usage-based billing for credits beyond your included allowance"

  metadata = {
    ampledata_tier         = "enterprise"
    ampledata_product_type = "metered"
  }
}

resource "stripe_price" "enterprise_base" {
  product     = stripe_product.enterprise_base.id
  currency    = "usd"
  unit_amount = 29900

  recurring {
    interval = "month"
  }

  metadata = {
    ampledata_price_type = "base"
    ampledata_tier       = "enterprise"
  }
}

resource "stripe_price" "enterprise_metered" {
  product             = stripe_product.enterprise_metered.id
  currency            = "usd"
  unit_amount_decimal = "1"

  recurring {
    interval   = "month"
    usage_type = "metered"
    meter      = stripe_billing_meter.enrichment_credits.id
  }

  metadata = {
    ampledata_price_type = "metered"
    ampledata_tier       = "enterprise"
  }
}
