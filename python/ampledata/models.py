from typing import Any


class EnrichmentRequest:
    keys: list[str]  # Columns that uniquely identify rows
    columns_to_enrich: list[str]  # Target columns for enrichment
    data: list[dict[str, Any]]  # The actual rows to enrich
    search_context: dict[str, Any] | None = None  # Optional metadata for queries
