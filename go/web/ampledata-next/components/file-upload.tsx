"use client"

import type React from "react"

import { useState, useCallback } from "react"
import { Card } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Upload, FileText, FileJson, Table } from "lucide-react"
import type { DataRow } from "@/lib/types"

interface FileUploadProps {
  onFileUpload: (data: DataRow[], columns: string[], fileName: string) => void
}

export function FileUpload({ onFileUpload }: FileUploadProps) {
  const [isDragging, setIsDragging] = useState(false)

  const parseCSV = (text: string): { data: DataRow[]; columns: string[] } => {
    const lines = text.trim().split("\n")
    const columns = lines[0].split(",").map((col) => col.trim())
    const data: DataRow[] = lines.slice(1).map((line) => {
      const values = line.split(",").map((val) => val.trim())
      const row: DataRow = {}
      columns.forEach((col, index) => {
        row[col] = values[index] || null
      })
      return row
    })
    return { data, columns }
  }

  const parseJSON = (text: string): { data: DataRow[]; columns: string[] } => {
    const jsonData = JSON.parse(text)
    const data: DataRow[] = Array.isArray(jsonData) ? jsonData : [jsonData]
    const columns = data.length > 0 ? Object.keys(data[0]) : []
    return { data, columns }
  }

  const handleFile = useCallback(
    async (file: File) => {
      const text = await file.text()
      let parsedData: { data: DataRow[]; columns: string[] }

      if (file.name.endsWith(".csv")) {
        parsedData = parseCSV(text)
      } else if (file.name.endsWith(".json")) {
        parsedData = parseJSON(text)
      } else {
        // Try CSV first, then JSON
        try {
          parsedData = parseCSV(text)
        } catch {
          parsedData = parseJSON(text)
        }
      }

      onFileUpload(parsedData.data, parsedData.columns, file.name)
    },
    [onFileUpload],
  )

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setIsDragging(false)

      const file = e.dataTransfer.files[0]
      if (file) {
        handleFile(file)
      }
    },
    [handleFile],
  )

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setIsDragging(false)
  }, [])

  const handleFileInput = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const file = e.target.files?.[0]
      if (file) {
        handleFile(file)
      }
    },
    [handleFile],
  )

  return (
    <div className="flex min-h-[calc(100vh-12rem)] items-center justify-center">
      <Card className="w-full max-w-2xl border-2 border-dashed">
        <div
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
          className={`flex flex-col items-center justify-center gap-6 p-12 transition-colors ${
            isDragging ? "bg-accent" : ""
          }`}
        >
          <div className="rounded-full bg-primary/10 p-6">
            <Upload className="h-12 w-12 text-primary" />
          </div>

          <div className="text-center">
            <h2 className="text-2xl font-semibold tracking-tight text-foreground">Upload your data</h2>
            <p className="mt-2 text-balance text-muted-foreground">Drag and drop your file here, or click to browse</p>
          </div>

          <div className="flex flex-wrap items-center justify-center gap-4">
            <div className="flex items-center gap-2 rounded-lg bg-muted px-4 py-2">
              <FileText className="h-5 w-5 text-muted-foreground" />
              <span className="text-sm font-medium">CSV</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg bg-muted px-4 py-2">
              <FileJson className="h-5 w-5 text-muted-foreground" />
              <span className="text-sm font-medium">JSON</span>
            </div>
            <div className="flex items-center gap-2 rounded-lg bg-muted px-4 py-2">
              <Table className="h-5 w-5 text-muted-foreground" />
              <span className="text-sm font-medium">TXT</span>
            </div>
          </div>

          <label htmlFor="file-upload">
            <Button asChild size="lg">
              <span>
                <input
                  id="file-upload"
                  type="file"
                  className="sr-only"
                  accept=".csv,.json,.txt"
                  onChange={handleFileInput}
                />
                Choose File
              </span>
            </Button>
          </label>

          <p className="text-xs text-muted-foreground">Maximum file size: 10MB</p>
        </div>
      </Card>
    </div>
  )
}
