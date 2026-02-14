package billing

type SubscriptionTier struct {
	ID                       string
	DisplayName              string
	MonthlyPriceCents        int64
	IncludedTokens           int64
	OveragePriceCentsDecimal string
	ProductID                string
	BasePriceID              string
	MeteredPriceID           string
}

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

var TierOrder = []string{"starter", "pro", "enterprise"}

func GetTier(id string) *SubscriptionTier {
	return Tiers[id]
}
