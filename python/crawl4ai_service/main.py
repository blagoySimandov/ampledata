from contextlib import asynccontextmanager
from fastapi import FastAPI
from pydantic import BaseModel
from crawler import AmpleCrawler

crawler: AmpleCrawler | None = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global crawler
    crawler = AmpleCrawler()
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
    if crawler is None:
        return CrawlResponse(content="", success=False)
    try:
        parts = await crawler.crawl(urls=request.urls, query=request.query)
    except Exception as e:
        return CrawlResponse(content=str(e), success=False)

    return CrawlResponse(content="\n\n---\n\n".join(parts), success=True)


@app.get("/health")
async def health():
    return {"status": "ok"}
