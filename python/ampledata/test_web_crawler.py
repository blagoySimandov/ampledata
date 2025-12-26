import asyncio
from web_crawler import WebCrawler
from os import write


async def test_wikipedia_crawler():
    crawler = WebCrawler()
    try:
        urls = [
            "https://en.wikipedia.org/wiki/Python_(programming_language)",
            "https://en.wikipedia.org/wiki/Artificial_intelligence",
            "https://en.wikipedia.org/wiki/Web_scraping",
        ]

        result = await crawler.async_crawl(urls, "python ai_capabilities")

        print(f"URLs crawled: {len(urls)}")
        print(f"Content length: {len(result)} characters")
        print(f"\nFirst 300 characters:\n{result[:300]}\n...")
        with open("crawl_data.md", "w") as f:
            f.write(result)

        assert isinstance(result, str) and len(result) > 0
    finally:
        await crawler.close()


if __name__ == "__main__":
    asyncio.run(test_wikipedia_crawler())
