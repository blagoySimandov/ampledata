from enricher import Enricher
import json
from web_search import SerperWebSearcher
from crawl_decision_maker import GroqDecisionMaker
from query_builder import QueryBuilder
from groq import Groq
import os


def main():
    columns_to_search_for = ["revenue", "employees"]
    builder = QueryBuilder(
        columns_to_search_for=columns_to_search_for,
        notes_for_columns=["annual revenue", "employee count"],
    )
    groq_client = Groq(
        api_key=os.environ.get("GROQ_GEMINI")
    )  # TODO: change to GROQ_API_KEY
    crawl_decision_maker = GroqDecisionMaker(
        columns_to_enrich=columns_to_search_for,
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
