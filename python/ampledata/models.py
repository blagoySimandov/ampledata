from typing import Any
from dataclasses import dataclass
from enum import Enum


class ColumnType(str, Enum):
    STRING = "string"
    NUMBER = "number"
    BOOLEAN = "boolean"
    DATE = "date"


@dataclass
class ColumnMetadata:
    name: str
    type: ColumnType
    description: str | None = None


class EnrichmentRequest:
    keys: list[str]
    columns_to_enrich: list[str]
    data: list[dict[str, Any]]
    search_context: dict[str, Any] | None = None
