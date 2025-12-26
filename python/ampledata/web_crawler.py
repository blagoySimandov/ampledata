from crawl4ai import (
    AsyncWebCrawler,
    CrawlerRunConfig,
    CacheMode,
    types,
)
from crawl4ai.content_filter_strategy import BM25ContentFilter
from crawl4ai.markdown_generation_strategy import DefaultMarkdownGenerator
from abc import abstractmethod, ABC


class IWebCrawler(ABC):
    @abstractmethod
    async def async_crawl(self, urls: list[str], query: str) -> str:
        pass

    @abstractmethod
    async def close(self):
        pass


class WebCrawler(IWebCrawler):
    def __init__(self):
        self.crawler = AsyncWebCrawler()
        self._started = False

    def _get_run_config(self, query: str) -> CrawlerRunConfig:
        return CrawlerRunConfig(
            cache_mode=CacheMode.ENABLED,
            markdown_generator=DefaultMarkdownGenerator(
                content_filter=BM25ContentFilter(
                    user_query=query,
                )
            ),
            word_count_threshold=10,
            excluded_tags=["form", "header", "footer", "nav"],
            magic=True,
            simulate_user=True,
        )

    async def _ensure_started(self):
        if not self._started:
            await self.crawler.start()
            self._started = True

    async def async_crawl(self, urls: list[str], query: str) -> str:
        await self._ensure_started()

        results_markdown = []
        crawl_results: types.RunManyReturn[str] = await self.crawler.arun_many(
            urls=urls, config=self._get_run_config(query)
        )

        for result in crawl_results:
            if result.success and result.markdown:
                md_content = (
                    result.markdown.fit_markdown or result.markdown.raw_markdown
                )
                results_markdown.append(md_content)

        return "\n\n---\n\n".join(results_markdown)

    async def close(self):
        if self._started:
            await self.crawler.close()
            self._started = False
