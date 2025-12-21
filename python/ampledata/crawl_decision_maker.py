from dataclasses import dataclass
import json
from typing import Any
from google_search_results_model import GoogleSearchResults
from utils import clean_json_md
from groq import Groq
from abc import ABC, abstractmethod
from models import ColumnMetadata


@dataclass
class CrawlDecision:
    urls_to_crawl: list[str]
    extracted_data: dict[str, Any] | None
    reasoning: str


class ICrawlDecisionMaker(ABC):
    columns_metadata: list[ColumnMetadata]

    @abstractmethod
    def make_decision(
        self,
        serp_results: GoogleSearchResults,
        row_key: str,
        max_urls: int = 3,
    ) -> CrawlDecision:
        pass


class GroqDecisionMaker(ICrawlDecisionMaker):
    def __init__(
        self, columns_metadata: list[ColumnMetadata], groq_client: Groq | None = None
    ):
        self.columns_metadata = columns_metadata
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
        1. Can we extract the needed data directly from snippets? (return empty urls_to_crawl)
        2. If not, which URLs should we crawl? (return list of URLs)
        """

        prompt = self._build_prompt(
            serp_results, self.columns_metadata, row_key, max_urls
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
        return self._parse_response(content, serp_results)

    def _build_prompt(
        self,
        serp_results: GoogleSearchResults,
        columns_metadata: list[ColumnMetadata],
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

        columns_info = []
        for col in columns_metadata:
            desc = f" ({col.description})" if col.description else ""
            columns_info.append(f"- {col.name} [type: {col.type.value}]{desc}")

        columns_text = "\n".join(columns_info)

        return f"""You are a data extraction assistant. Analyze these search results for "{entity}" and decide how to proceed.

## Columns We Need to Extract
{columns_text}

## Answer Box
{answer_box.get("snippet", "None")}

## Search Results
{organic_text}

## People Also Ask
{chr(10).join([f"Q: {p.get('question', '')} A: {p.get('snippet', '')}" for p in people_also_ask])}

## Your Task

1. First, check if the answer box and snippets already contain enough information to fill ALL the columns we need.
2. If YES: Extract the data directly and return empty urls_to_crawl array
   - IMPORTANT: Extract each value in the CORRECT DATA TYPE as specified in the column metadata
   - For number types: use numeric values without quotes (e.g., 1000)
   - For string types: use quoted strings
   - For boolean types: use true/false without quotes
   - For date types: use ISO 8601 format (YYYY-MM-DD)

3. If NO: Select up to {max_urls} URLs to crawl, prioritizing:
   - Wikipedia
   - Reliable data sources (SEC filings, financial sites)
   - Avoid SEO aggregator sites when primary sources are available

## Response Format (JSON only, no markdown)
{{
    "urls_to_crawl": ["url1", "url2"] or [],
    "extracted_data": {{"column_name": value_with_correct_type}} or null,
    "reasoning": "Explanation of decision"
}}"""

    def _parse_response(
        self, content: str, serp_results: GoogleSearchResults
    ) -> CrawlDecision:
        content = clean_json_md(content)
        try:
            data = json.loads(content)
            return CrawlDecision(
                urls_to_crawl=data.get("urls_to_crawl", []),
                extracted_data=data.get("extracted_data"),
                reasoning=data.get("reasoning", ""),
            )
        except json.JSONDecodeError:
            organic = serp_results.get("organic", [])[:3]
            fallback_urls = [r.get("link") for r in organic if r.get("link")]

            return CrawlDecision(
                urls_to_crawl=fallback_urls,
                extracted_data=None,
                reasoning=f"Failed to parse LLM response: {content[:100]}. Falling back to top URLs.",
            )
