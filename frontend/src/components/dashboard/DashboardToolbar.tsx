import { useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { Plus, Settings, X } from 'lucide-react'
import { useIsMobile } from '@/hooks/use-mobile'
import { useDashboardStore } from '@/stores/dashboardStore'
import { TIMEFRAME_OPTIONS } from '@/types/dashboard'

export function DashboardToolbar() {
  const isMobile = useIsMobile()
  const {
    isEditMode,
    setEditMode,
    timeframe,
    setTimeframe,
    setAddPanelOpen,
  } = useDashboardStore()

  // Auto-exit edit mode when switching to mobile view
  useEffect(() => {
    if (isMobile && isEditMode) {
      setEditMode(false)
    }
  }, [isMobile, isEditMode, setEditMode])

  const handleTimeframeChange = (value: string) => {
    const newTimeframe = TIMEFRAME_OPTIONS.find((t) => t.value === value)
    if (newTimeframe) {
      setTimeframe(newTimeframe)
    }
  }

  return (
    <div className="flex items-center gap-3">
      {/* Timeframe Selector */}
      <div className="flex items-center gap-2">
        <label className="text-sm text-muted-foreground whitespace-nowrap hidden sm:inline">Timeframe</label>
        <Select
          value={timeframe.value}
          onChange={(e) => handleTimeframeChange(e.target.value)}
          className="w-32 sm:w-40"
        >
          {TIMEFRAME_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </Select>
      </div>

      {/* Edit Mode Actions - hidden on mobile */}
      {!isMobile && (
        isEditMode ? (
          <>
            <Button
              variant="outline"
              onClick={() => setAddPanelOpen(true)}
            >
              <Plus className="h-4 w-4 mr-1" />
              <span className="hidden sm:inline">Add Widget</span>
              <span className="sm:hidden">Add</span>
            </Button>
            <Button
              variant="secondary"
              onClick={() => setEditMode(false)}
            >
              <X className="h-4 w-4 mr-1" />
              Done
            </Button>
          </>
        ) : (
          <Button
            variant="outline"
            onClick={() => setEditMode(true)}
          >
            <Settings className="h-4 w-4 mr-1" />
            Edit
          </Button>
        )
      )}
    </div>
  )
}
