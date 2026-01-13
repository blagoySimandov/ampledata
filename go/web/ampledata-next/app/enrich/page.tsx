"use client"

import { useState } from "react"
import { FileUpload } from "@/components/file-upload"
import { DataGrid } from "@/components/data-grid"
import { EnrichmentDrawer } from "@/components/enrichment-drawer"
import { Button } from "@/components/ui/button"
import { ThemeToggle } from "@/components/theme-toggle"
import { useUser } from "@/hooks"
import type { DataRow } from "@/lib/types"
import { Download, Table } from "lucide-react"

export default function DataEnrichmentPage() {
  const user = useUser()
  const [data, setData] = useState<DataRow[]>([])
  const [columns, setColumns] = useState<string[]>([])
  const [fileName, setFileName] = useState<string>("")
  const [isEnriching, setIsEnriching] = useState(false)
  const [enrichmentProgress, setEnrichmentProgress] = useState(0)
  const [drawerOpen, setDrawerOpen] = useState(false)

  if (!user) {
    return (
      <div className="flex items-center justify-center min-h-[50vh]">
        <p className="text-lg">Loading...</p>
      </div>
    )
  }

  const handleFileUpload = (uploadedData: DataRow[], uploadedColumns: string[], name: string) => {
    setData(uploadedData)
    setColumns(uploadedColumns)
    setFileName(name)
  }

  const handleAddColumn = (columnName: string) => {
    setColumns([...columns, columnName])
    const updatedData = data.map((row) => ({ ...row, [columnName]: null }))
    setData(updatedData)
  }

  const handleColumnNameChange = (oldName: string, newName: string) => {
    const updatedColumns = columns.map((col) => (col === oldName ? newName : col))
    setColumns(updatedColumns)

    const updatedData = data.map((row) => {
      const newRow = { ...row }
      newRow[newName] = row[oldName]
      delete newRow[oldName]
      return newRow
    })
    setData(updatedData)
  }

  const handleCellChange = (rowIndex: number, columnName: string, value: string) => {
    const updatedData = [...data]
    updatedData[rowIndex] = { ...updatedData[rowIndex], [columnName]: value }
    setData(updatedData)
  }

  const handleAddRow = () => {
    const newRow: DataRow = {}
    columns.forEach((col) => {
      newRow[col] = null
    })
    setData([...data, newRow])
  }

  const handleEnrich = async (columnName: string, dataType: string) => {
    setIsEnriching(true)
    setEnrichmentProgress(0)
    setDrawerOpen(true)

    // Simulate enrichment process
    const totalRows = data.length
    const enrichedData = [...data]

    for (let i = 0; i < totalRows; i++) {
      await new Promise((resolve) => setTimeout(resolve, 100))

      // Simulate enrichment based on data type
      let enrichedValue: string | number | boolean
      switch (dataType) {
        case "email":
          enrichedValue = `enriched${i}@example.com`
          break
        case "phone":
          enrichedValue = `+1-555-${String(Math.floor(Math.random() * 9000) + 1000)}`
          break
        case "company":
          enrichedValue = ["Acme Corp", "TechStart Inc", "Global Industries", "Innovation Labs"][
            Math.floor(Math.random() * 4)
          ]
          break
        case "location":
          enrichedValue = ["New York, NY", "San Francisco, CA", "Austin, TX", "Seattle, WA"][
            Math.floor(Math.random() * 4)
          ]
          break
        case "number":
          enrichedValue = Math.floor(Math.random() * 1000)
          break
        case "boolean":
          enrichedValue = Math.random() > 0.5
          break
        default:
          enrichedValue = `Enriched ${dataType} ${i + 1}`
      }

      enrichedData[i] = { ...enrichedData[i], [columnName]: enrichedValue }
      setEnrichmentProgress(((i + 1) / totalRows) * 100)
    }

    setData(enrichedData)
    setIsEnriching(false)
  }

  const handleExport = () => {
    // Convert data to CSV
    const csv = [
      columns.join(","),
      ...data.map((row) =>
        columns
          .map((col) => {
            const value = row[col]
            if (value === null || value === undefined) return ""
            return typeof value === "string" && value.includes(",") ? `"${value}"` : value
          })
          .join(","),
      ),
    ].join("\n")

    // Download CSV
    const blob = new Blob([csv], { type: "text/csv" })
    const url = URL.createObjectURL(blob)
    const a = document.createElement("a")
    a.href = url
    a.download = `enriched-${fileName || "data"}.csv`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card">
        <div className="container mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary">
                <Table className="h-5 w-5 text-primary-foreground" />
              </div>
              <div>
                <h1 className="text-xl font-semibold tracking-tight text-foreground sm:text-2xl">Data Enrichment</h1>
                <p className="text-sm text-muted-foreground">Upload, enrich, and export your data</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <ThemeToggle />
              {data.length > 0 && (
                <Button onClick={handleExport} size="sm" className="gap-2">
                  <Download className="h-4 w-4" />
                  <span className="hidden sm:inline">Export</span>
                </Button>
              )}
            </div>
          </div>
        </div>
      </header>

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
  )
}
