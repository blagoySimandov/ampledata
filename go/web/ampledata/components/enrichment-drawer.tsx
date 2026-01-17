"use client"

import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Progress } from "@/components/ui/progress"
import { CheckCircle2, Loader2 } from "lucide-react"

interface EnrichmentDrawerProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  progress: number
  isEnriching: boolean
  totalRows: number
}

export function EnrichmentDrawer({ open, onOpenChange, progress, isEnriching, totalRows }: EnrichmentDrawerProps) {
  const rowsProcessed = Math.floor((progress / 100) * totalRows)

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent className="sm:max-w-md">
        <SheetHeader>
          <SheetTitle>Data Enrichment</SheetTitle>
          <SheetDescription>{isEnriching ? "Enriching your data..." : "Enrichment complete"}</SheetDescription>
        </SheetHeader>

        <div className="mt-8 space-y-6">
          <div className="space-y-3">
            <div className="flex items-center justify-between text-sm">
              <span className="text-muted-foreground">Progress</span>
              <span className="font-semibold text-foreground">{Math.round(progress)}%</span>
            </div>
            <Progress value={progress} className="h-2" />
            <p className="text-xs text-muted-foreground">
              {rowsProcessed} of {totalRows} rows processed
            </p>
          </div>

          <div className="rounded-lg border bg-card p-4">
            {isEnriching ? (
              <div className="flex items-start gap-3">
                <Loader2 className="mt-0.5 h-5 w-5 animate-spin text-primary" />
                <div className="space-y-1">
                  <p className="text-sm font-medium text-foreground">Processing data</p>
                  <p className="text-xs text-muted-foreground">
                    This may take a few moments depending on the size of your dataset
                  </p>
                </div>
              </div>
            ) : (
              <div className="flex items-start gap-3">
                <CheckCircle2 className="mt-0.5 h-5 w-5 text-green-600 dark:text-green-500" />
                <div className="space-y-1">
                  <p className="text-sm font-medium text-foreground">Enrichment complete</p>
                  <p className="text-xs text-muted-foreground">All {totalRows} rows have been successfully enriched</p>
                </div>
              </div>
            )}
          </div>

          <div className="space-y-3 rounded-lg border bg-muted/50 p-4">
            <h4 className="text-sm font-semibold text-foreground">Status</h4>
            <div className="space-y-2 text-xs">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Total Rows</span>
                <span className="font-medium text-foreground">{totalRows}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Processed</span>
                <span className="font-medium text-foreground">{rowsProcessed}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Remaining</span>
                <span className="font-medium text-foreground">{totalRows - rowsProcessed}</span>
              </div>
            </div>
          </div>
        </div>
      </SheetContent>
    </Sheet>
  )
}
