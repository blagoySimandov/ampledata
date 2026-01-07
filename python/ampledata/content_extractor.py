from abc import ABC, abstractmethod
from typing import Any
from dataclasses import dataclass
import json
from groq import Groq
from models import ColumnMetadata
from utils import clean_json_md


@dataclass
class FieldConfidence:
    score: float  # 0.0 to 1.0
    reason: str   # Brief explanation


@dataclass
class ContentExtractionResult:
    extracted_data: dict[str, Any]
    confidence: dict[str, FieldConfidence]
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

CRITICAL: NEVER infer, estimate, or make up data. Only extract information that is explicitly stated in the content.
- If you see "10000+" do NOT convert it to "10001" - use the exact value or omit if uncertain
- If information is approximate, partial, or unclear, reflect that uncertainty in the confidence score

IMPORTANT: Extract each value in the CORRECT DATA TYPE as specified in the column metadata:
- For number types: use numeric values without quotes (e.g., 1000)
- For string types: use quoted strings
- For boolean types: use true/false without quotes
- For date types: use ISO 8601 format (YYYY-MM-DD)

If a field cannot be found in the content, omit it from the response.

## Response Format (JSON only, no markdown)
{{
    "extracted_data": {{"field_name": value_with_correct_type}},
    "confidence": {{
        "field_name": {{
            "score": 0.95,
            "reason": "Brief 1-sentence explanation"
        }}
    }},
    "reasoning": "Overall extraction summary"
}}

## Confidence Scoring Guidelines
- 1.0: Exact match, explicitly stated
- 0.8-0.9: Clear statement, minor interpretation needed
- 0.6-0.7: Partial information or context-based inference
- 0.4-0.5: Significant uncertainty or approximation
- <0.4: High uncertainty, possibly derived or estimated"""

    def _parse_response(self, content: str) -> ContentExtractionResult:
        content = clean_json_md(content)
        try:
            data = json.loads(content)

            # Parse confidence data
            confidence_dict = {}
            raw_confidence = data.get("confidence", {})
            for key, val in raw_confidence.items():
                if isinstance(val, dict):
                    confidence_dict[key] = FieldConfidence(
                        score=val.get("score", 0.0),
                        reason=val.get("reason", "")
                    )

            return ContentExtractionResult(
                extracted_data=data.get("extracted_data", {}),
                confidence=confidence_dict,
                reasoning=data.get("reasoning", ""),
            )
        except json.JSONDecodeError:
            return ContentExtractionResult(
                extracted_data={},
                confidence={},
                reasoning=f"Failed to parse LLM response: {content[:100]}",
            )
