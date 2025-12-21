from enricher import Enricher
import json
from web_search import SerperWebSearcher
from crawl_decision_maker import GroqDecisionMaker
from query_builder import QueryBuilder
from groq import Groq
import os
from models import ColumnMetadata, ColumnType


def main():
    columns_metadata = [
        ColumnMetadata(
            name="revenue", type=ColumnType.NUMBER, description="annual revenue in USD"
        ),
        ColumnMetadata(
            name="employees", type=ColumnType.NUMBER, description="employee count"
        ),
    ]

    columns_to_search_for = [col.name for col in columns_metadata]
    # notes_for_columns = [col.description or col.name for col in columns_metadata]

    builder = QueryBuilder(
        columns_to_search_for=columns_to_search_for,
    )
    groq_client = Groq(api_key=os.environ.get("GROQ_GEMINI"))
    crawl_decision_maker = GroqDecisionMaker(
        columns_metadata=columns_metadata,
        groq_client=groq_client,
    )
    web_searcher = SerperWebSearcher()
    enricher = Enricher(
        builder, crawl_decision_maker=crawl_decision_maker, web_searcher=web_searcher
    )

    print(json.dumps(enricher._enrich_keys(["apple", "stripe"]), indent=2))
    pass


if __name__ == "__main__":
    main()
