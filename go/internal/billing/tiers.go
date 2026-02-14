package billing

// SubscriptionTier defines a subscription plan tier.
type SubscriptionTier struct {
	ID                       string
	DisplayName              string
	MonthlyPriceCents        int64
	IncludedTokens           int64
	OveragePriceCentsDecimal string // sub-cent precision, e.g. "3.5"
	ProductID                string // set by SyncStripeCatalog
	BasePriceID              string // set by SyncStripeCatalog
	MeteredPriceID           string // set by SyncStripeCatalog
}

// Tiers holds all available subscription tiers keyed by tier ID.
var Tiers = map[string]*SubscriptionTier{
	"starter": {
		ID:                       "starter",
		DisplayName:              "Starter",
		MonthlyPriceCents:        2900,
		IncludedTokens:           1000,
		OveragePriceCentsDecimal: "3.5",
	},
	"pro": {
		ID:                       "pro",
		DisplayName:              "Pro",
		MonthlyPriceCents:        9900,
		IncludedTokens:           5000,
		OveragePriceCentsDecimal: "2",
	},
	"enterprise": {
		ID:                       "enterprise",
		DisplayName:              "Enterprise",
		MonthlyPriceCents:        29900,
		IncludedTokens:           25000,
		OveragePriceCentsDecimal: "1",
	},
}

// TierOrder defines the display ordering of tiers.
var TierOrder = []string{"starter", "pro", "enterprise"}

// GetTier returns a tier by its ID.
func GetTier(id string) *SubscriptionTier {
	return Tiers[id]
}

// GetTierByProductID finds a tier by its Stripe product ID.
func GetTierByProductID(productID string) *SubscriptionTier {
	for _, t := range Tiers {
		if t.ProductID == productID {
			return t
		}
	}
	return nil
}
