from crawl4ai import AsyncWebCrawler, CrawlerRunConfig, CacheMode
from crawl4ai.content_filter_strategy import BM25ContentFilter
from crawl4ai.markdown_generation_strategy import DefaultMarkdownGenerator


class AmpleCrawler:
    def __init__(self, cache_mode: CacheMode = CacheMode.ENABLED):
        self._crawler = AsyncWebCrawler()
        self._cache_mode = cache_mode

    async def start(self):
        await self._crawler.start()

    async def close(self):
        await self._crawler.close()

    async def __aenter__(self):
        await self.start()
        return self

    async def __aexit__(self, *args):
        await self.close()

    # query is currently not used. it can be used for smart filtering of content in the future
    # but unfortunately it currently doesn not work well.
    def _make_config(self, _query: str) -> CrawlerRunConfig:
        return CrawlerRunConfig(
            cache_mode=self._cache_mode,
            markdown_generator=DefaultMarkdownGenerator(),
            excluded_tags=["form", "header", "footer", "nav", "script"],
            magic=True,
            simulate_user=True,
        )

    async def crawl(self, urls: list[str], query: str) -> list[str]:
        results = await self._crawler.arun_many(
            urls=urls, config=self._make_config(query)
        )
        return [
            result.markdown.fit_markdown or result.markdown.raw_markdown
            for result in results  # type: ignore
            if result.success and result.markdown
        ]
