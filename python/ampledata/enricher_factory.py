from enricher import Enricher, AsyncEnricher
from query_builder import IQueryBuilder, QueryBuilder
from crawl_decision_maker import ICrawlDecisionMaker, GroqDecisionMaker
from web_search import IWebSearcher, SerperWebSearcher
from web_crawler import IWebCrawler, WebCrawler
from content_extractor import IContentExtractor, GroqContentExtractor
from models import ColumnMetadata
from groq import Groq
from state_manager import IStateManager, StateManager, IStateStore


class EnricherFactory:
    @staticmethod
    def create(
        columns_metadata: list[ColumnMetadata],
        entity_type: str | None = None,
        query_builder: IQueryBuilder | None = None,
        crawl_decision_maker: ICrawlDecisionMaker | None = None,
        web_searcher: IWebSearcher | None = None,
        web_crawler: IWebCrawler | None = None,
        content_extractor: IContentExtractor | None = None,
        groq_client: Groq | None = None,
    ) -> Enricher:
        columns_to_search_for = [col.name for col in columns_metadata]

        if query_builder is None:
            query_builder = QueryBuilder(
                columns_to_search_for=columns_to_search_for,
                entity_type=entity_type,
            )

        if crawl_decision_maker is None:
            crawl_decision_maker = GroqDecisionMaker(
                columns_metadata=columns_metadata,
                groq_client=groq_client,
            )

        if web_searcher is None:
            web_searcher = SerperWebSearcher()

        if web_crawler is None:
            web_crawler = WebCrawler()

        if content_extractor is None:
            content_extractor = GroqContentExtractor(
                columns_metadata=columns_metadata,
                groq_client=groq_client,
            )

        return Enricher(
            query_builder=query_builder,
            crawl_decision_maker=crawl_decision_maker,
            web_searcher=web_searcher,
            web_crawler=web_crawler,
            content_extractor=content_extractor,
        )

    @staticmethod
    def create_with_defaults(
        columns_metadata: list[ColumnMetadata],
    ) -> Enricher:
        return EnricherFactory.create(
            columns_metadata=columns_metadata,
        )

    @staticmethod
    def create_async(
        columns_metadata: list[ColumnMetadata],
        state_manager: IStateManager | None = None,
        state_store: IStateStore | None = None,
        entity_type: str | None = None,
        query_builder: IQueryBuilder | None = None,
        crawl_decision_maker: ICrawlDecisionMaker | None = None,
        web_searcher: IWebSearcher | None = None,
        web_crawler: IWebCrawler | None = None,
        content_extractor: IContentExtractor | None = None,
        groq_client: Groq | None = None,
        concurrency: int = 5,
    ) -> AsyncEnricher:
        columns_to_search_for = [col.name for col in columns_metadata]

        if state_manager is None:
            if state_store is None:
                raise ValueError(
                    "Either state_manager or state_store must be provided for AsyncEnricher"
                )
            state_manager = StateManager(store=state_store)

        if query_builder is None:
            query_builder = QueryBuilder(
                columns_to_search_for=columns_to_search_for,
                entity_type=entity_type,
            )

        if crawl_decision_maker is None:
            crawl_decision_maker = GroqDecisionMaker(
                columns_metadata=columns_metadata,
                groq_client=groq_client,
            )

        if web_searcher is None:
            web_searcher = SerperWebSearcher()

        if web_crawler is None:
            web_crawler = WebCrawler()

        if content_extractor is None:
            content_extractor = GroqContentExtractor(
                columns_metadata=columns_metadata,
                groq_client=groq_client,
            )

        return AsyncEnricher(
            state_manager=state_manager,
            query_builder=query_builder,
            crawl_decision_maker=crawl_decision_maker,
            web_searcher=web_searcher,
            web_crawler=web_crawler,
            content_extractor=content_extractor,
            concurrency=concurrency,
        )

    @staticmethod
    def create_async_with_defaults(
        columns_metadata: list[ColumnMetadata],
        state_store: IStateStore,
        concurrency: int = 5,
    ) -> AsyncEnricher:
        return EnricherFactory.create_async(
            columns_metadata=columns_metadata,
            state_store=state_store,
            concurrency=concurrency,
        )
