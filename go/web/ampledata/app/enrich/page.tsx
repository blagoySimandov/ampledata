"use client";

import { useState, useEffect, useRef } from "react";
import { FileUpload } from "@/components/file-upload";
import { DataGrid } from "@/components/data-grid";
import { EnrichmentDrawer } from "@/components/enrichment-drawer";
import { useUser } from "@/hooks";
import type { DataRow, Column } from "@/lib/types";
import { Header } from "./components";
import { toast } from "sonner";
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
		case "string":
		case "number":
		case "boolean":
		case "date":
			return dataType as ColumnType;
		default:
			return "string";
	}
}

export default function DataEnrichmentPage() {
	const user = useUser();
	const [data, setData] = useState<DataRow[]>([]);
	const [columns, setColumns] = useState<Column[]>([]);
	const [fileName, setFileName] = useState<string>("");
	const [currentJobId, setCurrentJobId] = useState<string | null>(null);
	const [isEnriching, setIsEnriching] = useState(false);
	const [enrichmentProgress, setEnrichmentProgress] = useState(0);
	const [drawerOpen, setDrawerOpen] = useState(false);

	const requestSignedURL = useRequestSignedURL();
	const uploadFile = useUploadFile();
	const startJob = useStartJob();



	const handleFileUpload = async (
		uploadedData: DataRow[],
		uploadedColumns: string[],
		name: string,
		file: File
	) => {
		setData(uploadedData);
		const columnsWithTypes: Column[] = uploadedColumns.map((col) => ({
			name: col,
			dataType: "string",
		}));
		setColumns(columnsWithTypes);
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

	const handleAddColumn = (columnName: string, dataType: string) => {
		const newColumn: Column = {
			name: columnName,
			dataType: dataType,
		};
		setColumns([...columns, newColumn]);
		const updatedData = data.map((row) => ({ ...row, [columnName]: null }));
		setData(updatedData);
	};

	const handleColumnNameChange = (oldName: string, newName: string) => {
		const updatedColumns = columns.map((col) =>
			col.name === oldName ? { ...col, name: newName } : col
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
			newRow[col.name] = null;
		});
		setData([...data, newRow]);
	};

	const handleRemoveRow = (rowIndex: number) => {
		const updatedData = data.filter((_, index) => index !== rowIndex);
		setData(updatedData);
	};

	const handleRemoveColumn = (columnName: string) => {
		setColumns(columns.filter((col) => col.name !== columnName));
		const updatedData = data.map((row) => {
			const newRow = { ...row };
			delete newRow[columnName];
			return newRow;
		});
		setData(updatedData);
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
	const toastIdRef = useRef<string | number | null>(null);

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
				if (enrichmentColumnName) {
					setColumns((prevColumns) =>
						prevColumns.map((col) =>
							col.name === enrichmentColumnName
								? { ...col, isEnriching: false }
								: col
						)
					);
				}

				if (toastIdRef.current) {
					if (jobProgress.data.status === "COMPLETED") {
						toast.success(
							`Enrichment complete! ${totalRows} rows processed successfully.`,
							{
								id: toastIdRef.current,
								duration: 5000,
								action: {
									label: "View Details",
									onClick: () => setDrawerOpen(true),
								},
							}
						);
					} else {
						toast.error("Enrichment was cancelled.", {
							id: toastIdRef.current,
							duration: 5000,
						});
					}
					toastIdRef.current = null;
				}
			}
		}
	}, [jobProgress.data, enrichmentColumnName]);

	const isEnrichmentInProgress =
		isEnriching &&
		jobProgress.data &&
		jobProgress.data.status !== "COMPLETED" &&
		jobProgress.data.status !== "CANCELLED";

	useEffect(() => {
		if (drawerOpen && toastIdRef.current) {
			toast.dismiss(toastIdRef.current);
			toastIdRef.current = null;
			return;
		}

		if (!drawerOpen && isEnrichmentInProgress) {
			const completedRows = jobProgress.data?.rows_by_stage.COMPLETED || 0;
			const totalRows = jobProgress.data?.total_rows || data.length;
			const progress =
				totalRows > 0 ? (completedRows / totalRows) * 100 : 0;
			const rowsProcessed = Math.floor((progress / 100) * totalRows);

			const toastMessage = `Enriching data... ${rowsProcessed} of ${totalRows} rows processed (${Math.round(progress)}%)`;

			if (!toastIdRef.current) {
				const toastId = toast.loading(toastMessage, {
					duration: Infinity,
					action: {
						label: "View Details",
						onClick: () => setDrawerOpen(true),
					},
				});
				toastIdRef.current = toastId;
			} else {
				toast.loading(toastMessage, {
					id: toastIdRef.current,
					duration: Infinity,
					action: {
						label: "View Details",
						onClick: () => setDrawerOpen(true),
					},
				});
			}
		} else if (!drawerOpen && isEnriching && enrichmentJobId && !jobProgress.data) {
			if (!toastIdRef.current) {
				const toastId = toast.loading(
					`Enriching data... Starting enrichment process...`,
					{
						duration: Infinity,
						action: {
							label: "View Details",
							onClick: () => setDrawerOpen(true),
						},
					}
				);
				toastIdRef.current = toastId;
			}
		}
	}, [
		drawerOpen,
		isEnrichmentInProgress,
		isEnriching,
		enrichmentJobId,
		jobProgress.data,
		enrichmentProgress,
		data.length,
	]);

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

	const convertDataToCSVFile = (): File => {
		const csv = [
			columns.map((col) => col.name).join(","),
			...data.map((row) =>
				columns
					.map((col) => {
						const value = row[col.name];
						if (value === null || value === undefined) return "";
						return typeof value === "string" && value.includes(",")
							? `"${value}"`
							: value;
					})
					.join(",")
			),
		].join("\n");

		const blob = new Blob([csv], { type: "text/csv" });
		return new File([blob], fileName || "data.csv", { type: "text/csv" });
	};

	const handleEnrich = async (
		keyColumn: string,
		columnName: string,
		dataType: string
	) => {
		setColumns(
			columns.map((col) =>
				col.name === columnName ? { ...col, isEnriching: true } : col
			)
		);

		setIsEnriching(true);
		setEnrichmentProgress(0);
		setDrawerOpen(false);
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
			const csvFile = convertDataToCSVFile();

			const signedURLResponse = await requestSignedURL.mutateAsync({
				contentType: csvFile.type || "text/csv",
				length: csvFile.size,
			});

			await uploadFile.mutateAsync({
				signedUrl: signedURLResponse.url,
				file: csvFile,
			});

			const newJobId = signedURLResponse.jobId;
			setCurrentJobId(newJobId);

			const startResponse = await startJob.mutateAsync({
				jobId: newJobId,
				data: {
					key_column: keyColumn,
					columns_metadata: columnsMetadata,
				},
			});

			setEnrichmentJobId(startResponse.job_id);

			const toastId = toast.loading(
				`Enriching data... 0 of ${data.length} rows processed (0%)`,
				{
					duration: Infinity,
					action: {
						label: "View Details",
						onClick: () => setDrawerOpen(true),
					},
				}
			);
			toastIdRef.current = toastId;
		} catch (error) {
			console.error("Failed to start enrichment job:", error);
			if (toastIdRef.current) {
				toast.error("Failed to start enrichment job", {
					id: toastIdRef.current,
				});
				toastIdRef.current = null;
			}
			setColumns(
				columns.map((col) =>
					col.name === columnName ? { ...col, isEnriching: false } : col
				)
			);
			setIsEnriching(false);
			setDrawerOpen(false);
			setEnrichmentKeyColumn(null);
			setEnrichmentColumnName(null);
		}
	};

	const handleExport = () => {
		const csv = [
			columns.map((col) => col.name).join(","),
			...data.map((row) =>
				columns
					.map((col) => {
						const value = row[col.name];
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

	if (!user) {
		return (
			<div className="flex items-center justify-center min-h-[50vh]">
				<p className="text-lg">Loading...</p>
			</div>
		);
	}

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
						onRemoveColumn={handleRemoveColumn}
						onRemoveRow={handleRemoveRow}
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
