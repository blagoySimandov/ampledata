from abc import ABC, abstractmethod
from typing import Any
from dataclasses import dataclass
import json
from groq import Groq
from models import ColumnMetadata
from utils import clean_json_md


@dataclass
class ContentExtractionResult:
    extracted_data: dict[str, Any]
    reasoning: str


class IContentExtractor(ABC):
    @abstractmethod
    def extract(
        self,
        markdown_content: str,
        columns_metadata: list[ColumnMetadata],
        entity: str,
    ) -> ContentExtractionResult:
        pass


class GroqContentExtractor(IContentExtractor):
    def __init__(
        self, columns_metadata: list[ColumnMetadata], groq_client: Groq | None = None
    ):
        self.columns_metadata = columns_metadata
        self.client = groq_client or Groq()
        self.model = "openai/gpt-oss-20b"

    def extract(
        self,
        markdown_content: str,
        columns_metadata: list[ColumnMetadata],
        entity: str,
    ) -> ContentExtractionResult:
        prompt = self._build_extraction_prompt(
            markdown_content, columns_metadata, entity
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

    def _build_extraction_prompt(
        self,
        markdown_content: str,
        columns_metadata: list[ColumnMetadata],
        entity: str,
    ) -> str:
        columns_info = []
        for col in columns_metadata:
            desc = f" ({col.description})" if col.description else ""
            columns_info.append(f"- {col.name} [type: {col.type.value}]{desc}")

        columns_text = "\n".join(columns_info)

        truncated_content = markdown_content[:8000]

        return f"""You are a data extraction specialist. Extract the following fields from the provided website content about {entity}.

## Fields to Extract (ONLY extract these fields)
{columns_text}

## Website Content
{truncated_content}

## Your Task

Extract ONLY the fields listed above from the website content. Do not extract any other fields.

IMPORTANT: Extract each value in the CORRECT DATA TYPE as specified in the column metadata:
- For number types: use numeric values without quotes (e.g., 1000)
- For string types: use quoted strings
- For boolean types: use true/false without quotes
- For date types: use ISO 8601 format (YYYY-MM-DD)

If a field cannot be found in the content, omit it from the response.

## Response Format (JSON only, no markdown)
{{
    "extracted_data": {{"field_name": value_with_correct_type}},
    "reasoning": "Explanation of what was extracted from the content and how you found each field"
}}"""

    def _parse_response(self, content: str) -> ContentExtractionResult:
        content = clean_json_md(content)
        try:
            data = json.loads(content)
            return ContentExtractionResult(
                extracted_data=data.get("extracted_data", {}),
                reasoning=data.get("reasoning", ""),
            )
        except json.JSONDecodeError:
            return ContentExtractionResult(
                extracted_data={},
                reasoning=f"Failed to parse LLM response: {content[:100]}",
            )
