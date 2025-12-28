from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession, async_sessionmaker
from sqlalchemy import Table, Column, Integer, String, Boolean, DateTime, Text, MetaData, select, update, func, ForeignKey, UniqueConstraint, Index
from sqlalchemy.dialects.postgresql import JSONB
from state_manager import IStateStore, RowState, RowStage, JobProgress
from datetime import datetime
import logging

logger = logging.getLogger(__name__)


class PostgresStateStore(IStateStore):
    def __init__(
        self,
        connection_string: str,
        pool_size: int = 10,
        max_overflow: int = 20,
        pool_pre_ping: bool = True,
        echo: bool = False,
    ):
        self.engine = create_async_engine(
            connection_string,
            pool_size=pool_size,
            max_overflow=max_overflow,
            pool_pre_ping=pool_pre_ping,
            echo=echo,
        )
        self.async_session_maker = async_sessionmaker(
            self.engine,
            class_=AsyncSession,
            expire_on_commit=False,
        )
        self.metadata = MetaData()
        self._define_tables()

    def _define_tables(self):
        self.jobs_table = Table(
            "jobs",
            self.metadata,
            Column("job_id", String(255), primary_key=True),
            Column("total_rows", Integer, nullable=False),
            Column("started_at", DateTime, nullable=False, server_default=func.now()),
            Column("status", String(20), nullable=False, server_default="RUNNING"),
            Column("created_at", DateTime, nullable=False, server_default=func.now()),
            Column(
                "updated_at",
                DateTime,
                nullable=False,
                server_default=func.now(),
                onupdate=func.now(),
            ),
        )

        self.row_states_table = Table(
            "row_states",
            self.metadata,
            Column("id", Integer, primary_key=True, autoincrement=True),
            Column("job_id", String(255), ForeignKey("jobs.job_id", ondelete="CASCADE"), nullable=False),
            Column("key", String(500), nullable=False),
            Column("stage", String(50), nullable=False),
            Column("serp_data", JSONB),
            Column("decision", JSONB),
            Column("crawl_results", JSONB),
            Column("extracted_data", JSONB),
            Column("error", Text),
            Column("created_at", DateTime, nullable=False, server_default=func.now()),
            Column(
                "updated_at",
                DateTime,
                nullable=False,
                server_default=func.now(),
                onupdate=func.now(),
            ),
            UniqueConstraint("job_id", "key", name="uq_job_key"),
            Index("idx_row_states_job_id", "job_id"),
            Index("idx_row_states_stage", "job_id", "stage"),
            Index("idx_row_states_updated_at", "updated_at"),
        )

    async def initialize_database(self):
        async with self.engine.begin() as conn:
            await conn.run_sync(self.metadata.create_all)

    async def save_row_state(self, job_id: str, state: RowState) -> None:
        async with self.async_session_maker() as session:
            async with session.begin():
                stmt = select(self.row_states_table).where(
                    self.row_states_table.c.job_id == job_id,
                    self.row_states_table.c.key == state.key,
                )
                result = await session.execute(stmt)
                existing = result.first()

                row_data = {
                    "job_id": job_id,
                    "key": state.key,
                    "stage": state.stage.name,
                    "serp_data": state.serp_data,
                    "decision": state.decision,
                    "crawl_results": state.crawl_results,
                    "extracted_data": state.extracted_data,
                    "error": state.error,
                    "updated_at": datetime.utcnow(),
                }

                if existing:
                    stmt = (
                        update(self.row_states_table)
                        .where(
                            self.row_states_table.c.job_id == job_id,
                            self.row_states_table.c.key == state.key,
                        )
                        .values(**row_data)
                    )
                    await session.execute(stmt)
                else:
                    row_data["created_at"] = state.created_at
                    stmt = self.row_states_table.insert().values(**row_data)
                    await session.execute(stmt)

    async def get_row_state(self, job_id: str, key: str) -> RowState | None:
        async with self.async_session_maker() as session:
            stmt = select(self.row_states_table).where(
                self.row_states_table.c.job_id == job_id,
                self.row_states_table.c.key == key,
            )
            result = await session.execute(stmt)
            row = result.first()

            if not row:
                return None

            try:
                stage = RowStage[row.stage]
            except KeyError:
                logger.error(f"Invalid stage value in database: {row.stage}")
                stage = RowStage.FAILED

            return RowState(
                key=row.key,
                stage=stage,
                serp_data=row.serp_data,
                decision=row.decision,
                crawl_results=row.crawl_results,
                extracted_data=row.extracted_data,
                error=row.error,
                created_at=row.created_at,
                updated_at=row.updated_at,
            )

    async def get_rows_at_stage(
        self, job_id: str, stage: RowStage
    ) -> list[RowState]:
        async with self.async_session_maker() as session:
            stmt = (
                select(self.row_states_table)
                .where(
                    self.row_states_table.c.job_id == job_id,
                    self.row_states_table.c.stage == stage.name,
                )
                .order_by(self.row_states_table.c.created_at)
            )

            result = await session.execute(stmt)
            rows = result.all()

            states = []
            for row in rows:
                try:
                    row_stage = RowStage[row.stage]
                except KeyError:
                    logger.error(f"Invalid stage value in database: {row.stage}")
                    row_stage = RowStage.FAILED

                states.append(
                    RowState(
                        key=row.key,
                        stage=row_stage,
                        serp_data=row.serp_data,
                        decision=row.decision,
                        crawl_results=row.crawl_results,
                        extracted_data=row.extracted_data,
                        error=row.error,
                        created_at=row.created_at,
                        updated_at=row.updated_at,
                    )
                )

            return states

    async def get_job_progress(self, job_id: str) -> JobProgress:
        async with self.async_session_maker() as session:
            job_stmt = select(self.jobs_table).where(
                self.jobs_table.c.job_id == job_id
            )
            job_result = await session.execute(job_stmt)
            job_row = job_result.first()

            if not job_row:
                raise ValueError(f"Job {job_id} not found")

            count_stmt = (
                select(self.row_states_table.c.stage, func.count().label("count"))
                .where(self.row_states_table.c.job_id == job_id)
                .group_by(self.row_states_table.c.stage)
            )

            count_result = await session.execute(count_stmt)
            rows_by_stage = {}
            for row in count_result.all():
                try:
                    stage = RowStage[row.stage]
                    rows_by_stage[stage] = row.count
                except KeyError:
                    logger.error(f"Invalid stage value in database: {row.stage}")

            from state_manager import JobStatus

            return JobProgress(
                job_id=job_id,
                total_rows=job_row.total_rows,
                rows_by_stage=rows_by_stage,
                started_at=job_row.started_at,
                status=JobStatus(job_row.status),
            )

    async def create_job(
        self, job_id: str, total_rows: int, status: "JobStatus" = None
    ) -> None:
        from state_manager import JobStatus

        if status is None:
            status = JobStatus.RUNNING

        async with self.async_session_maker() as session:
            async with session.begin():
                stmt = self.jobs_table.insert().values(
                    job_id=job_id,
                    total_rows=total_rows,
                    started_at=datetime.utcnow(),
                    status=status.value,
                )
                await session.execute(stmt)

    async def bulk_create_rows(self, job_id: str, row_keys: list[str]) -> None:
        from state_manager import RowStage

        async with self.async_session_maker() as session:
            async with session.begin():
                now = datetime.utcnow()
                rows_data = [
                    {
                        "job_id": job_id,
                        "key": key,
                        "stage": RowStage.PENDING.name,
                        "created_at": now,
                        "updated_at": now,
                    }
                    for key in row_keys
                ]

                if rows_data:
                    stmt = self.row_states_table.insert().values(rows_data)
                    await session.execute(stmt)

    async def set_job_status(self, job_id: str, status: "JobStatus") -> None:
        async with self.async_session_maker() as session:
            async with session.begin():
                stmt = (
                    update(self.jobs_table)
                    .where(self.jobs_table.c.job_id == job_id)
                    .values(status=status.value, updated_at=datetime.utcnow())
                )
                await session.execute(stmt)

    async def get_job_status(self, job_id: str) -> "JobStatus":
        from state_manager import JobStatus

        async with self.async_session_maker() as session:
            stmt = select(self.jobs_table.c.status).where(
                self.jobs_table.c.job_id == job_id
            )
            result = await session.execute(stmt)
            row = result.first()

            if not row:
                raise ValueError(f"Job {job_id} not found")

            return JobStatus(row.status)

    async def close(self):
        await self.engine.dispose()
