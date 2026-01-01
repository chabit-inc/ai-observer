export interface DocSubsection {
  id: string
  title: string
}

export interface DocTableRow {
  cells: string[]
}

export interface DocTable {
  headers: string[]
  rows: DocTableRow[]
}

export interface DocContent {
  type: 'paragraph' | 'heading' | 'list' | 'note' | 'table'
  text?: string
  items?: string[]
  level?: 2 | 3
  table?: DocTable
}

export interface DocSection {
  id: string
  title: string
  subsections?: DocSubsection[]
  content: DocContent[]
}

export const DOCUMENTATION_SECTIONS: DocSection[] = [
  {
    id: 'overview',
    title: 'Overview',
    content: [
      {
        type: 'paragraph',
        text: 'AI Observer is a unified observability dashboard for monitoring AI coding assistants. It collects and displays telemetry data from Claude Code, Gemini CLI, and OpenAI Codex CLI, giving you real-time visibility into token usage, costs, API performance, and session activity.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Navigation',
      },
      {
        type: 'paragraph',
        text: 'The sidebar on the left provides access to all main sections:',
      },
      {
        type: 'list',
        items: [
          'Dashboards - Customizable overview with widgets showing key metrics',
          'Metrics - Time-series charts for detailed metric analysis',
          'Logs - Searchable log records with severity filtering',
          'Traces - Distributed trace visualization with waterfall view',
          'Documentation - This help page',
        ],
      },
      {
        type: 'heading',
        level: 3,
        text: 'Real-time Updates',
      },
      {
        type: 'paragraph',
        text: 'AI Observer receives telemetry via WebSocket, so data updates automatically as your AI tools send metrics, logs, and traces. The connection status indicator in the header shows whether live updates are active.',
      },
    ],
  },
  {
    id: 'dashboard',
    title: 'Dashboard',
    subsections: [
      { id: 'dashboard-creating', title: 'Creating Dashboards' },
      { id: 'dashboard-managing', title: 'Managing Dashboards' },
      { id: 'dashboard-widgets', title: 'Widget Types' },
      { id: 'dashboard-editing', title: 'Edit Mode' },
    ],
    content: [
      {
        type: 'paragraph',
        text: 'Dashboards provide a customizable overview of your AI tool telemetry. You can create multiple dashboards with different widget configurations to monitor various aspects of your AI coding assistants.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Creating Dashboards',
      },
      {
        type: 'paragraph',
        text: 'To create a new dashboard, click the "New Dashboard" button in the sidebar under the Dashboards section. A new dashboard will be created with a default name that you can customize.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Managing Dashboards',
      },
      {
        type: 'paragraph',
        text: 'Each dashboard in the sidebar has a context menu (three dots icon) with the following options:',
      },
      {
        type: 'list',
        items: [
          'Rename - Change the dashboard name',
          'Set as Default - Mark this dashboard as the default, shown when you visit the root URL',
          'Delete - Remove the dashboard permanently',
        ],
      },
      {
        type: 'paragraph',
        text: 'The default dashboard is marked with a star icon in the sidebar.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Widget Types',
      },
      {
        type: 'paragraph',
        text: 'When adding widgets, you can choose from two categories:',
      },
      {
        type: 'note',
        text: 'Built-in Widgets',
      },
      {
        type: 'list',
        items: [
          'Total Traces - Count of trace spans received',
          'Metrics Count - Number of metric data points',
          'Log Count - Total log records',
          'Error Rate - Percentage of spans with error status',
          'Active Services - List of services sending telemetry',
          'Recent Traces - Table of the most recent traces with details',
        ],
      },
      {
        type: 'note',
        text: 'Metric Widgets',
      },
      {
        type: 'list',
        items: [
          'Metric Value - Display a single metric value (e.g., total tokens, session count)',
          'Metric Chart - Time-series bar chart with optional breakdown by attribute',
        ],
      },
      {
        type: 'paragraph',
        text: 'Metric widgets can be configured to show data from specific services and metrics. For chart widgets, you can optionally select a breakdown attribute to see the metric split by different values (e.g., token usage by type).',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Edit Mode',
      },
      {
        type: 'paragraph',
        text: 'Click the "Edit" button in the dashboard toolbar to enter edit mode. In edit mode, you can:',
      },
      {
        type: 'list',
        items: [
          'Add widgets using the "Add Widget" button',
          'Remove widgets by clicking the X button on each widget',
          'Rearrange widgets by dragging them to new positions',
          'Swap widget positions by dropping one widget onto another',
        ],
      },
      {
        type: 'paragraph',
        text: 'The dashboard uses a 4-column grid layout. Widgets occupy different column widths: stat widgets take 1 column, charts take 2 columns, and the recent activity table spans all 4 columns.',
      },
      {
        type: 'note',
        text: 'Time Range: Use the timeframe dropdown in the toolbar to change the time period for all widgets (1 minute to 30 days).',
      },
    ],
  },
  {
    id: 'traces',
    title: 'Traces',
    subsections: [
      { id: 'traces-filtering', title: 'Filtering & Search' },
      { id: 'traces-waterfall', title: 'Waterfall View' },
      { id: 'traces-details', title: 'Span Details' },
    ],
    content: [
      {
        type: 'paragraph',
        text: 'The Traces page shows distributed traces from your AI coding tools. Each trace represents a series of related operations, such as an API request and its associated tool executions.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Filtering & Search',
      },
      {
        type: 'paragraph',
        text: 'Use the controls at the top of the page to filter traces:',
      },
      {
        type: 'list',
        items: [
          'Timeframe - Select a time range from 1 minute to 30 days (default: 7 days)',
          'Service - Filter by a specific service name, or show all services',
          'Search - Full-text search across span names, errors, and attributes',
        ],
      },
      {
        type: 'paragraph',
        text: 'When new traces arrive via WebSocket, a badge appears showing the count of new traces. Click the refresh button to load the latest data.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Waterfall View',
      },
      {
        type: 'paragraph',
        text: 'Click on any trace in the list to expand it and see the waterfall visualization. The waterfall shows:',
      },
      {
        type: 'list',
        items: [
          'Hierarchical span structure with parent-child relationships',
          'Duration bars scaled to the trace timeline',
          'Color-coded status (green for OK, red for ERROR)',
          'Span names with icons indicating the span kind (CLIENT, SERVER, INTERNAL)',
        ],
      },
      {
        type: 'paragraph',
        text: 'Spans with children can be expanded or collapsed by clicking the chevron icon. This helps navigate complex traces with many nested operations.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Span Details',
      },
      {
        type: 'paragraph',
        text: 'Click on any span in the waterfall to view its details:',
      },
      {
        type: 'list',
        items: [
          'Attributes - Key-value metadata attached to the span (e.g., model name, token counts)',
          'Events - Timestamped events that occurred during the span execution',
          'Status - Error messages and status codes for failed spans',
        ],
      },
      {
        type: 'note',
        text: 'Codex CLI traces: Codex CLI uses a single trace per session, so long sessions produce traces with many spans. AI Observer splits these into manageable units in the trace list.',
      },
    ],
  },
  {
    id: 'metrics',
    title: 'Metrics',
    subsections: [
      { id: 'metrics-selection', title: 'Metric Selection' },
      { id: 'metrics-charts', title: 'Chart Visualization' },
      { id: 'metrics-timerange', title: 'Time Range' },
    ],
    content: [
      {
        type: 'paragraph',
        text: 'The Metrics page displays time-series data for the metrics collected from your AI coding tools. This includes token usage, API latency, costs, and various operational metrics.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Metric Selection',
      },
      {
        type: 'paragraph',
        text: 'Use the dropdown menus to select which data to display:',
      },
      {
        type: 'list',
        items: [
          'Service - Optionally filter to a specific service',
          'Metric - Choose from available metrics, grouped by source tool:',
        ],
      },
      {
        type: 'note',
        text: 'Claude Code metrics include: Sessions, Token Usage, Cost, Lines of Code, Pull Requests, Commits, Edit Decisions, Active Time',
      },
      {
        type: 'note',
        text: 'Gemini CLI metrics include: Sessions, Token Usage, Cost, API Requests, API Latency, Tool Calls, File Operations, Agent Runs, and more',
      },
      {
        type: 'note',
        text: 'Codex CLI metrics include: Token Usage, Cost (derived from log events)',
      },
      {
        type: 'paragraph',
        text: 'When you select a metric, its metadata is displayed below the selection: the metric name, unit of measurement, and description.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Chart Visualization',
      },
      {
        type: 'paragraph',
        text: 'The metric data is displayed as a bar chart with time on the x-axis. Two visualization modes are available:',
      },
      {
        type: 'list',
        items: [
          'Stacked - Bars are stacked on top of each other, showing cumulative totals',
          'Grouped - Bars are placed side-by-side for direct comparison between series',
        ],
      },
      {
        type: 'paragraph',
        text: 'Use the toggle buttons above the chart to switch between modes. The legend below the chart shows all data series with their colors.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Time Range',
      },
      {
        type: 'paragraph',
        text: 'Select a time range from the dropdown to adjust the displayed period. Available ranges span from 1 minute to 30 days. The chart automatically adjusts its aggregation interval based on the selected timeframe:',
      },
      {
        type: 'list',
        items: [
          'Short ranges (1-15 minutes) - Per-minute aggregation',
          'Medium ranges (1-24 hours) - 5-minute to hourly aggregation',
          'Long ranges (7-30 days) - Daily aggregation',
        ],
      },
      {
        type: 'paragraph',
        text: 'The chart auto-refreshes periodically based on the selected timeframe. A badge shows when new data has arrived.',
      },
    ],
  },
  {
    id: 'logs',
    title: 'Logs',
    subsections: [
      { id: 'logs-filtering', title: 'Filtering' },
      { id: 'logs-details', title: 'Log Details' },
      { id: 'logs-correlation', title: 'Trace Correlation' },
    ],
    content: [
      {
        type: 'paragraph',
        text: 'The Logs page displays structured log records from your AI coding tools. Logs capture events like user prompts, API requests, tool executions, and errors.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Filtering',
      },
      {
        type: 'paragraph',
        text: 'Use the filter controls to narrow down the displayed logs:',
      },
      {
        type: 'list',
        items: [
          'Service - Filter by a specific service name',
          'Severity - Filter by log level (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)',
          'Search - Full-text search in log messages',
        ],
      },
      {
        type: 'paragraph',
        text: 'Severity levels are color-coded: ERROR and FATAL appear in red, WARN in yellow, and informational levels in neutral colors.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Log Details',
      },
      {
        type: 'paragraph',
        text: 'Each log entry shows:',
      },
      {
        type: 'list',
        items: [
          'Timestamp - When the log was recorded',
          'Severity badge - The log level',
          'Service name - Which tool generated the log',
          'Message preview - Truncated log body',
        ],
      },
      {
        type: 'paragraph',
        text: 'Click on any log entry to expand it and see the full details:',
      },
      {
        type: 'list',
        items: [
          'Full message - Complete log body text',
          'Attributes - Structured metadata (key-value pairs) attached to the log',
        ],
      },
      {
        type: 'heading',
        level: 3,
        text: 'Trace Correlation',
      },
      {
        type: 'paragraph',
        text: 'Many logs are correlated with traces, allowing you to see the broader context of an operation. When a log has trace context, the expanded view shows:',
      },
      {
        type: 'list',
        items: [
          'Trace ID - Links the log to a distributed trace',
          'Span ID - Identifies the specific span within the trace',
        ],
      },
      {
        type: 'note',
        text: 'Real-time updates: When new logs arrive, a badge shows the count. Click the refresh button to load the latest entries.',
      },
    ],
  },
  {
    id: 'telemetry',
    title: 'Telemetry Reference',
    subsections: [
      { id: 'telemetry-claude', title: 'Claude Code' },
      { id: 'telemetry-gemini', title: 'Gemini CLI' },
      { id: 'telemetry-codex', title: 'Codex CLI' },
    ],
    content: [
      {
        type: 'paragraph',
        text: 'AI Observer collects OpenTelemetry metrics and events from various AI coding tools. Each metric includes metadata for display names, descriptions, units, and breakdown attributes for multi-series visualization.',
      },
      {
        type: 'heading',
        level: 3,
        text: 'Claude Code Metrics',
      },
      {
        type: 'table',
        table: {
          headers: ['Metric Name', 'Display Name', 'Unit', 'Description'],
          rows: [
            { cells: ['claude_code.session.count', 'Sessions', 'count', 'Number of coding sessions started'] },
            { cells: ['claude_code.lines_of_code.count', 'Lines of Code', 'count', 'Lines of code added or removed'] },
            { cells: ['claude_code.pull_request.count', 'Pull Requests', 'count', 'Number of pull requests created'] },
            { cells: ['claude_code.commit.count', 'Commits', 'count', 'Number of commits made'] },
            { cells: ['claude_code.cost.usage', 'Cost', 'USD', 'Total cost incurred in USD'] },
            { cells: ['claude_code.token.usage', 'Token Usage', 'tokens', 'Token consumption by type'] },
            { cells: ['claude_code.code_edit_tool.decision', 'Edit Decisions', 'count', 'Code edit tool usage decisions'] },
            { cells: ['claude_code.active_time.total', 'Active Time', 'seconds', 'Total active coding time'] },
          ],
        },
      },
      {
        type: 'heading',
        level: 3,
        text: 'Gemini CLI Metrics',
      },
      {
        type: 'table',
        table: {
          headers: ['Metric Name', 'Display Name', 'Unit', 'Description'],
          rows: [
            { cells: ['gemini_cli.session.count', 'Sessions', 'count', 'Number of CLI sessions'] },
            { cells: ['gemini_cli.tool.call.count', 'Tool Calls', 'count', 'Number of tool invocations'] },
            { cells: ['gemini_cli.tool.call.latency', 'Tool Latency', 'ms', 'Tool call execution time'] },
            { cells: ['gemini_cli.api.request.count', 'API Requests', 'count', 'Number of API requests made'] },
            { cells: ['gemini_cli.api.request.latency', 'API Latency', 'ms', 'API request latency'] },
            { cells: ['gemini_cli.token.usage', 'Token Usage', 'tokens', 'Token consumption'] },
            { cells: ['gemini_cli.file.operation.count', 'File Operations', 'count', 'File read/write operations'] },
            { cells: ['gemini_cli.lines.changed', 'Lines Changed', 'count', 'Code lines modified'] },
            { cells: ['gemini_cli.agent.run.count', 'Agent Runs', 'count', 'Number of agent executions'] },
            { cells: ['gemini_cli.agent.duration', 'Agent Duration', 'ms', 'Agent execution time'] },
            { cells: ['gemini_cli.agent.turns', 'Agent Turns', 'count', 'Conversation turns per agent'] },
            { cells: ['gemini_cli.chat_compression', 'Chat Compression', 'count', 'Chat message compression events'] },
            { cells: ['gemini_cli.startup.duration', 'Startup Duration', 'ms', 'CLI startup time'] },
            { cells: ['gemini_cli.memory.usage', 'Memory Usage', 'bytes', 'Memory consumption'] },
            { cells: ['gemini_cli.cpu.usage', 'CPU Usage', '%', 'CPU utilization'] },
            { cells: ['gen_ai.client.token.usage', 'GenAI Token Usage', 'tokens', 'Generic AI token usage'] },
            { cells: ['gen_ai.client.operation.duration', 'GenAI Op Duration', 'seconds', 'Generic AI operation duration'] },
          ],
        },
      },
      {
        type: 'note',
        text: 'Gemini CLI Derived Metrics',
      },
      {
        type: 'paragraph',
        text: 'AI Observer computes delta metrics from cumulative counters to show per-interval changes:',
      },
      {
        type: 'table',
        table: {
          headers: ['Metric Name', 'Display Name', 'Description'],
          rows: [
            { cells: ['gemini_cli.session.count.delta', 'Sessions', 'Sessions per interval'] },
            { cells: ['gemini_cli.token.usage.delta', 'Token Usage', 'Tokens consumed per interval'] },
            { cells: ['gemini_cli.api.request.count.delta', 'API Requests', 'API requests per interval'] },
            { cells: ['gemini_cli.file.operation.count.delta', 'File Operations', 'File operations per interval'] },
            { cells: ['gen_ai.client.token.usage.delta', 'GenAI Token Usage', 'Token consumption per interval'] },
          ],
        },
      },
      {
        type: 'heading',
        level: 3,
        text: 'Codex CLI Metrics & Events',
      },
      {
        type: 'paragraph',
        text: 'Codex CLI exports logs and traces directly. AI Observer derives the following metrics from log events:',
      },
      {
        type: 'note',
        text: 'Codex CLI Derived Metrics',
      },
      {
        type: 'table',
        table: {
          headers: ['Metric Name', 'Display Name', 'Unit', 'Description'],
          rows: [
            { cells: ['codex_cli_rs.token.usage', 'Token Usage', 'tokens', 'Tokens by type (input/output/cache/reasoning/tool)'] },
            { cells: ['codex_cli_rs.cost.usage', 'Cost', 'USD', 'Session cost in USD'] },
          ],
        },
      },
      {
        type: 'note',
        text: 'Codex CLI Events',
      },
      {
        type: 'table',
        table: {
          headers: ['Event Name', 'Display Name', 'Description'],
          rows: [
            { cells: ['codex.conversation_starts', 'Sessions', 'Conversation session started'] },
            { cells: ['codex.api_request', 'API Requests', 'API request made'] },
            { cells: ['codex.user_prompt', 'User Prompts', 'User prompt submitted'] },
            { cells: ['codex.tool_decision', 'Tool Decisions', 'Tool execution decisions'] },
            { cells: ['codex.tool_result', 'Tool Results', 'Tool execution outcomes'] },
          ],
        },
      },
      {
        type: 'heading',
        level: 3,
        text: 'Breakdown Attributes',
      },
      {
        type: 'paragraph',
        text: 'Many metrics support breakdown attributes for multi-series visualization. When viewing a chart, you can split the data by various dimensions:',
      },
      {
        type: 'list',
        items: [
          'Token Usage metrics - breakdown by type (input, output, cache) or by model',
          'Lines Changed metrics - breakdown by type (added vs removed)',
          'Tool metrics - breakdown by function_name, decision, or success status',
          'API metrics - breakdown by model, status_code, or error_type',
        ],
      },
      {
        type: 'note',
        text: 'The breakdown attribute can be configured per-widget in the dashboard. If not specified, the default breakdown (usually type) is used.',
      },
    ],
  },
]
