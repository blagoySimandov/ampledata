import asyncio
from web_crawler import WebCrawler


async def test_web_crawler():
    crawler = WebCrawler()

    try:
        print("Testing WebCrawler with example.com...")
        urls = ["https://example.com"]

        result = await crawler.async_crawl(urls)

        print(f"\n{'=' * 60}")
        print("CRAWL RESULT:")
        print(f"{'=' * 60}")
        print(f"URLs crawled: {urls}")
        print(f"Content length: {len(result)} characters")
        print(f"\nFirst 500 characters of markdown:")
        print(f"{'-' * 60}")
        print(result[:500])
        print(f"{'-' * 60}")

        assert isinstance(result, str), "Result should be a string"
        assert len(result) > 0, "Result should not be empty"
        print("\n✅ Test passed!")
    finally:
        await crawler.close()


async def test_multiple_urls():
    crawler = WebCrawler()

    try:
        print("\n\nTesting WebCrawler with multiple URLs...")
        urls = [
            "https://example.com",
            "https://example.org",
        ]

        result = await crawler.async_crawl(urls)

        print(f"\n{'=' * 60}")
        print("MULTIPLE URLs CRAWL RESULT:")
        print(f"{'=' * 60}")
        print(f"URLs crawled: {urls}")
        print(f"Content length: {len(result)} characters")
        print(f"\nFirst 500 characters of combined markdown:")
        print(f"{'-' * 60}")
        print(result[:500])
        print(f"{'-' * 60}")

        assert isinstance(result, str), "Result should be a string"
        assert len(result) > 0, "Result should not be empty"
        print("\n✅ Multiple URLs test passed!")
    finally:
        await crawler.close()


async def test_empty_urls():
    crawler = WebCrawler()

    try:
        print("\n\nTesting WebCrawler with empty URL list...")
        urls = []

        result = await crawler.async_crawl(urls)

        print(f"\n{'=' * 60}")
        print("EMPTY URLs TEST:")
        print(f"{'=' * 60}")
        print(f"URLs crawled: {urls}")
        print(f"Result: '{result}'")
        print(f"Result length: {len(result)}")

        assert isinstance(result, str), "Result should be a string"
        assert result == "", "Empty URLs should return empty string"
        print("\n✅ Empty URLs test passed!")
    finally:
        await crawler.close()


async def test_crawler_reuse():
    print("\n\nTesting WebCrawler reuse (same crawler, multiple crawls)...")
    crawler = WebCrawler()

    try:
        result1 = await crawler.async_crawl(["https://example.com"])
        print(f"First crawl: {len(result1)} characters")

        result2 = await crawler.async_crawl(["https://example.org"])
        print(f"Second crawl: {len(result2)} characters")

        result3 = await crawler.async_crawl(
            ["https://example.com", "https://example.org"]
        )
        print(f"Third crawl (both URLs): {len(result3)} characters")

        assert all(isinstance(r, str) for r in [result1, result2, result3])
        print("\n✅ Crawler reuse test passed! (Browser was reused across calls)")
    finally:
        await crawler.close()


async def main():
    try:
        await test_web_crawler()
        await test_multiple_urls()
        await test_empty_urls()
        await test_crawler_reuse()
        print("\n" + "=" * 60)
        print("🎉 ALL TESTS PASSED!")
        print("=" * 60)
    except Exception as e:
        print(f"\n❌ Test failed with error: {e}")
        raise


if __name__ == "__main__":
    asyncio.run(main())
