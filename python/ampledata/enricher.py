from google_search_results_model import GoogleSearchResults
from web_search import WebSearcher, SerperWebSearcher
from query_builder import IQueryBuilder
from crawl_decision_maker import ICrawlDecisionMaker, CrawlDecision
from dataclasses import asdict


class Enricher:
    def __init__(
        self,
        query_builder: IQueryBuilder,
        crawl_decision_maker: ICrawlDecisionMaker,
        web_searcher: WebSearcher = SerperWebSearcher(),
    ):
        self.web_searcher = web_searcher
        self.query_builder = query_builder
        self.crawl_decision_maker = crawl_decision_maker
        pass

    def _search_query(self, query) -> GoogleSearchResults:
        return self.web_searcher.search(query)

    def _build_query(self, entity: str) -> str:
        return self.query_builder.build(entity)

    def _enrich_keys(self, row_keys: list[str]) -> list[CrawlDecision]:
        results = []
        for key in row_keys:
            query = self._build_query(key)
            serp: GoogleSearchResults = self._search_query(query)
            decision = self.crawl_decision_maker.make_decision(serp, key)
            results.append(asdict(decision))

        return results
