package models

type SearchParameters struct {
	Q      string `json:"q"`
	Type   string `json:"type"`
	Engine string `json:"engine"`
}

type KnowledgeGraphAttributes struct {
	CustomerService *string `json:"customer_service,omitempty"`
	Founders        *string `json:"founders,omitempty"`
	Founded         *string `json:"founded,omitempty"`
	Headquarters    *string `json:"headquarters,omitempty"`
	COO             *string `json:"coo,omitempty"`
	CEO             *string `json:"ceo,omitempty"`
}

type KnowledgeGraph struct {
	Title             *string                    `json:"title,omitempty"`
	Type              *string                    `json:"type,omitempty"`
	ImageURL          *string                    `json:"imageUrl,omitempty"`
	Description       *string                    `json:"description,omitempty"`
	DescriptionSource *string                    `json:"descriptionSource,omitempty"`
	DescriptionLink   *string                    `json:"descriptionLink,omitempty"`
	Attributes        *KnowledgeGraphAttributes  `json:"attributes,omitempty"`
}

type Sitelink struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type OrganicResult struct {
	Title       *string     `json:"title,omitempty"`
	Link        *string     `json:"link,omitempty"`
	Snippet     *string     `json:"snippet,omitempty"`
	Sitelinks   []Sitelink  `json:"sitelinks,omitempty"`
	Date        *string     `json:"date,omitempty"`
	Position    *int        `json:"position,omitempty"`
	Rating      *float64    `json:"rating,omitempty"`
	RatingMax   *int        `json:"ratingMax,omitempty"`
	RatingCount *int        `json:"ratingCount,omitempty"`
}

type PeopleAlsoAskItem struct {
	Question string `json:"question"`
	Snippet  string `json:"snippet"`
	Title    string `json:"title"`
	Link     string `json:"link"`
}

type RelatedSearch struct {
	Query string `json:"query"`
}

type GoogleSearchResults struct {
	SearchParameters SearchParameters    `json:"searchParameters"`
	KnowledgeGraph   *KnowledgeGraph     `json:"knowledgeGraph,omitempty"`
	Organic          []OrganicResult     `json:"organic"`
	PeopleAlsoAsk    []PeopleAlsoAskItem `json:"peopleAlsoAsk,omitempty"`
	RelatedSearches  []RelatedSearch     `json:"relatedSearches,omitempty"`
	Credits          *int                `json:"credits,omitempty"`
}
