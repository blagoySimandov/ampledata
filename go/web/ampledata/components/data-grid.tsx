"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Select,
	SelectContent,
	SelectItem,
	SelectTrigger,
	SelectValue,
} from "@/components/ui/select";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { Plus, Sparkles, PlusCircle, CircleMinus } from "lucide-react";
import type { DataRow, Column } from "@/lib/types";

interface DataGridProps {
	data: DataRow[];
	columns: Column[];
	onAddColumn: (columnName: string, dataType: string) => void;
	onEnrich: (keyColumn: string, columnName: string, dataType: string) => void;
	onColumnNameChange: (oldName: string, newName: string) => void;
	onCellChange: (rowIndex: number, columnName: string, value: string) => void;
	onAddRow: () => void;
	onRemoveColumn: (columnName: string) => void;
	onRemoveRow: (rowIndex: number) => void;
	isEnriching: boolean;
}

export function DataGrid({
	data,
	columns,
	onAddColumn,
	onEnrich,
	onColumnNameChange,
	onCellChange,
	onAddRow,
	onRemoveColumn,
	onRemoveRow,
	isEnriching,
}: DataGridProps) {
	const [newColumnName, setNewColumnName] = useState("");
	const [newColumnDataType, setNewColumnDataType] = useState("");
	const [selectedKeyColumn, setSelectedKeyColumn] = useState("");
	const [selectedColumn, setSelectedColumn] = useState("");
	const [addDialogOpen, setAddDialogOpen] = useState(false);
	const [enrichDialogOpen, setEnrichDialogOpen] = useState(false);
	const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
	const [columnToDelete, setColumnToDelete] = useState<string | null>(null);
	const [deleteRowDialogOpen, setDeleteRowDialogOpen] = useState(false);
	const [rowToDelete, setRowToDelete] = useState<number | null>(null);

	const handleAddColumn = () => {
		if (newColumnName.trim() && newColumnDataType) {
			onAddColumn(newColumnName.trim(), newColumnDataType);
			setNewColumnName("");
			setNewColumnDataType("");
			setAddDialogOpen(false);
		}
	};

	const handleEnrich = () => {
		if (selectedKeyColumn && selectedColumn) {
			const columnToEnrich = columns.find(
				(col) => col.name === selectedColumn
			);
			if (columnToEnrich) {
				onEnrich(
					selectedKeyColumn,
					selectedColumn,
					columnToEnrich.dataType
				);
				setEnrichDialogOpen(false);
				setSelectedKeyColumn("");
				setSelectedColumn("");
			}
		}
	};

	const handleDeleteClick = (columnName: string) => {
		setColumnToDelete(columnName);
		setDeleteDialogOpen(true);
	};

	const handleDeleteConfirm = () => {
		if (columnToDelete) {
			onRemoveColumn(columnToDelete);
			setDeleteDialogOpen(false);
			setColumnToDelete(null);
		}
	};

	const handleDeleteCancel = () => {
		setDeleteDialogOpen(false);
		setColumnToDelete(null);
	};

	const handleDeleteRowClick = (rowIndex: number) => {
		setRowToDelete(rowIndex);
		setDeleteRowDialogOpen(true);
	};

	const handleDeleteRowConfirm = () => {
		if (rowToDelete !== null) {
			onRemoveRow(rowToDelete);
			setDeleteRowDialogOpen(false);
			setRowToDelete(null);
		}
	};

	const handleDeleteRowCancel = () => {
		setDeleteRowDialogOpen(false);
		setRowToDelete(null);
	};

	const emptyColumns = columns.filter(
		(col) =>
			!col.isEnriching &&
			data.every(
				(row) =>
					row[col.name] === null ||
					row[col.name] === undefined ||
					row[col.name] === ""
			)
	);

	const columnsWithData = columns.filter((col) =>
		data.some(
			(row) =>
				row[col.name] !== null &&
				row[col.name] !== undefined &&
				row[col.name] !== ""
		)
	);

	return (
		<div className="space-y-4">
			<div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
				<div className="flex items-center gap-2">
					<p className="text-sm text-muted-foreground">
						{data.length} rows × {columns.length} columns
					</p>
				</div>

				<div className="flex flex-wrap gap-2">
					<Button
						variant="outline"
						size="sm"
						className="gap-2 bg-transparent"
						onClick={onAddRow}
					>
						<PlusCircle className="h-4 w-4" />
						Add Row
					</Button>

					<Dialog
						open={addDialogOpen}
						onOpenChange={setAddDialogOpen}
					>
						<DialogTrigger asChild>
							<Button
								variant="outline"
								size="sm"
								className="gap-2 bg-transparent"
							>
								<Plus className="h-4 w-4" />
								Add Column
							</Button>
						</DialogTrigger>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Add New Column</DialogTitle>
								<DialogDescription>
									Create a new column to enrich your data
								</DialogDescription>
							</DialogHeader>
							<div className="space-y-4 py-4">
								<div className="space-y-2">
									<Label htmlFor="column-name">
										Column Name
									</Label>
									<Input
										id="column-name"
										placeholder="e.g., Email, Phone, Company"
										value={newColumnName}
										onChange={(e) =>
											setNewColumnName(e.target.value)
										}
										onKeyDown={(e) =>
											e.key === "Enter" &&
											handleAddColumn()
										}
									/>
								</div>
								<div className="space-y-2">
									<Label htmlFor="new-column-data-type">
										Data Type
									</Label>
									<Select
										value={newColumnDataType}
										onValueChange={setNewColumnDataType}
									>
										<SelectTrigger id="new-column-data-type">
											<SelectValue placeholder="Select data type" />
										</SelectTrigger>
										<SelectContent>
											<SelectItem value="string">
												String
											</SelectItem>
											<SelectItem value="number">
												Number
											</SelectItem>
											<SelectItem value="boolean">
												Boolean
											</SelectItem>
											<SelectItem value="date">
												Date
											</SelectItem>
										</SelectContent>
									</Select>
								</div>
							</div>
							<DialogFooter>
								<Button
									onClick={handleAddColumn}
									disabled={
										!newColumnName.trim() ||
										!newColumnDataType
									}
								>
									Add Column
								</Button>
							</DialogFooter>
						</DialogContent>
					</Dialog>

					<Dialog
						open={enrichDialogOpen}
						onOpenChange={setEnrichDialogOpen}
					>
						<DialogTrigger asChild>
							<Button
								size="sm"
								className="gap-2"
								disabled={
									isEnriching ||
									emptyColumns.length === 0 ||
									columnsWithData.length === 0
								}
							>
								<Sparkles className="h-4 w-4" />
								Enrich Data
							</Button>
						</DialogTrigger>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Enrich Data</DialogTitle>
								<DialogDescription>
									Select a column and data type to enrich
								</DialogDescription>
							</DialogHeader>
							<div className="space-y-4 py-4">
								<div className="space-y-2">
									<Label htmlFor="key-column-select">
										Key Column
									</Label>
									<Select
										value={selectedKeyColumn}
										onValueChange={setSelectedKeyColumn}
									>
										<SelectTrigger id="key-column-select">
											<SelectValue placeholder="Select key column" />
										</SelectTrigger>
										<SelectContent>
											{columnsWithData.map((col) => (
												<SelectItem
													key={col.name}
													value={col.name}
												>
													{col.name}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
								</div>
								<div className="space-y-2">
									<Label htmlFor="column-select">
										Column to Enrich
									</Label>
									<Select
										value={selectedColumn}
										onValueChange={setSelectedColumn}
									>
										<SelectTrigger id="column-select">
											<SelectValue placeholder="Select column" />
										</SelectTrigger>
										<SelectContent>
											{emptyColumns.map((col) => (
												<SelectItem
													key={col.name}
													value={col.name}
												>
													{col.name} ({col.dataType})
												</SelectItem>
											))}
										</SelectContent>
									</Select>
								</div>
							</div>
							<DialogFooter>
								<Button
									onClick={handleEnrich}
									disabled={
										!selectedKeyColumn || !selectedColumn
									}
								>
									Start Enrichment
								</Button>
							</DialogFooter>
						</DialogContent>
					</Dialog>

					<Dialog
						open={deleteDialogOpen}
						onOpenChange={setDeleteDialogOpen}
					>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Delete Column</DialogTitle>
								<DialogDescription>
									Are you sure you want to delete the column
									&quot;{columnToDelete}&quot;? This action
									cannot be undone and will remove all data in
									this column.
								</DialogDescription>
							</DialogHeader>
							<DialogFooter>
								<Button
									variant="outline"
									onClick={handleDeleteCancel}
								>
									Cancel
								</Button>
								<Button
									variant="destructive"
									onClick={handleDeleteConfirm}
								>
									Delete
								</Button>
							</DialogFooter>
						</DialogContent>
					</Dialog>

					<Dialog
						open={deleteRowDialogOpen}
						onOpenChange={setDeleteRowDialogOpen}
					>
						<DialogContent>
							<DialogHeader>
								<DialogTitle>Delete Row</DialogTitle>
								<DialogDescription>
									Are you sure you want to delete row{" "}
									{rowToDelete !== null
										? rowToDelete + 1
										: ""}
									? This action cannot be undone.
								</DialogDescription>
							</DialogHeader>
							<DialogFooter>
								<Button
									variant="outline"
									onClick={handleDeleteRowCancel}
								>
									Cancel
								</Button>
								<Button
									variant="destructive"
									onClick={handleDeleteRowConfirm}
								>
									Delete
								</Button>
							</DialogFooter>
						</DialogContent>
					</Dialog>
				</div>
			</div>

			<Card className="overflow-hidden">
				<div className="overflow-x-auto">
					<table className="w-full">
						<thead>
							<tr className="border-b bg-muted/50">
								{columns.map((col, index) => (
									<th
										key={index}
										className="px-2 py-3 text-left w-10"
									>
										<div className="flex items-center gap-2">
											<Input
												defaultValue={col.name}
												disabled={col.isEnriching}
												onBlur={(e) => {
													const newName =
														e.target.value.trim();
													if (
														newName &&
														newName !== col.name
													) {
														onColumnNameChange(
															col.name,
															newName
														);
													} else {
														e.target.value =
															col.name;
													}
												}}
												onKeyDown={(e) => {
													if (e.key === "Enter") {
														e.currentTarget.blur();
													}
												}}
												className="h-8 flex-1 text-sm font-semibold border-0 outline-0 bg-transparent hover:bg-muted/50 focus-visible:ring-1 focus-visible:ring-primary transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
											/>
											<Button
												variant="ghost"
												size="icon"
												onClick={() =>
													handleDeleteClick(col.name)
												}
												disabled={col.isEnriching}
												className="h-8 w-8 shrink-0 text-muted-foreground hover:text-destructive disabled:opacity-50 disabled:cursor-not-allowed"
											>
												<CircleMinus className="h-4 w-4 text-destructive" />
											</Button>
										</div>
									</th>
								))}
								<th className="px-2 py-3 text-center" />
							</tr>
						</thead>
						<tbody>
							{data.map((row, rowIndex) => (
								<tr
									key={rowIndex}
									className="border-b last:border-0 hover:bg-muted/30"
								>
									{columns.map((col, colIndex) => (
										<td
											key={colIndex}
											className="px-2 py-2"
										>
											<Input
												value={
													row[col.name] !== null &&
													row[col.name] !== undefined
														? String(row[col.name])
														: ""
												}
												onChange={(e) =>
													onCellChange(
														rowIndex,
														col.name,
														e.target.value
													)
												}
												disabled={col.isEnriching}
												placeholder="—"
												className="h-8 border-0 bg-transparent text-sm focus-visible:ring-1 focus-visible:ring-primary disabled:opacity-50 disabled:cursor-not-allowed"
											/>
										</td>
									))}
									<td className="text-center w-4">
										<Button
											variant="ghost"
											size="icon"
											onClick={() =>
												handleDeleteRowClick(rowIndex)
											}
											className="h-8 shrink-0 text-muted-foreground hover:text-destructive disabled:opacity-50 disabled:cursor-not-allowed"
										>
											<CircleMinus className="h-4 w-4 text-destructive" />
										</Button>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</Card>
		</div>
	);
}
