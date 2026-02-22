variable "stripe_api_key" {
  description = "Stripe secret API key (sk_test_... or sk_live_...)"
  type        = string
  sensitive   = true
}

variable "enrichment_meter_event_name" {
  description = "Event name for the enrichment credits billing meter. Must match ENRICHMENT_COST_METER_NAME env var in the Go app."
  type        = string
  default     = "enrichment_credits"
}
