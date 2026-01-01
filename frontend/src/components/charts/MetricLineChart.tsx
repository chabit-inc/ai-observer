import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'
import {
  CHART_COLORS,
  compactTooltipStyle,
  compactAxisStyle,
  type MetricChartProps,
} from './chartUtils'

export function MetricLineChart({
  data,
  series,
  colorMap,
  formatYAxis,
  tooltipFormatter,
  xAxisInterval,
  compact = false,
  showLegend = false,
}: MetricChartProps) {
  // Calculate default x-axis interval for compact mode
  const defaultInterval = compact
    ? Math.max(0, Math.floor(data.length / 4) - 1)
    : 0

  const interval = xAxisInterval ?? defaultInterval

  return (
    <ResponsiveContainer width="100%" height="100%">
      <LineChart
        data={data}
        margin={
          compact
            ? { top: 5, right: 25, left: 0, bottom: 0 }
            : { top: 5, right: 20, left: 0, bottom: 5 }
        }
      >
        <CartesianGrid
          strokeDasharray="3 3"
          stroke={compact ? 'hsl(var(--border))' : undefined}
        />
        <XAxis
          dataKey="time"
          tick={{ fontSize: compact ? 10 : 12 }}
          interval={interval}
          {...(compact ? compactAxisStyle : {})}
        />
        <YAxis
          tick={{ fontSize: compact ? 10 : 12 }}
          width={compact ? 50 : undefined}
          tickFormatter={formatYAxis}
          {...(compact ? compactAxisStyle : {})}
        />
        <Tooltip
          offset={compact ? 20 : undefined}
          formatter={tooltipFormatter}
          {...(compact ? compactTooltipStyle : {})}
        />
        {showLegend && (
          <Legend
            formatter={(value) => {
              const s = series.find((s) => s.key === value)
              return s?.label || value
            }}
          />
        )}
        {series.map((s) => (
          <Line
            key={s.key}
            type="monotone"
            dataKey={s.key}
            stroke={colorMap.get(s.key) || CHART_COLORS[0]}
            strokeWidth={2}
            dot={false}
            isAnimationActive={false}
          />
        ))}
      </LineChart>
    </ResponsiveContainer>
  )
}
