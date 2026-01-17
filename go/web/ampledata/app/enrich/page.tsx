"use client";

import { useState, useEffect } from "react";
import { FileUpload } from "@/components/file-upload";
import { DataGrid } from "@/components/data-grid";
import { EnrichmentDrawer } from "@/components/enrichment-drawer";
import { useUser } from "@/hooks";
import type { DataRow } from "@/lib/types";
import { Header } from "./components";
import {
	useRequestSignedURL,
	useUploadFile,
	useStartJob,
	useJobProgress,
	useJobResults,
	type ColumnType,
	type ColumnMetadata,
} from "@/api";

function mapDataTypeToColumnType(dataType: string): ColumnType {
	switch (dataType) {
		case "email":
		case "phone":
		case "company":
		case "location":
		case "text":
			return "string";
		case "number":
			return "number";
		case "boolean":
			return "boolean";
		default:
			return "string";
	}
}

export default function DataEnrichmentPage() {
	const user = useUser();
	const [data, setData] = useState<DataRow[]>([]);
	const [columns, setColumns] = useState<string[]>([]);
	const [fileName, setFileName] = useState<string>("");
	const [currentJobId, setCurrentJobId] = useState<string | null>(null);
	const [isEnriching, setIsEnriching] = useState(false);
	const [enrichmentProgress, setEnrichmentProgress] = useState(0);
	const [drawerOpen, setDrawerOpen] = useState(false);

	const requestSignedURL = useRequestSignedURL();
	const uploadFile = useUploadFile();
	const startJob = useStartJob();

	if (!user) {
		return (
			<div className="flex items-center justify-center min-h-[50vh]">
				<p className="text-lg">Loading...</p>
			</div>
		);
	}

	const handleFileUpload = async (
		uploadedData: DataRow[],
		uploadedColumns: string[],
		name: string,
		file: File
	) => {
		setData(uploadedData);
		setColumns(uploadedColumns);
		setFileName(name);

		try {
			const signedURLResponse = await requestSignedURL.mutateAsync({
				contentType: file.type || "text/csv",
				length: file.size,
			});

			await uploadFile.mutateAsync({
				signedUrl: signedURLResponse.url,
				file: file,
			});

			setCurrentJobId(signedURLResponse.jobId);
		} catch (error) {
			console.error("Failed to upload file:", error);
		}
	};

	const handleAddColumn = (columnName: string) => {
		setColumns([...columns, columnName]);
		const updatedData = data.map((row) => ({ ...row, [columnName]: null }));
		setData(updatedData);
	};

	const handleColumnNameChange = (oldName: string, newName: string) => {
		const updatedColumns = columns.map((col) =>
			col === oldName ? newName : col
		);
		setColumns(updatedColumns);

		const updatedData = data.map((row) => {
			const newRow = { ...row };
			newRow[newName] = row[oldName];
			delete newRow[oldName];
			return newRow;
		});
		setData(updatedData);
	};

	const handleCellChange = (
		rowIndex: number,
		columnName: string,
		value: string
	) => {
		const updatedData = [...data];
		updatedData[rowIndex] = {
			...updatedData[rowIndex],
			[columnName]: value,
		};
		setData(updatedData);
	};

	const handleAddRow = () => {
		const newRow: DataRow = {};
		columns.forEach((col) => {
			newRow[col] = null;
		});
		setData([...data, newRow]);
	};

	const [enrichmentJobId, setEnrichmentJobId] = useState<string | null>(null);
	const [enrichmentKeyColumn, setEnrichmentKeyColumn] = useState<
		string | null
	>(null);
	const [enrichmentColumnName, setEnrichmentColumnName] = useState<
		string | null
	>(null);
	const [resultsFetched, setResultsFetched] = useState(false);
	const [shouldPollProgress, setShouldPollProgress] = useState(true);

	const jobProgress = useJobProgress(enrichmentJobId || "", {
		enabled: !!enrichmentJobId && isEnriching,
		refetchInterval: shouldPollProgress ? 2000 : undefined,
	});

	const jobResults = useJobResults(enrichmentJobId || "", undefined, {
		enabled:
			!!enrichmentJobId &&
			jobProgress.data?.status === "COMPLETED" &&
			!resultsFetched,
	});

	useEffect(() => {
		if (jobProgress.data) {
			const completedRows = jobProgress.data.rows_by_stage.COMPLETED || 0;
			const totalRows = jobProgress.data.total_rows;
			const progress =
				totalRows > 0 ? (completedRows / totalRows) * 100 : 0;
			setEnrichmentProgress(progress);

			if (
				jobProgress.data.status === "COMPLETED" ||
				jobProgress.data.status === "CANCELLED"
			) {
				setShouldPollProgress(false);
				setIsEnriching(false);
			}
		}
	}, [jobProgress.data]);

	useEffect(() => {
		if (
			jobResults.data &&
			jobResults.data.length > 0 &&
			enrichmentKeyColumn &&
			enrichmentColumnName &&
			!resultsFetched
		) {
			const enrichedData = [...data];
			jobResults.data.forEach((result) => {
				const rowIndex = enrichedData.findIndex(
					(row) => String(row[enrichmentKeyColumn]) === result.key
				);
				if (rowIndex !== -1) {
					const extractedValue =
						result.extracted_data[enrichmentColumnName];
					if (extractedValue !== undefined) {
						enrichedData[rowIndex][enrichmentColumnName] =
							extractedValue as string | number | boolean | null;
					}
				}
			});
			setData(enrichedData);
			setResultsFetched(true);
			setEnrichmentKeyColumn(null);
			setEnrichmentColumnName(null);
		}
	}, [
		jobResults.data,
		enrichmentKeyColumn,
		enrichmentColumnName,
		resultsFetched,
		data,
		columns,
	]);

	const handleEnrich = async (
		keyColumn: string,
		columnName: string,
		dataType: string
	) => {
		if (!currentJobId) {
			console.error("No job ID available");
			return;
		}

		setIsEnriching(true);
		setEnrichmentProgress(0);
		setDrawerOpen(true);
		setEnrichmentKeyColumn(keyColumn);
		setEnrichmentColumnName(columnName);
		setResultsFetched(false);
		setShouldPollProgress(true);

		const columnType = mapDataTypeToColumnType(dataType);
		const columnsMetadata: ColumnMetadata[] = [
			{
				name: columnName,
				type: columnType,
				description: `Enriched ${dataType} data`,
			},
		];

		try {
			const startResponse = await startJob.mutateAsync({
				jobId: currentJobId,
				data: {
					key_column: keyColumn,
					columns_metadata: columnsMetadata,
				},
			});

			setEnrichmentJobId(startResponse.job_id);
		} catch (error) {
			console.error("Failed to start enrichment job:", error);
			setIsEnriching(false);
			setDrawerOpen(false);
			setEnrichmentKeyColumn(null);
			setEnrichmentColumnName(null);
		}
	};

	const handleExport = () => {
		const csv = [
			columns.join(","),
			...data.map((row) =>
				columns
					.map((col) => {
						const value = row[col];
						if (value === null || value === undefined) return "";
						return typeof value === "string" && value.includes(",")
							? `"${value}"`
							: value;
					})
					.join(",")
			),
		].join("\n");

		const blob = new Blob([csv], { type: "text/csv" });
		const url = URL.createObjectURL(blob);
		const a = document.createElement("a");
		a.href = url;
		a.download = `enriched-${fileName || "data"}.csv`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	};

	return (
		<div className="min-h-screen bg-background">
			<Header hasData={data.length > 0} onExport={handleExport} />

			<main className="container mx-auto px-4 py-6 sm:px-6 lg:px-8">
				{data.length === 0 ? (
					<FileUpload onFileUpload={handleFileUpload} />
				) : (
					<DataGrid
						data={data}
						columns={columns}
						onAddColumn={handleAddColumn}
						onEnrich={handleEnrich}
						onColumnNameChange={handleColumnNameChange}
						onCellChange={handleCellChange}
						onAddRow={handleAddRow}
						isEnriching={isEnriching}
					/>
				)}
			</main>

			<EnrichmentDrawer
				open={drawerOpen}
				onOpenChange={setDrawerOpen}
				progress={enrichmentProgress}
				isEnriching={isEnriching}
				totalRows={data.length}
			/>
		</div>
	);
}
