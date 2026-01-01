from contextlib import asynccontextmanager
from fastapi import FastAPI
from pydantic import BaseModel
from crawl4ai import AsyncWebCrawler, CrawlerRunConfig, CacheMode
from crawl4ai.content_filter_strategy import BM25ContentFilter
from crawl4ai.markdown_generation_strategy import DefaultMarkdownGenerator

crawler = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global crawler
    crawler = AsyncWebCrawler()
    await crawler.start()
    yield
    if crawler:
        await crawler.close()


app = FastAPI(lifespan=lifespan)


class CrawlRequest(BaseModel):
    urls: list[str]
    query: str


class CrawlResponse(BaseModel):
    content: str
    success: bool


@app.post("/crawl")
async def crawl(request: CrawlRequest) -> CrawlResponse:
    config = CrawlerRunConfig(
        cache_mode=CacheMode.ENABLED,
        markdown_generator=DefaultMarkdownGenerator(
            content_filter=BM25ContentFilter(user_query=request.query)
        ),
        word_count_threshold=10,
        excluded_tags=["form", "header", "footer", "nav"],
        magic=True,
        simulate_user=True,
    )
    if crawler is None:
        return CrawlResponse(content="", success=False)
    results = await crawler.arun_many(urls=request.urls, config=config)

    results_markdown = []
    for result in results:  # type: ignore
        if result.success and result.markdown:
            md_content = result.markdown.fit_markdown or result.markdown.raw_markdown
            results_markdown.append(md_content)

    content = "\n\n---\n\n".join(results_markdown)
    return CrawlResponse(content=content, success=True)


@app.get("/health")
async def health():
    return {"status": "ok"}
