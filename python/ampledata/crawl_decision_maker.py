from dataclasses import dataclass
import json
from google_search_results_model import GoogleSearchResults
from utils import clean_json_md
from groq import Groq
from abc import ABC, abstractmethod


@dataclass
class CrawlDecision:
    should_crawl: bool
    urls_to_crawl: list[str]
    extracted_data: dict[str, str] | None  # If we can skip crawling
    reasoning: str


class ICrawlDecisionMaker(ABC):
    columns_to_search_for: list[str]

    @abstractmethod
    def make_decision(
        self,
        serp_results: GoogleSearchResults,
        row_key: str,
        max_urls: int = 3,
    ) -> CrawlDecision:
        pass


class GroqDecisionMaker(ICrawlDecisionMaker):
    def __init__(self, columns_to_enrich: list[str], groq_client: Groq | None = None):
        self.columns_to_search_for = columns_to_enrich
        self.client = groq_client or Groq()
        self.model = "openai/gpt-oss-120b"

    def make_decision(
        self,
        serp_results: GoogleSearchResults,
        row_key: str,
        max_urls: int = 3,
    ) -> CrawlDecision:
        """
        Analyze SERP results and decide:
        1. Can we extract the needed data directly from snippets? (skip crawl)
        2. If not, which URLs should we crawl?
        """

        prompt = self._build_prompt(
            serp_results, self.columns_to_search_for, row_key, max_urls
        )

        response = self.client.chat.completions.create(
            model=self.model,
            messages=[{"role": "user", "content": prompt}],
            temperature=0,
            max_completion_tokens=2048,
            reasoning_effort="medium",
        )
        content = response.choices[0].message.content
        if content is None:
            raise Exception("No response from LLM")
        return self._parse_response(content)

    def _build_prompt(
        self,
        serp_results: GoogleSearchResults,
        columns_to_enrich: list[str],
        entity: str,
        max_urls: int,
    ) -> str:
        answer_box = serp_results.get("answerBox", {})
        organic = serp_results.get("organic", [])[:10]
        people_also_ask = serp_results.get("peopleAlsoAsk", [])[:3]

        organic_text = ""
        for r in organic:
            organic_text += f"""
Position {r.get("position", "?")}: {r.get("title", "")}
URL: {r.get("link", "")}
Snippet: {r.get("snippet", "")}
---"""

        return f"""You are a data extraction assistant. Analyze these search results for "{entity}" and decide how to proceed.

## Columns We Need to Extract
{", ".join(columns_to_enrich)}

## Answer Box
{answer_box.get("snippet", "None")}

## Search Results
{organic_text}

## People Also Ask
{chr(10).join([f"Q: {p.get('question', '')} A: {p.get('snippet', '')}" for p in people_also_ask])}

## Your Task

1. First, check if the answer box and snippets already contain enough information to fill ALL the columns we need.
2. If YES: Extract the data directly and set should_crawl=false
3. If NO: Select up to {max_urls} URLs to crawl, prioritizing:
   - Wikipedia
   - Reliable data sources (SEC filings, financial sites)
   - Avoid SEO aggregator sites when primary sources are available

## Response Format (JSON only, no markdown)
{{
    "should_crawl": true/false,
    "urls_to_crawl": ["url1", "url2"],
    "extracted_data": {{"column_name": "value"}} or null,
    "reasoning": "Why did we decide to crawl these URLs?"
}}"""

    def _parse_response(self, content: str) -> CrawlDecision:
        content = clean_json_md(content)
        try:
            data = json.loads(content)
            return CrawlDecision(
                should_crawl=data.get("should_crawl", True),
                urls_to_crawl=data.get("urls_to_crawl", []),
                extracted_data=data.get("extracted_data"),
                reasoning=data.get("reasoning", ""),
            )
        except json.JSONDecodeError:
            # Fallback: crawl top 3
            return CrawlDecision(
                should_crawl=True,
                urls_to_crawl=[],
                extracted_data=None,
                reasoning=f"Failed to parse LLM response: {content[:100]}",
            )
