from abc import ABC, abstractmethod


class IQueryBuilder(ABC):
    @abstractmethod
    def build(self, entity: str) -> str:
        pass


class QueryBuilder(IQueryBuilder):
    def __init__(
        self,
        columns_to_search_for: list[str],
        entity_type: str | None = None,
    ):
        """

        entity_type: Optional disambiguation hint (e.g., "company", "person", "restaurant")
        Will be appended to the query to help the search engine.
        """
        self.columns_to_search_for = columns_to_search_for
        self.entity_type = entity_type

    def build(self, entity: str) -> str:
        parts = []

        parts.append(entity)

        if self.entity_type:  # append entity hint if provided
            parts.append(self.entity_type)

        parts.extend(self.columns_to_search_for)

        return " ".join(parts)
