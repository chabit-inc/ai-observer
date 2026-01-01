import { useMemo, useCallback } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { BarChart as BarChartIcon } from 'lucide-react'
import type { WidgetConfig, TimeframeOption } from '@/types/dashboard'
import { MetricBarChart, CHART_COLORS } from '@/components/charts'
import { useMetricData } from '@/contexts/MetricDataContext'
import {
  getMetricMetadata,
  getSeriesLabel,
  formatMetricValue,
  getSourceDisplayName,
  getServiceDisplayName,
} from '@/lib/metricMetadata'

interface MetricChartWidgetProps {
  widgetId: string
  title: string
  config: WidgetConfig
  timeframe: TimeframeOption
  fromTime: Date
  toTime: Date
}

// Format timestamp for X-axis based on timeframe duration
const formatTickLabel = (timestamp: number, durationSeconds: number): string => {
  const date = new Date(timestamp)
  const DAY_SECONDS = 24 * 60 * 60

  if (durationSeconds <= DAY_SECONDS) {
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } else if (durationSeconds <= 7 * DAY_SECONDS) {
    return (
      date.toLocaleDateString([], { month: 'short', day: 'numeric' }) +
      ' ' +
      date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    )
  } else {
    return date.toLocaleDateString([], { month: 'short', day: 'numeric' })
  }
}

export function MetricChartWidget({
  widgetId,
  title,
  config,
  timeframe,
}: MetricChartWidgetProps) {
  // Get data from context (batched fetch)
  const { series, loading, error } = useMetricData(widgetId)

  // Get metadata for the configured metric
  const metadata = useMemo(
    () => (config.metricName ? getMetricMetadata(config.metricName) : null),
    [config.metricName]
  )

  // Helper to get series key from labels (used for data keys in chart)
  const getSeriesKey = useCallback(
    (labels?: Record<string, string>) => {
      if (!labels) return 'value'
      const breakdownKey = config.breakdownAttribute || 'type'
      if (labels[breakdownKey]) {
        return labels.service
          ? `${labels.service}:${labels[breakdownKey]}`
          : labels[breakdownKey]
      }
      if (labels.type) {
        return labels.service ? `${labels.service}:${labels.type}` : labels.type
      }
      return labels.service || 'value'
    },
    [config.breakdownAttribute]
  )

  // Helper to get human-readable series label for display
  const getDisplayLabel = useCallback(
    (labels?: Record<string, string>) => {
      if (!metadata || !labels) return labels?.service || 'value'
      return getSeriesLabel(labels, metadata, config.breakdownAttribute)
    },
    [metadata, config.breakdownAttribute]
  )

  // Format Y-axis values using metadata
  const formatYAxis = useCallback(
    (value: number) => {
      if (!metadata) return value.toString()
      return formatMetricValue(value, metadata.unit)
    },
    [metadata]
  )

  // Generate chart data directly from backend data (backend fills missing buckets)
  const chartData = useMemo(() => {
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

    // Build lookup map: seriesKey -> timestamp -> value
    const seriesKeys = series.map((s) => getSeriesKey(s.labels))
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
        time: formatTickLabel(timestamp, timeframe.durationSeconds),
      }
      for (const key of seriesKeys) {
        const value = dataMap.get(key)?.get(timestamp)
        point[key] = value ?? 0
      }
      return point
    })
  }, [series, timeframe.durationSeconds, getSeriesKey])

  // Color map for series
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

  // Map series keys to display labels for tooltip
  const keyToLabelMap = useMemo(() => {
    const map = new Map<string, string>()
    for (const s of series) {
      const key = getSeriesKey(s.labels)
      const label = getDisplayLabel(s.labels)
      map.set(key, label)
    }
    return map
  }, [series, getSeriesKey, getDisplayLabel])

  // Custom tooltip formatter
  const tooltipFormatter = useCallback(
    (value: number | undefined, name: string | undefined): [string, string] => {
      if (value === undefined) return ['—', name || 'value']
      const displayLabel = name ? keyToLabelMap.get(name) || name : 'value'
      const formattedValue = metadata
        ? formatMetricValue(value, metadata.unit)
        : value.toString()
      return [formattedValue, displayLabel]
    },
    [metadata, keyToLabelMap]
  )

  // Get service name from config, series labels, or fall back to source display name
  const serviceName = useMemo(() => {
    if (config.service) return getServiceDisplayName(config.service)
    for (const s of series) {
      if (s.labels?.service) return getServiceDisplayName(s.labels.service)
    }
    // Fall back to source display name from metric metadata
    if (metadata?.source && metadata.source !== 'unknown') {
      return getSourceDisplayName(metadata.source)
    }
    return null
  }, [config.service, series, metadata])

  return (
    <Card className="border-0 shadow-none h-full flex flex-col">
      <CardHeader className="flex flex-row items-start justify-between space-y-0 p-3 pb-0">
        <div className="flex flex-col min-w-0 flex-1">
          <span className="text-xs text-muted-foreground truncate h-4">
            {serviceName || '\u00A0'}
          </span>
          <CardTitle className="text-base font-medium truncate">
            {title}
          </CardTitle>
        </div>
        <BarChartIcon className="h-4 w-4 text-muted-foreground shrink-0" />
      </CardHeader>
      <CardContent className="flex-1 p-3 pt-2">
        {loading ? (
          <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">
            Loading...
          </div>
        ) : error ? (
          <div className="h-32 flex items-center justify-center text-destructive text-sm text-center px-2">
            {error}
          </div>
        ) : chartData.length === 0 || !config.metricName ? (
          <div className="h-32 flex items-center justify-center text-muted-foreground text-sm">
            {config.metricName ? 'No data' : 'Not configured'}
          </div>
        ) : (
          <div className="h-32">
            <MetricBarChart
              data={chartData}
              series={chartSeries}
              colorMap={colorMap}
              formatYAxis={formatYAxis}
              tooltipFormatter={tooltipFormatter}
              compact={true}
              stacked={config.chartStacked ?? true}
            />
          </div>
        )}
      </CardContent>
    </Card>
  )
}
