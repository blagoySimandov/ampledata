from google_search_results_model import GoogleSearchResults
from web_search import IWebSearcher, SerperWebSearcher
from query_builder import IQueryBuilder
from crawl_decision_maker import ICrawlDecisionMaker
from web_crawler import IWebCrawler
from content_extractor import IContentExtractor, ContentExtractionResult
from abc import ABC, abstractmethod
from state_manager import IStateManager, RowState, RowStage, JobStoppedException, JobStatus
from typing import Awaitable, Callable
import asyncio


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
    def __init__(
        self,
        state_manager: IStateManager,
        query_builder: IQueryBuilder,
        crawl_decision_maker: ICrawlDecisionMaker,
        web_searcher: IWebSearcher,
        web_crawler: IWebCrawler | None = None,
        content_extractor: IContentExtractor | None = None,
        concurrency: int = 5,
    ):
        self.state_manager = state_manager
        self.query_builder = query_builder
        self.crawl_decision_maker = crawl_decision_maker
        self.web_searcher = web_searcher
        self.web_crawler = web_crawler
        self.content_extractor = content_extractor
        self.concurrency = concurrency

    async def enrich_keys(
        self, row_keys: list[str] | None = None, job_id: str | None = None
    ) -> list[dict]:
        if job_id is None:
            if row_keys is None:
                raise ValueError("Either job_id or row_keys must be provided")
            job_id = self.state_manager.generate_job_id()
            await self.state_manager.initialize_job(job_id, row_keys)
        else:
            if row_keys is not None:
                raise ValueError(
                    "Cannot provide both job_id and row_keys; job_id implies resume"
                )

            await self.state_manager.resume(job_id)

        try:
            # Stage 1: Fetch SERPs
            await self._run_stage(job_id, RowStage.SERP_FETCHED, self._fetch_serp)

            # Stage 2: Make decisions
            await self._run_stage(job_id, RowStage.DECISION_MADE, self._make_decision)

            # Stage 3: Crawl
            await self._run_stage(job_id, RowStage.CRAWLED, self._crawl)

            # Stage 4: Extract/enrich
            await self._run_stage(job_id, RowStage.ENRICHED, self._extract_content)

            # Mark complete
            await self._run_stage(job_id, RowStage.COMPLETED, self._finalize)

            await self.state_manager.complete(job_id)

        except JobStoppedException:
            pass  # Graceful exit on stop (pause/cancel)

        # Collect results
        return await self._collect_results(job_id)

    async def _run_stage(
        self,
        job_id: str,
        target_stage: RowStage,
        handler: Callable[[str, RowState], Awaitable[dict]],
    ) -> None:
        pending = await self.state_manager.get_pending_for_stage(job_id, target_stage)
        semaphore = asyncio.Semaphore(self.concurrency)

        async def process_one(state: RowState):
            async with semaphore:
                try:
                    data_update = await handler(job_id, state)
                    await self.state_manager.transition(
                        job_id, state.key, target_stage, data_update
                    )
                except JobStoppedException:
                    raise
                except Exception as e:
                    await self.state_manager.transition(
                        job_id, state.key, RowStage.FAILED, {"error": str(e)}
                    )

        await asyncio.gather(*[process_one(s) for s in pending])

    async def _fetch_serp(self, job_id: str, state: RowState) -> dict:
        query = self.query_builder.build(state.key)
        serp = await asyncio.to_thread(self.web_searcher.search, query)
        return {"serp_data": {"query": query, "results": serp}}

    async def _make_decision(self, job_id: str, state: RowState) -> dict:
        if state.serp_data is None:
            raise ValueError("SERP data is missing")
        serp = GoogleSearchResults(state.serp_data["results"])
        decision = self.crawl_decision_maker.make_decision(serp, state.key)
        return {
            "decision": {
                "urls_to_crawl": decision.urls_to_crawl,
                "extracted_data": decision.extracted_data,
                "reasoning": decision.reasoning,
                "missing_columns": decision.get_missing_columns(
                    self.crawl_decision_maker.columns_metadata
                ),
            }
        }

    async def _crawl(self, job_id: str, state: RowState) -> dict:
        decision = state.decision
        if not decision or not self.web_crawler or not state.serp_data:
            raise ValueError("Missing required data")

        if not decision["urls_to_crawl"] or not self.web_crawler:
            return {"crawl_results": {"content": None, "sources": []}}

        content = await self.web_crawler.async_crawl(
            decision["urls_to_crawl"], query=state.serp_data["query"]
        )
        return {
            "crawl_results": {"content": content, "sources": decision["urls_to_crawl"]}
        }

    async def _extract_content(self, job_id: str, state: RowState) -> dict:
        crawl = state.crawl_results
        decision = state.decision

        if not decision or not crawl or not self.content_extractor:
            raise ValueError("Missing decision, crawl results, or content extractor")

        if not crawl["content"] or not self.content_extractor:
            return {"extracted_data": decision["extracted_data"]}

        missing_cols_meta = [
            col
            for col in self.crawl_decision_maker.columns_metadata
            if col.name in decision["missing_columns"]
        ]

        result = self.content_extractor.extract(
            crawl["content"],
            missing_cols_meta,
            state.key,
        )

        merged = {**(decision["extracted_data"] or {}), **(result.extracted_data or {})}
        return {"extracted_data": merged}

    async def _finalize(self, job_id: str, state: RowState) -> dict:
        return {}  # Just mark as complete

    async def _collect_results(
        self, job_id: str, start: int = 0, limit: int = 100
    ) -> list[dict]:
        progress = await self.state_manager.get_progress(job_id)
        # You'd fetch all rows and format them here
        # This is a simplified version
        return []

    async def close(self):
        if self.web_crawler:
            await self.web_crawler.close()
