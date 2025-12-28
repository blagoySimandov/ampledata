from google_search_results_model import GoogleSearchResults
from web_search import IWebSearcher, SerperWebSearcher
from query_builder import IQueryBuilder
from crawl_decision_maker import ICrawlDecisionMaker
from web_crawler import IWebCrawler
from content_extractor import IContentExtractor, ContentExtractionResult
from abc import ABC, abstractmethod


class IEnricher(ABC):
    @abstractmethod
    async def enrich_keys(self, row_keys: list[str]) -> list[dict]:
        pass


class Enricher(IEnricher):
    def __init__(
        self,
        query_builder: IQueryBuilder,
        crawl_decision_maker: ICrawlDecisionMaker,
        web_searcher: IWebSearcher = SerperWebSearcher(),
        web_crawler: IWebCrawler | None = None,
        content_extractor: IContentExtractor | None = None,
    ):
        self.web_searcher = web_searcher
        self.query_builder = query_builder
        self.crawl_decision_maker = crawl_decision_maker
        self.web_crawler = web_crawler
        self.content_extractor = content_extractor

    def _search_query(self, query) -> GoogleSearchResults:
        return self.web_searcher.search(query)

    def _build_query(self, entity: str) -> str:
        return self.query_builder.build(entity)

    async def enrich_keys(self, row_keys: list[str]) -> list[dict]:
        results = []
        for key in row_keys:
            query = self._build_query(key)
            serp: GoogleSearchResults = self._search_query(query)
            decision = self.crawl_decision_maker.make_decision(serp, key)

            missing_columns = decision.get_missing_columns(
                self.crawl_decision_maker.columns_metadata
            )

            sources = []
            content_extraction_reasoning = None
            content_extracted_data = None

            if (
                decision.urls_to_crawl
                and missing_columns
                and self.web_crawler
                and self.content_extractor
            ):
                missing_columns_metadata = [
                    col
                    for col in self.crawl_decision_maker.columns_metadata
                    if col.name in missing_columns
                ]
                markdown_content = await self.web_crawler.async_crawl(
                    decision.urls_to_crawl, query=query
                )

                if markdown_content:
                    sources = decision.urls_to_crawl
                    extraction_result: ContentExtractionResult = (
                        self.content_extractor.extract(
                            markdown_content,
                            missing_columns_metadata,
                            key,
                        )
                    )
                    content_extracted_data = extraction_result.extracted_data
                    content_extraction_reasoning = extraction_result.reasoning

                    if decision.extracted_data:
                        decision.extracted_data.update(content_extracted_data)
                    else:
                        decision.extracted_data = content_extracted_data

            result = {}
            result["extracted_data"] = decision.extracted_data
            result["reasoning"] = decision.reasoning
            result["sources"] = sources
            result["content_extracted_data"] = content_extracted_data
            result["content_extraction_reasoning"] = content_extraction_reasoning
            results.append(result)

        return results

    async def close(self):
        if self.web_crawler:
            await self.web_crawler.close()


class AsyncEnricher(IEnricher):
    def __init__(self):
        pass
