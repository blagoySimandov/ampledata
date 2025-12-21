from typing import TypedDict


class SearchParameters(TypedDict):
    q: str
    type: str
    engine: str


class KnowledgeGraphAttributes(TypedDict, total=False):
    """CEO, COO, Founders, etc. - all optional as different entities have different attributes"""

    customer_service: str
    founders: str
    founded: str
    headquarters: str
    coo: str
    ceo: str


class KnowledgeGraph(TypedDict, total=False):
    """Knowledge graph info - some fields like attributes or type are optional"""

    title: str
    type: str
    imageUrl: str
    description: str
    descriptionSource: str
    descriptionLink: str
    attributes: KnowledgeGraphAttributes


class Sitelink(TypedDict):
    title: str
    link: str


class OrganicResult(TypedDict, total=False):
    """Organic search results - some fields like date and rating are optional"""

    title: str
    link: str
    snippet: str
    sitelinks: list[Sitelink]
    date: str
    position: int
    rating: float
    ratingMax: int
    ratingCount: int


class PeopleAlsoAskItem(TypedDict):
    question: str
    snippet: str
    title: str
    link: str


class RelatedSearch(TypedDict):
    query: str


class GoogleSearchResults(TypedDict):
    searchParameters: SearchParameters
    knowledgeGraph: KnowledgeGraph
    organic: list[OrganicResult]
    peopleAlsoAsk: list[PeopleAlsoAskItem]
    relatedSearches: list[RelatedSearch]
    credits: int
