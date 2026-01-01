import { useEffect, useState, useRef, useMemo, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Select } from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Layers, BarChart3 } from 'lucide-react'
import { api } from '@/lib/api'
import type { TimeSeries } from '@/types/metrics'
import { MetricBarChart, CHART_COLORS } from '@/components/charts'
import { useTelemetryStore } from '@/stores/telemetryStore'
import {
  getMetricMetadata,
  getSeriesLabel,
  formatMetricValue,
  getSourceDisplayName,
  getServiceDisplayName,
  METRIC_CATALOG,
  type SourceTool,
} from '@/lib/metricMetadata'
import { toast } from 'sonner'
import { TIMEFRAME_OPTIONS } from '@/types/dashboard'
import { formatIntervalSeconds } from '@/lib/utils'
import { getLocalStorageValue } from '@/hooks/useLocalStorage'

const DEFAULT_TIMEFRAME = '15m'
const TIMEFRAME_STORAGE_KEY = 'ai-observer-metrics-timeframe'

export function MetricsPage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const [metricNames, setMetricNames] = useState<string[]>([])
  const [selectedMetric, setSelectedMetric] = useState<string>('')
  const [series, setSeries] = useState<TimeSeries[]>([])
  const [services, setServices] = useState<string[]>([])
  const [selectedService, setSelectedService] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date())
  const [stacked, setStacked] = useState(true)

  // Get timeframe from URL (for shareable links), then localStorage, then default
  const timeframeParam = searchParams.get('timeframe')
    || getLocalStorageValue(TIMEFRAME_STORAGE_KEY, DEFAULT_TIMEFRAME)
  const selectedTimeframe = TIMEFRAME_OPTIONS.find(t => t.value === timeframeParam)
    || TIMEFRAME_OPTIONS.find(t => t.value === DEFAULT_TIMEFRAME)!

  const handleTimeframeChange = (value: string) => {
    // Persist to localStorage
    try {
      localStorage.setItem(TIMEFRAME_STORAGE_KEY, JSON.stringify(value))
    } catch {
      // Ignore storage errors
    }
    // Update URL params
    const newParams = new URLSearchParams(searchParams)
    if (value === DEFAULT_TIMEFRAME) {
      newParams.delete('timeframe')
    } else {
      newParams.set('timeframe', value)
    }
    setSearchParams(newParams)
  }

  // Real-time metrics from WebSocket
  const recentMetrics = useTelemetryStore((state) => state.recentMetrics)
  const prevMetricsCountRef = useRef(0)

  // Map service name to source tool
  const getSourceFromService = useCallback((service: string): SourceTool | null => {
    const normalized = service.toLowerCase().replace(/-/g, '_')
    if (normalized.includes('claude')) return 'claude_code'
    if (normalized.includes('gemini')) return 'gemini_cli'
    if (normalized.includes('codex')) return 'codex_cli_rs'
    return null
  }, [])

  // Fetch services on mount
  useEffect(() => {
    const fetchServices = async () => {
      try {
        const servicesData = await api.getServices()
        setServices(servicesData.services ?? [])
      } catch (err) {
        console.error('Failed to fetch services:', err)
        toast.error('Failed to fetch services')
      }
    }
    fetchServices()
  }, [])

  // Fetch metric names when service selection changes
  useEffect(() => {
    const abortController = new AbortController()

    const fetchMetricNames = async () => {
      try {
        const namesData = await api.getMetricNames(selectedService || undefined, { signal: abortController.signal })
        const names = namesData.names ?? []
        setMetricNames(names)
      } catch (err) {
        if (err instanceof Error && err.name === 'AbortError') {
          return // Ignore abort errors
        }
        console.error('Failed to fetch metric names:', err)
        toast.error('Failed to fetch metric names')
      }
    }
    fetchMetricNames()

    return () => abortController.abort()
  }, [selectedService])

  // Reset selected metric when service changes and current metric is not valid
  useEffect(() => {
    if (!selectedService) return // No filtering when no service selected

    const filterSource = getSourceFromService(selectedService)
    if (!filterSource) return

    // Check if current metric belongs to the selected service's source
    if (selectedMetric) {
      const meta = getMetricMetadata(selectedMetric)
      if (meta.source !== filterSource && meta.source !== 'unknown') {
        setSelectedMetric('')
      }
    }
  }, [selectedService, selectedMetric, getSourceFromService])

  // Auto-refresh at the X-axis tick interval rate
  // Recharts interval=N skips N ticks, so actual spacing is (N+1) data points
  useEffect(() => {
    const refreshMs = selectedTimeframe.intervalSeconds * 1000 * (selectedTimeframe.tickInterval + 1)
    const interval = setInterval(() => {
      setLastUpdate(new Date())
    }, refreshMs)
    return () => clearInterval(interval)
  }, [selectedTimeframe.intervalSeconds, selectedTimeframe.tickInterval])

  // Also refresh when new metrics arrive via WebSocket (debounced)
  useEffect(() => {
    if (recentMetrics.length > prevMetricsCountRef.current) {
      const timer = setTimeout(() => {
        setLastUpdate(new Date())
      }, 500)
      prevMetricsCountRef.current = recentMetrics.length
      return () => clearTimeout(timer)
    }
  }, [recentMetrics.length])

  useEffect(() => {
    if (!selectedMetric) {
      setSeries([])
      return
    }

    const abortController = new AbortController()

    const fetchSeries = async () => {
      // Only show loading on initial load, not on refreshes
      const isInitialLoad = series.length === 0
      if (isInitialLoad) {
        setLoading(true)
      }
      try {
        const now = new Date()
        const from = new Date(now.getTime() - selectedTimeframe.durationSeconds * 1000)

        const data = await api.getMetricSeries({
          name: selectedMetric,
          service: selectedService || undefined,
          from: from.toISOString(),
          to: now.toISOString(),
          intervalSeconds: selectedTimeframe.intervalSeconds,
        }, { signal: abortController.signal })
        setSeries(data.series ?? [])
      } catch (err) {
        if (err instanceof Error && err.name === 'AbortError') {
          return // Ignore abort errors
        }
        console.error('Failed to fetch metric series:', err)
        toast.error('Failed to fetch metric series')
      } finally {
        if (!abortController.signal.aborted && isInitialLoad) {
          setLoading(false)
        }
      }
    }
    fetchSeries()

    return () => abortController.abort()
  }, [selectedMetric, selectedService, lastUpdate, selectedTimeframe])

  // Get metadata for selected metric
  const metadata = useMemo(
    () => (selectedMetric ? getMetricMetadata(selectedMetric) : null),
    [selectedMetric]
  )

  // Group metric names by source - filter by service if selected
  const groupedMetrics = useMemo(() => {
    const groups: Record<SourceTool | 'other', string[]> = {
      claude_code: [],
      gemini_cli: [],
      codex_cli_rs: [],
      unknown: [],
      other: [],
    }

    // Determine which source to filter by based on selected service
    const filterSource = selectedService ? getSourceFromService(selectedService) : null

    // Start with catalog metrics (filtered by source if service selected)
    const allMetricNames = new Set<string>()
    for (const m of METRIC_CATALOG) {
      if (!filterSource || m.source === filterSource) {
        allMetricNames.add(m.name)
      }
    }

    // Add metrics from the database (already filtered by service via API)
    for (const name of metricNames) {
      allMetricNames.add(name)
    }

    // Group by source
    for (const name of allMetricNames) {
      const meta = getMetricMetadata(name)
      if (meta.source === 'unknown') {
        groups.other.push(name)
      } else {
        groups[meta.source].push(name)
      }
    }

    // Sort each group alphabetically
    for (const key of Object.keys(groups) as (SourceTool | 'other')[]) {
      groups[key].sort()
    }

    return groups
  }, [metricNames, selectedService, getSourceFromService])

  // Helper to get series key from labels (prefer type if available, else service)
  const getSeriesKey = useCallback((labels?: Record<string, string>) => {
    if (labels?.type) {
      return labels.service ? `${labels.service}:${labels.type}` : labels.type
    }
    return labels?.service || 'value'
  }, [])

  // Helper to get human-readable series label
  const getDisplayLabel = useCallback((labels?: Record<string, string>) => {
    if (!metadata || !labels) return labels?.service || 'value'
    return getSeriesLabel(labels, metadata)
  }, [metadata])

  // Format Y-axis values
  const formatYAxis = useCallback((value: number) => {
    if (!metadata) return value.toString()
    return formatMetricValue(value, metadata.unit)
  }, [metadata])

  // Custom tooltip formatter
  const tooltipFormatter = useCallback((value: number | undefined, name: string | undefined): [string, string] => {
    if (value === undefined) return ['—', name || 'value']
    // Find the series to get its labels
    const seriesItem = series.find(s => getSeriesKey(s.labels) === name)
    const displayLabel = seriesItem ? getDisplayLabel(seriesItem.labels) : (name || 'value')
    const formattedValue = metadata ? formatMetricValue(value, metadata.unit) : value.toString()
    return [formattedValue, displayLabel]
  }, [metadata, series, getSeriesKey, getDisplayLabel])

  // Generate chart data directly from backend data (backend fills missing buckets)
  const sortedData = useMemo(() => {
    if (series.length === 0) return []

    // Collect all unique timestamps from backend
    const allTimestamps = new Set<number>()
    for (const s of series) {
      for (const [timestamp] of s.datapoints) {
        allTimestamps.add(timestamp)
      }
    }

    // Sort timestamps
    const sortedTimestamps = Array.from(allTimestamps).sort((a, b) => a - b)

    // Get all series keys
    const seriesKeys = series.map((s) => getSeriesKey(s.labels))

    // Build lookup map: seriesKey -> timestamp -> value
    const dataMap = new Map<string, Map<number, number>>()
    for (const s of series) {
      const key = getSeriesKey(s.labels)
      const timestampMap = new Map<number, number>()
      for (const [timestamp, value] of s.datapoints) {
        timestampMap.set(timestamp, value)
      }
      dataMap.set(key, timestampMap)
    }

    // Create chart data from backend timestamps
    return sortedTimestamps.map((timestamp) => {
      const point: Record<string, number | string> = {
        timestamp,
        time: new Date(timestamp).toLocaleTimeString(),
      }
      for (const key of seriesKeys) {
        const value = dataMap.get(key)?.get(timestamp)
        point[key] = value ?? 0
      }
      return point
    })
  }, [series, getSeriesKey])

  // Create deterministic color map based on sorted series keys (memoized)
  const colorMap = useMemo(() => {
    const keys = series.map((s) => getSeriesKey(s.labels)).sort()
    const map = new Map<string, string>()
    keys.forEach((key, i) => {
      map.set(key, CHART_COLORS[i % CHART_COLORS.length])
    })
    return map
  }, [series, getSeriesKey])

  // Convert series to ChartSeries format for the chart component
  const chartSeries = useMemo(() => {
    return series.map((s) => ({
      key: getSeriesKey(s.labels),
      label: getDisplayLabel(s.labels),
    }))
  }, [series, getSeriesKey, getDisplayLabel])

  // Use timeframe-specific tick interval for X-axis labels
  const xAxisTickInterval = selectedTimeframe.tickInterval

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Metrics</h1>
        <p className="text-muted-foreground">
          View and analyze metric time series data
        </p>
      </div>

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex gap-4">
            <div className="flex-1">
              <label className="text-sm font-medium mb-2 block">Service</label>
              <Select
                value={selectedService}
                onChange={(e) => setSelectedService(e.target.value)}
              >
                <option value="">All Services</option>
                {services.map((s) => (
                  <option key={s} value={s}>
                    {getServiceDisplayName(s)}
                  </option>
                ))}
              </Select>
            </div>
            <div className="flex-1">
              <label className="text-sm font-medium mb-2 block">Metric</label>
              <Select
                value={selectedMetric}
                onChange={(e) => setSelectedMetric(e.target.value)}
              >
                <option value="">Select a metric</option>
                {groupedMetrics.claude_code.length > 0 && (
                  <optgroup label={getSourceDisplayName('claude_code')}>
                    {groupedMetrics.claude_code.map((name) => (
                      <option key={name} value={name}>{getMetricMetadata(name).displayName}</option>
                    ))}
                  </optgroup>
                )}
                {groupedMetrics.gemini_cli.length > 0 && (
                  <optgroup label={getSourceDisplayName('gemini_cli')}>
                    {groupedMetrics.gemini_cli.map((name) => (
                      <option key={name} value={name}>{getMetricMetadata(name).displayName}</option>
                    ))}
                  </optgroup>
                )}
                {groupedMetrics.codex_cli_rs.length > 0 && (
                  <optgroup label={getSourceDisplayName('codex_cli_rs')}>
                    {groupedMetrics.codex_cli_rs.map((name) => (
                      <option key={name} value={name}>{getMetricMetadata(name).displayName}</option>
                    ))}
                  </optgroup>
                )}
                {groupedMetrics.other.length > 0 && (
                  <optgroup label="Other">
                    {groupedMetrics.other.map((name) => (
                      <option key={name} value={name}>{name}</option>
                    ))}
                  </optgroup>
                )}
              </Select>
            </div>
            <div className="flex-1">
              <label className="text-sm font-medium mb-2 block">Timeframe</label>
              <Select
                value={selectedTimeframe.value}
                onChange={(e) => handleTimeframeChange(e.target.value)}
              >
                {TIMEFRAME_OPTIONS.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </Select>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Chart */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2">
                <CardTitle>{metadata?.displayName || selectedMetric || 'Select a metric'}</CardTitle>
                {metadata && (
                  <Badge variant="outline" className="text-xs">
                    {metadata.unit.displayUnit}
                  </Badge>
                )}
              </div>
              <CardDescription>
                {metadata?.description || `${selectedTimeframe.label}, ${formatIntervalSeconds(selectedTimeframe.intervalSeconds)} intervals`}
              </CardDescription>
              {metadata && (
                <p className="text-xs text-muted-foreground mt-1">
                  {selectedTimeframe.label}, {formatIntervalSeconds(selectedTimeframe.intervalSeconds)} intervals
                </p>
              )}
            </div>
            <div className="flex items-center gap-2">
              {recentMetrics.length > 0 && (
                <Badge variant="secondary" className="text-xs">
                  {recentMetrics.length} real-time updates
                </Badge>
              )}
              <div className="flex border rounded-md">
                <Button
                  variant={stacked ? "secondary" : "ghost"}
                  size="sm"
                  className="h-8 px-2 rounded-r-none"
                  onClick={() => setStacked(true)}
                  title="Stacked bars"
                >
                  <Layers className="h-4 w-4" />
                </Button>
                <Button
                  variant={!stacked ? "secondary" : "ghost"}
                  size="sm"
                  className="h-8 px-2 rounded-l-none"
                  onClick={() => setStacked(false)}
                  title="Grouped bars"
                >
                  <BarChart3 className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {loading ? (
            <div className="h-80 flex items-center justify-center text-muted-foreground">
              Loading...
            </div>
          ) : sortedData.length === 0 ? (
            <div className="h-80 flex items-center justify-center text-muted-foreground">
              {selectedMetric
                ? 'No data available for this metric'
                : 'Select a metric to view data'}
            </div>
          ) : (
            <div className="h-80">
              <MetricBarChart
                data={sortedData}
                series={chartSeries}
                colorMap={colorMap}
                formatYAxis={formatYAxis}
                tooltipFormatter={tooltipFormatter}
                xAxisInterval={xAxisTickInterval}
                showLegend={true}
                stacked={stacked}
              />
            </div>
          )}
        </CardContent>
      </Card>

      {/* Metric Info */}
      {metricNames.length === 0 && (
        <Card>
          <CardContent className="pt-6">
            <p className="text-muted-foreground text-center">
              No metrics data available yet. Configure your AI coding tool to send metrics telemetry.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
