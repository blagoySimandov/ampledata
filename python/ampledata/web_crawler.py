from crawl4ai import AsyncWebCrawler
from abc import abstractmethod, ABC


class IWebCrawler(ABC):
    @abstractmethod
    async def async_crawl(self, urls: list[str]) -> str:
        pass

    @abstractmethod
    async def close(self):
        pass


class WebCrawler(IWebCrawler):
    def __init__(self):
        self.crawler: AsyncWebCrawler = AsyncWebCrawler()
        self._started = False

    async def _ensure_started(self):
        if not self._started:
            await self.crawler.start()
            self._started = True

    async def async_crawl(self, urls: list[str]) -> str:
        await self._ensure_started()

        results_markdown = []
        crawl_results = await self.crawler.arun_many(urls=urls)

        for result in crawl_results:  # type: ignore
            if result.success and result.markdown:
                results_markdown.append(result.markdown)

        return "\n\n".join(results_markdown)

    async def close(self):
        if self._started:
            await self.crawler.close()
            self._started = False
