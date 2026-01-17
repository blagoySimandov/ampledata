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
import { Plus, Sparkles, PlusCircle } from "lucide-react";
import type { DataRow } from "@/lib/types";

interface DataGridProps {
	data: DataRow[];
	columns: string[];
	onAddColumn: (columnName: string) => void;
	onEnrich: (keyColumn: string, columnName: string, dataType: string) => void;
	onColumnNameChange: (oldName: string, newName: string) => void;
	onCellChange: (rowIndex: number, columnName: string, value: string) => void;
	onAddRow: () => void;
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
	isEnriching,
}: DataGridProps) {
	const [newColumnName, setNewColumnName] = useState("");
	const [selectedKeyColumn, setSelectedKeyColumn] = useState("");
	const [selectedColumn, setSelectedColumn] = useState("");
	const [selectedDataType, setSelectedDataType] = useState("");
	const [addDialogOpen, setAddDialogOpen] = useState(false);
	const [enrichDialogOpen, setEnrichDialogOpen] = useState(false);
	const [editingColumn, setEditingColumn] = useState<string | null>(null);
	const [editingColumnValue, setEditingColumnValue] = useState("");

	const handleAddColumn = () => {
		if (newColumnName.trim()) {
			onAddColumn(newColumnName.trim());
			setNewColumnName("");
			setAddDialogOpen(false);
		}
	};

	const handleEnrich = () => {
		if (selectedKeyColumn && selectedColumn && selectedDataType) {
			onEnrich(selectedKeyColumn, selectedColumn, selectedDataType);
			setEnrichDialogOpen(false);
			setSelectedKeyColumn("");
			setSelectedColumn("");
			setSelectedDataType("");
		}
	};

	const handleColumnNameEdit = (oldName: string) => {
		setEditingColumn(oldName);
		setEditingColumnValue(oldName);
	};

	const handleColumnNameSave = (oldName: string) => {
		if (editingColumnValue.trim() && editingColumnValue !== oldName) {
			onColumnNameChange(oldName, editingColumnValue.trim());
		}
		setEditingColumn(null);
		setEditingColumnValue("");
	};

	const emptyColumns = columns.filter((col) =>
		data.every(
			(row) =>
				row[col] === null || row[col] === undefined || row[col] === ""
		)
	);

	const columnsWithData = columns.filter((col) =>
		data.some(
			(row) =>
				row[col] !== null && row[col] !== undefined && row[col] !== ""
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
							</div>
							<DialogFooter>
								<Button
									onClick={handleAddColumn}
									disabled={!newColumnName.trim()}
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
													key={col}
													value={col}
												>
													{col}
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
													key={col}
													value={col}
												>
													{col}
												</SelectItem>
											))}
										</SelectContent>
									</Select>
								</div>
								<div className="space-y-2">
									<Label htmlFor="data-type-select">
										Data Type
									</Label>
									<Select
										value={selectedDataType}
										onValueChange={setSelectedDataType}
									>
										<SelectTrigger id="data-type-select">
											<SelectValue placeholder="Select data type" />
										</SelectTrigger>
										<SelectContent>
											<SelectItem value="email">
												Email
											</SelectItem>
											<SelectItem value="phone">
												Phone Number
											</SelectItem>
											<SelectItem value="company">
												Company Name
											</SelectItem>
											<SelectItem value="location">
												Location
											</SelectItem>
											<SelectItem value="number">
												Number
											</SelectItem>
											<SelectItem value="boolean">
												Boolean
											</SelectItem>
											<SelectItem value="text">
												Text
											</SelectItem>
										</SelectContent>
									</Select>
								</div>
							</div>
							<DialogFooter>
								<Button
									onClick={handleEnrich}
									disabled={
										!selectedKeyColumn ||
										!selectedColumn ||
										!selectedDataType
									}
								>
									Start Enrichment
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
										className="px-2 py-3 text-left"
									>
										{editingColumn === col ? (
											<Input
												value={editingColumnValue}
												onChange={(e) =>
													setEditingColumnValue(
														e.target.value
													)
												}
												onBlur={() =>
													handleColumnNameSave(col)
												}
												onKeyDown={(e) => {
													if (e.key === "Enter")
														handleColumnNameSave(
															col
														);
													if (e.key === "Escape")
														setEditingColumn(null);
												}}
												autoFocus
												className="h-8 text-sm font-semibold"
											/>
										) : (
											<button
												onClick={() =>
													handleColumnNameEdit(col)
												}
												className="w-full text-left text-sm font-semibold text-foreground hover:text-primary transition-colors"
											>
												{col}
											</button>
										)}
									</th>
								))}
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
													row[col] !== null &&
													row[col] !== undefined
														? String(row[col])
														: ""
												}
												onChange={(e) =>
													onCellChange(
														rowIndex,
														col,
														e.target.value
													)
												}
												placeholder="—"
												className="h-8 border-0 bg-transparent text-sm focus-visible:ring-1 focus-visible:ring-primary"
											/>
										</td>
									))}
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</Card>
		</div>
	);
}
