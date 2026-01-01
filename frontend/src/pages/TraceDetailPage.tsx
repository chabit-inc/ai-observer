import { useEffect, useState, useMemo } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { WaterfallView } from '@/components/traces/WaterfallView'
import { api } from '@/lib/api'
import { formatDuration, formatTimestamp, getStatusColor } from '@/lib/utils'
import { getServiceDisplayName } from '@/lib/metricMetadata'
import type { Span, TraceOverview } from '@/types/traces'
import { ArrowLeft } from 'lucide-react'
import { toast } from 'sonner'

export function TraceDetailPage() {
  const { traceId } = useParams<{ traceId: string }>()
  const navigate = useNavigate()
  const [spans, setSpans] = useState<Span[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!traceId) return

    const abortController = new AbortController()

    const fetchSpans = async () => {
      setLoading(true)
      setError(null)
      try {
        const data = await api.getTraceSpans(traceId, { signal: abortController.signal })
        if (data.spans && data.spans.length > 0) {
          setSpans(data.spans)
        } else {
          setError('No spans found for this trace')
        }
      } catch (err) {
        if (err instanceof Error && err.name === 'AbortError') {
          return
        }
        console.error('Failed to fetch trace spans:', err)
        setError('Failed to load trace')
        toast.error('Failed to load trace')
      } finally {
        if (!abortController.signal.aborted) {
          setLoading(false)
        }
      }
    }

    fetchSpans()

    return () => abortController.abort()
  }, [traceId])

  // Compute trace overview from spans
  const traceOverview: TraceOverview | null = useMemo(() => {
    if (spans.length === 0) return null

    // Find root span (no parent or parent not in this trace)
    const spanIds = new Set(spans.map(s => s.spanId))
    const rootSpan = spans.find(s => !s.parentSpanId || !spanIds.has(s.parentSpanId))

    // Calculate total duration from earliest start to latest end
    let minStart = Number.MAX_SAFE_INTEGER
    let maxEnd = Number.MIN_SAFE_INTEGER
    let hasError = false

    for (const span of spans) {
      const start = new Date(span.timestamp).getTime()
      const end = start + span.duration / 1_000_000
      if (start < minStart) minStart = start
      if (end > maxEnd) maxEnd = end
      if (span.statusCode === 'ERROR') hasError = true
    }

    const totalDuration = (maxEnd - minStart) * 1_000_000 // Convert back to nanoseconds

    return {
      traceId: traceId!,
      rootSpan: rootSpan?.spanName || 'Unknown',
      serviceName: rootSpan?.serviceName || spans[0]?.serviceName || 'Unknown',
      startTime: rootSpan?.timestamp || spans[0]?.timestamp || new Date().toISOString(),
      duration: totalDuration,
      spanCount: spans.length,
      status: hasError ? 'ERROR' : 'OK'
    }
  }, [spans, traceId])

  const handleBack = () => {
    navigate('/traces')
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center gap-4">
        <button
          onClick={handleBack}
          className="p-2 -ml-2 rounded-md hover:bg-accent transition-colors"
          aria-label="Back to traces"
        >
          <ArrowLeft className="h-5 w-5" />
        </button>
        <h1 className="text-2xl font-bold tracking-tight font-mono">
          Trace {traceId}
        </h1>
      </div>

      {/* Content */}
      {loading ? (
        <Card>
          <CardContent className="py-8">
            <div className="text-center text-muted-foreground">
              Loading trace...
            </div>
          </CardContent>
        </Card>
      ) : error ? (
        <Card>
          <CardContent className="py-8">
            <div className="text-center text-muted-foreground">
              {error}
            </div>
          </CardContent>
        </Card>
      ) : traceOverview ? (
        <>
          {/* Trace Summary */}
          <Card>
            <CardContent className="py-4">
              <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm">
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">Status:</span>
                  <Badge className={getStatusColor(traceOverview.status)}>
                    {traceOverview.status}
                  </Badge>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">Service:</span>
                  <span className="font-medium">{getServiceDisplayName(traceOverview.serviceName)}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">Duration:</span>
                  <span className="font-medium font-mono">{formatDuration(traceOverview.duration)}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">Spans:</span>
                  <span className="font-medium">{traceOverview.spanCount}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-muted-foreground">Started:</span>
                  <span className="font-medium">{formatTimestamp(traceOverview.startTime)}</span>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Waterfall View */}
          <Card>
            <CardContent className="py-4">
              <WaterfallView spans={spans} trace={traceOverview} />
            </CardContent>
          </Card>
        </>
      ) : null}
    </div>
  )
}
