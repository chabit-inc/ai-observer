import {
  BarChart,
  Bar,
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
  standardTooltipStyle,
  compactAxisStyle,
  type MetricChartProps,
} from './chartUtils'

export function MetricBarChart({
  data,
  series,
  colorMap,
  formatYAxis,
  tooltipFormatter,
  xAxisInterval,
  compact = false,
  showLegend = false,
  stacked = true,
}: MetricChartProps) {
  // Calculate default x-axis interval for compact mode
  const defaultInterval = compact
    ? Math.max(0, Math.floor(data.length / 4) - 1)
    : 0

  const interval = xAxisInterval ?? defaultInterval

  return (
    <ResponsiveContainer width="100%" height="100%">
      <BarChart
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
          {...(compact ? compactTooltipStyle : standardTooltipStyle)}
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
          <Bar
            key={s.key}
            dataKey={s.key}
            fill={colorMap.get(s.key) || CHART_COLORS[0]}
            isAnimationActive={false}
            barSize={compact ? 8 : 12}
            stackId={stacked ? 'stack' : undefined}
          />
        ))}
      </BarChart>
    </ResponsiveContainer>
  )
}
