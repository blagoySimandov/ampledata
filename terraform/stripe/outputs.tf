output "enrichment_meter_id" {
  description = "ID of the enrichment credits billing meter"
  value       = stripe_meter.enrichment_credits.id
}

# Starter
output "starter_base_price_id" {
  description = "Price ID for the Starter tier flat monthly fee — set as STRIPE_STARTER_BASE_PRICE_ID"
  value       = stripe_price.starter_base.id
}

output "starter_metered_price_id" {
  description = "Price ID for the Starter tier metered overage — set as STRIPE_STARTER_METERED_PRICE_ID"
  value       = stripe_price.starter_metered.id
}

# Pro
output "pro_base_price_id" {
  description = "Price ID for the Pro tier flat monthly fee — set as STRIPE_PRO_BASE_PRICE_ID"
  value       = stripe_price.pro_base.id
}

output "pro_metered_price_id" {
  description = "Price ID for the Pro tier metered overage — set as STRIPE_PRO_METERED_PRICE_ID"
  value       = stripe_price.pro_metered.id
}

# Enterprise
output "enterprise_base_price_id" {
  description = "Price ID for the Enterprise tier flat monthly fee — set as STRIPE_ENTERPRISE_BASE_PRICE_ID"
  value       = stripe_price.enterprise_base.id
}

output "enterprise_metered_price_id" {
  description = "Price ID for the Enterprise tier metered overage — set as STRIPE_ENTERPRISE_METERED_PRICE_ID"
  value       = stripe_price.enterprise_metered.id
}

# Convenience: print the env-var block you can paste into your deployment config
output "env_var_block" {
  description = "Copy-paste this block into your deployment environment"
  value       = <<-EOT
    STRIPE_STARTER_BASE_PRICE_ID=${stripe_price.starter_base.id}
    STRIPE_STARTER_METERED_PRICE_ID=${stripe_price.starter_metered.id}
    STRIPE_PRO_BASE_PRICE_ID=${stripe_price.pro_base.id}
    STRIPE_PRO_METERED_PRICE_ID=${stripe_price.pro_metered.id}
    STRIPE_ENTERPRISE_BASE_PRICE_ID=${stripe_price.enterprise_base.id}
    STRIPE_ENTERPRISE_METERED_PRICE_ID=${stripe_price.enterprise_metered.id}
    ENRICHMENT_COST_METER_NAME=${var.enrichment_meter_event_name}
  EOT
}
