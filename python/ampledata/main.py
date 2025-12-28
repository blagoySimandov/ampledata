from enricher_factory import EnricherFactory
from postgres_state_store import PostgresStateStore
import json
import asyncio
import os
from models import ColumnMetadata, ColumnType


async def main():
    columns_metadata = [
        ColumnMetadata(
            name="founded_year",
            type=ColumnType.NUMBER,
            description="year the company was founded",
        ),
        ColumnMetadata(
            name="headquarters_city",
            type=ColumnType.STRING,
            description="city where company headquarters is located",
        ),
        ColumnMetadata(
            name="ceo_name",
            type=ColumnType.STRING,
            description="current CEO full name",
        ),
        ColumnMetadata(
            name="num_products",
            type=ColumnType.NUMBER,
            description="number of major products or product lines",
        ),
    ]

    db_url = os.getenv(
        "DATABASE_URL_ENRICHMENT",
        "postgresql+asyncpg://enrichment:enrichment@localhost:5432/enrichment",
    )

    state_store = PostgresStateStore(connection_string=db_url, pool_size=10, echo=False)

    try:
        await state_store.initialize_database()

        enricher = EnricherFactory.create_async_with_defaults(
            columns_metadata=columns_metadata,
            state_store=state_store,
            concurrency=5,
        )

        try:
            results = await enricher.enrich_keys(
                ["databricks", "snowflake", "confluent"]
            )

            with open("results.json", "w") as f:
                f.write(json.dumps(results, indent=2))
        finally:
            await enricher.close()
    finally:
        await state_store.close()


if __name__ == "__main__":
    asyncio.run(main())
