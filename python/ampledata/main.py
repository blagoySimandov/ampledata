from enricher_factory import EnricherFactory
import json
import asyncio
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

    enricher = EnricherFactory.create_with_defaults(
        columns_metadata=columns_metadata,
    )

    try:
        with open("results.json", "w") as f:
            f.write(
                json.dumps(
                    await enricher._enrich_keys(
                        ["databricks", "snowflake", "confluent"]
                    ),
                    indent=2,
                )
            )
    finally:
        await enricher.close()


if __name__ == "__main__":
    asyncio.run(main())
