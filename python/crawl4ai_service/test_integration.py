import pytest
import pytest_asyncio
from crawl4ai import CacheMode
from crawler import AmpleCrawler

CRAWL_TARGETS = [
    {
        "name": "Wikipedia",
        "urls": ["https://en.wikipedia.org/wiki/Go_(programming_language)"],
        "query": "Go programming language",
    },
    {
        "name": "HackerNews",
        "urls": ["https://news.ycombinator.com/"],
        "query": "tech news",
    },
    {
        "name": "GitHubCrawl4ai",
        "urls": ["https://github.com/unclecode/crawl4ai"],
        "query": "crawl4ai web crawling library",
    },
]

CLOUDFLARE_SIGNALS = [
    "just a moment",
    "checking your browser",
    "enable javascript and cookies",
    "cf-browser-verification",
    "cf_clearance",
    "attention required",
    "ray id",
]


def assert_not_cloudflare_blocked(content: str, name: str) -> None:
    lower = content.lower()
    for signal in CLOUDFLARE_SIGNALS:
        assert signal not in lower, (
            f"[{name}] Response appears Cloudflare-blocked (found '{signal}')"
        )


@pytest_asyncio.fixture(scope="module")
async def crawler():
    async with AmpleCrawler(cache_mode=CacheMode.BYPASS) as c:
        yield c


@pytest.mark.asyncio(loop_scope="module")
@pytest.mark.parametrize(
    "target", CRAWL_TARGETS, ids=[t["name"] for t in CRAWL_TARGETS]
)
async def test_crawl_returns_content(crawler, target, save_crawl_output):
    parts = await crawler.crawl(urls=target["urls"], query=target["query"])

    assert parts, f"[{target['name']}] No results returned"

    for md in parts:
        assert md and md.strip(), f"[{target['name']}] Empty markdown"
        print("MD: ", md)
        save_crawl_output[target["urls"][0]] = md

    combined = "\n\n---\n\n".join(parts)
    assert_not_cloudflare_blocked(combined, target["name"])
