from abc import ABC, abstractmethod
from enum import Enum, auto
from dataclasses import dataclass, field
from typing import Any, Callable, Awaitable
import asyncio
from datetime import datetime

# workflow steps
# 1. gets serps for rows.... -> save in db
# 2. get serps from db and make decision using llm -> save in db
# 3. get decision based on decision crawl and enrich. save results in db
# 4. get crawl results + enrich results and query llm for missing
# columns using the new crawl snippets. -> save in db
# 5. keep track of progress during the whole process and somehow make it cancellable ?
#
# we need to keep track of state for a few differejt things
# 1.


class RowStage(Enum):
    PENDING = auto()
    SERP_FETCHED = auto()
    DECISION_MADE = auto()
    CRAWLED = auto()
    ENRICHED = auto()
    COMPLETED = auto()
    FAILED = auto()
    CANCELLED = auto()


@dataclass
class RowState:
    key: str
    stage: RowStage = RowStage.PENDING
    serp_data: dict | None = None
    decision: dict | None = None
    crawl_results: dict | None = None
    extracted_data: dict | None = None
    error: str | None = None
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)

    def to_dict(self) -> dict:
        return {
            "key": self.key,
            "stage": self.stage.name,
            "serp_data": self.serp_data,
            "decision": self.decision,
            "crawl_results": self.crawl_results,
            "extracted_data": self.extracted_data,
            "error": self.error,
            "created_at": self.created_at.isoformat(),
            "updated_at": self.updated_at.isoformat(),
        }


@dataclass
class JobProgress:
    job_id: str
    total_rows: int
    rows_by_stage: dict[RowStage, int] = field(default_factory=dict)
    started_at: datetime = field(default_factory=datetime.utcnow)
    is_cancelled: bool = False


class IStateStore(ABC):
    """Persistence layer - swap in postgres, redis, sqlite, etc."""

    @abstractmethod
    async def save_row_state(self, job_id: str, state: RowState) -> None:
        pass

    @abstractmethod
    async def get_row_state(self, job_id: str, key: str) -> RowState | None:
        pass

    @abstractmethod
    async def get_rows_at_stage(self, job_id: str, stage: RowStage) -> list[RowState]:
        pass

    @abstractmethod
    async def get_job_progress(self, job_id: str) -> JobProgress:
        pass

    @abstractmethod
    async def set_job_cancelled(self, job_id: str) -> None:
        pass

    @abstractmethod
    async def is_job_cancelled(self, job_id: str) -> bool:
        pass


class IStateManager(ABC):
    """Orchestrates transitions and checks cancellation."""

    @abstractmethod
    def generate_job_id(self) -> str:
        pass

    @abstractmethod
    async def initialize_job(self, job_id: str, row_keys: list[str]) -> None:
        pass

    @abstractmethod
    async def transition(
        self, job_id: str, key: str, to_stage: RowStage, data_update: dict | None = None
    ) -> RowState:
        pass

    @abstractmethod
    async def get_pending_for_stage(
        self, job_id: str, stage: RowStage
    ) -> list[RowState]:
        pass

    @abstractmethod
    async def cancel(self, job_id: str) -> None:
        pass

    @abstractmethod
    async def check_cancelled(self, job_id: str) -> bool:
        pass

    @abstractmethod
    async def get_progress(self, job_id: str) -> JobProgress:
        pass


class StateManager(IStateManager):
    def __init__(self, store: IStateStore):
        self.store = store
        self._cancel_events: dict[str, asyncio.Event] = {}

    async def initialize_job(self, job_id: str, row_keys: list[str]) -> None:
        self._cancel_events[job_id] = asyncio.Event()
        for key in row_keys:
            state = RowState(key=key)
            await self.store.save_row_state(job_id, state)

    def generate_job_id(self) -> str:
        return f"job_{datetime.utcnow().timestamp()}"

    async def transition(
        self, job_id: str, key: str, to_stage: RowStage, data_update: dict | None = None
    ) -> RowState:
        # Check cancellation before any transition
        if await self.check_cancelled(job_id):
            state = await self.store.get_row_state(job_id, key)
            if state and state.stage not in (RowStage.COMPLETED, RowStage.FAILED):
                state.stage = RowStage.CANCELLED
                state.updated_at = datetime.utcnow()
                await self.store.save_row_state(job_id, state)
            raise CancelledException(job_id)

        state = await self.store.get_row_state(job_id, key)
        if not state:
            raise ValueError(f"No state found for {key}")

        state.stage = to_stage
        state.updated_at = datetime.utcnow()

        if data_update:
            if "serp_data" in data_update:
                state.serp_data = data_update["serp_data"]
            if "decision" in data_update:
                state.decision = data_update["decision"]
            if "crawl_results" in data_update:
                state.crawl_results = data_update["crawl_results"]
            if "extracted_data" in data_update:
                state.extracted_data = data_update["extracted_data"]
            if "error" in data_update:
                state.error = data_update["error"]
                state.stage = RowStage.FAILED

        await self.store.save_row_state(job_id, state)
        return state

    async def get_pending_for_stage(
        self, job_id: str, stage: RowStage
    ) -> list[RowState]:
        # Map: which stage should rows be at to be ready for this step
        prerequisite = {
            RowStage.SERP_FETCHED: RowStage.PENDING,
            RowStage.DECISION_MADE: RowStage.SERP_FETCHED,
            RowStage.CRAWLED: RowStage.DECISION_MADE,
            RowStage.ENRICHED: RowStage.CRAWLED,
            RowStage.COMPLETED: RowStage.ENRICHED,
        }
        required_stage = prerequisite.get(stage, RowStage.PENDING)
        return await self.store.get_rows_at_stage(job_id, required_stage)

    async def cancel(self, job_id: str) -> None:
        await self.store.set_job_cancelled(job_id)
        if job_id in self._cancel_events:
            self._cancel_events[job_id].set()

    async def check_cancelled(self, job_id: str) -> bool:
        # Fast path: check in-memory event
        if job_id in self._cancel_events and self._cancel_events[job_id].is_set():
            return True
        # Slow path: check DB (for multi-process scenarios)
        return await self.store.is_job_cancelled(job_id)

    async def get_progress(self, job_id: str) -> JobProgress:
        return await self.store.get_job_progress(job_id)


class CancelledException(Exception):
    def __init__(self, job_id: str):
        self.job_id = job_id
        super().__init__(f"Job {job_id} was cancelled")
