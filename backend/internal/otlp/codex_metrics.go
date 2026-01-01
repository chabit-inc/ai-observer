package otlp

import (
	"strconv"
	"strings"
	"time"

	"github.com/tobilg/ai-observer/internal/api"
)

// CodexPricing contains per-token pricing for Codex models in USD
type CodexPricing struct {
	InputCostPerToken     float64
	OutputCostPerToken    float64
	CacheReadCostPerToken float64
}

// codexPricingTable maps model names to their pricing (per token in USD)
// Prices from OpenAI API pricing (https://platform.openai.com/docs/pricing)
// Converted from USD per 1M tokens to USD per token
var codexPricingTable = map[string]CodexPricing{
	// GPT-5 series
	"gpt-5":              {InputCostPerToken: 1.25e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-7},
	"gpt-5.1":            {InputCostPerToken: 1.25e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-7},
	"gpt-5.2":            {InputCostPerToken: 1.75e-6, OutputCostPerToken: 1.4e-5, CacheReadCostPerToken: 1.75e-7},
	"gpt-5-mini":         {InputCostPerToken: 2.5e-7, OutputCostPerToken: 2e-6, CacheReadCostPerToken: 2.5e-8},
	"gpt-5-nano":         {InputCostPerToken: 5e-8, OutputCostPerToken: 4e-7, CacheReadCostPerToken: 5e-9},
	"gpt-5-pro":          {InputCostPerToken: 1.5e-5, OutputCostPerToken: 1.2e-4, CacheReadCostPerToken: 0},
	"gpt-5.2-pro":        {InputCostPerToken: 2.1e-5, OutputCostPerToken: 1.68e-4, CacheReadCostPerToken: 0},
	"gpt-5-chat-latest":  {InputCostPerToken: 1.25e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-7},
	"gpt-5.1-chat-latest": {InputCostPerToken: 1.25e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-7},
	"gpt-5.2-chat-latest": {InputCostPerToken: 1.75e-6, OutputCostPerToken: 1.4e-5, CacheReadCostPerToken: 1.75e-7},
	"gpt-5-codex":        {InputCostPerToken: 1.25e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-7},
	"gpt-5.1-codex":      {InputCostPerToken: 1.25e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-7},
	"gpt-5.1-codex-max":  {InputCostPerToken: 1.25e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-7},
	"gpt-5.1-codex-mini": {InputCostPerToken: 2.5e-7, OutputCostPerToken: 2e-6, CacheReadCostPerToken: 2.5e-8},
	"gpt-5-search-api":   {InputCostPerToken: 1.25e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-7},
	"codex-mini-latest":  {InputCostPerToken: 1.5e-6, OutputCostPerToken: 6e-6, CacheReadCostPerToken: 3.75e-7},
	// GPT-4.1 series
	"gpt-4.1":      {InputCostPerToken: 2e-6, OutputCostPerToken: 8e-6, CacheReadCostPerToken: 5e-7},
	"gpt-4.1-mini": {InputCostPerToken: 4e-7, OutputCostPerToken: 1.6e-6, CacheReadCostPerToken: 1e-7},
	"gpt-4.1-nano": {InputCostPerToken: 1e-7, OutputCostPerToken: 4e-7, CacheReadCostPerToken: 2.5e-8},
	// GPT-4o series
	"gpt-4o":                       {InputCostPerToken: 2.5e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 1.25e-6},
	"gpt-4o-2024-05-13":            {InputCostPerToken: 5e-6, OutputCostPerToken: 1.5e-5, CacheReadCostPerToken: 0},
	"gpt-4o-mini":                  {InputCostPerToken: 1.5e-7, OutputCostPerToken: 6e-7, CacheReadCostPerToken: 7.5e-8},
	"gpt-4o-search-preview":        {InputCostPerToken: 2.5e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 0},
	"gpt-4o-mini-search-preview":   {InputCostPerToken: 1.5e-7, OutputCostPerToken: 6e-7, CacheReadCostPerToken: 0},
	"gpt-4o-realtime-preview":      {InputCostPerToken: 5e-6, OutputCostPerToken: 2e-5, CacheReadCostPerToken: 2.5e-6},
	"gpt-4o-mini-realtime-preview": {InputCostPerToken: 6e-7, OutputCostPerToken: 2.4e-6, CacheReadCostPerToken: 3e-7},
	"gpt-4o-audio-preview":         {InputCostPerToken: 2.5e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 0},
	"gpt-4o-mini-audio-preview":    {InputCostPerToken: 1.5e-7, OutputCostPerToken: 6e-7, CacheReadCostPerToken: 0},
	// Realtime/Audio models
	"gpt-realtime":      {InputCostPerToken: 4e-6, OutputCostPerToken: 1.6e-5, CacheReadCostPerToken: 4e-7},
	"gpt-realtime-mini": {InputCostPerToken: 6e-7, OutputCostPerToken: 2.4e-6, CacheReadCostPerToken: 6e-8},
	"gpt-audio":         {InputCostPerToken: 2.5e-6, OutputCostPerToken: 1e-5, CacheReadCostPerToken: 0},
	"gpt-audio-mini":    {InputCostPerToken: 6e-7, OutputCostPerToken: 2.4e-6, CacheReadCostPerToken: 0},
	// O-series reasoning models
	"o1":                     {InputCostPerToken: 1.5e-5, OutputCostPerToken: 6e-5, CacheReadCostPerToken: 7.5e-6},
	"o1-mini":                {InputCostPerToken: 1.1e-6, OutputCostPerToken: 4.4e-6, CacheReadCostPerToken: 5.5e-7},
	"o1-pro":                 {InputCostPerToken: 1.5e-4, OutputCostPerToken: 6e-4, CacheReadCostPerToken: 0},
	"o3":                     {InputCostPerToken: 2e-6, OutputCostPerToken: 8e-6, CacheReadCostPerToken: 5e-7},
	"o3-mini":                {InputCostPerToken: 1.1e-6, OutputCostPerToken: 4.4e-6, CacheReadCostPerToken: 5.5e-7},
	"o3-pro":                 {InputCostPerToken: 2e-5, OutputCostPerToken: 8e-5, CacheReadCostPerToken: 0},
	"o3-deep-research":       {InputCostPerToken: 1e-5, OutputCostPerToken: 4e-5, CacheReadCostPerToken: 2.5e-6},
	"o4-mini":                {InputCostPerToken: 1.1e-6, OutputCostPerToken: 4.4e-6, CacheReadCostPerToken: 2.75e-7},
	"o4-mini-deep-research":  {InputCostPerToken: 2e-6, OutputCostPerToken: 8e-6, CacheReadCostPerToken: 5e-7},
	// Other
	"computer-use-preview": {InputCostPerToken: 3e-6, OutputCostPerToken: 1.2e-5, CacheReadCostPerToken: 0},
}

// Token type constants matching claude_code.token.usage
const (
	TokenTypeInput     = "input"
	TokenTypeOutput    = "output"
	TokenTypeCacheRead = "cacheRead"
	TokenTypeReasoning = "reasoning"
	TokenTypeTool      = "tool"
)

// Metric names
const (
	CodexTokenUsageMetric = "codex_cli_rs.token.usage"
	CodexCostUsageMetric  = "codex_cli_rs.cost.usage"
)

// ExtractCodexMetrics extracts token usage and cost metrics from a codex.sse_event log record.
// Returns nil if the event is not a response.completed event or has no token data.
func ExtractCodexMetrics(
	logAttrs map[string]string,
	timestamp time.Time,
	serviceName string,
	resourceAttrs map[string]string,
	traceID string,
	spanID string,
) []api.MetricDataPoint {
	// Only process response.completed events
	eventKind, ok := logAttrs["event.kind"]
	if !ok || eventKind != "response.completed" {
		return nil
	}

	// Extract token counts
	inputTokens := parseIntAttr(logAttrs, "input_token_count")
	outputTokens := parseIntAttr(logAttrs, "output_token_count")
	cachedTokens := parseIntAttr(logAttrs, "cached_token_count")
	reasoningTokens := parseIntAttr(logAttrs, "reasoning_token_count")
	toolTokens := parseIntAttr(logAttrs, "tool_token_count")

	// If no tokens at all, skip
	if inputTokens == 0 && outputTokens == 0 && cachedTokens == 0 && reasoningTokens == 0 && toolTokens == 0 {
		return nil
	}

	// Extract model name
	model := logAttrs["model"]
	if model == "" {
		model = resourceAttrs["model"]
	}
	if model == "" {
		model = "unknown"
	}

	var metrics []api.MetricDataPoint

	// Create base metric attributes
	baseAttrs := map[string]string{
		"model": model,
	}

	// Helper to create a token usage metric data point
	createTokenMetric := func(tokenType string, value int64) api.MetricDataPoint {
		attrs := make(map[string]string)
		for k, v := range baseAttrs {
			attrs[k] = v
		}
		attrs["type"] = tokenType

		floatValue := float64(value)
		metricType := "sum"
		isMonotonic := true
		aggregationTemporality := int32(2) // CUMULATIVE

		return api.MetricDataPoint{
			Timestamp:              timestamp,
			ServiceName:            serviceName,
			MetricName:             CodexTokenUsageMetric,
			MetricDescription:      "Number of tokens consumed by Codex CLI",
			MetricUnit:             "tokens",
			ResourceAttributes:     resourceAttrs,
			Attributes:             attrs,
			MetricType:             metricType,
			Value:                  &floatValue,
			IsMonotonic:            &isMonotonic,
			AggregationTemporality: &aggregationTemporality,
		}
	}

	// Add token metrics for each type with non-zero values
	if inputTokens > 0 {
		metrics = append(metrics, createTokenMetric(TokenTypeInput, inputTokens))
	}
	if outputTokens > 0 {
		metrics = append(metrics, createTokenMetric(TokenTypeOutput, outputTokens))
	}
	if cachedTokens > 0 {
		metrics = append(metrics, createTokenMetric(TokenTypeCacheRead, cachedTokens))
	}
	if reasoningTokens > 0 {
		metrics = append(metrics, createTokenMetric(TokenTypeReasoning, reasoningTokens))
	}
	if toolTokens > 0 {
		metrics = append(metrics, createTokenMetric(TokenTypeTool, toolTokens))
	}

	// Calculate and add cost metric if model is known
	cost := CalculateCodexCost(model, inputTokens, cachedTokens, outputTokens)
	if cost != nil {
		metricType := "sum"
		isMonotonic := true
		aggregationTemporality := int32(2) // CUMULATIVE

		costMetric := api.MetricDataPoint{
			Timestamp:              timestamp,
			ServiceName:            serviceName,
			MetricName:             CodexCostUsageMetric,
			MetricDescription:      "Total cost in USD for Codex CLI usage",
			MetricUnit:             "USD",
			ResourceAttributes:     resourceAttrs,
			Attributes:             baseAttrs,
			MetricType:             metricType,
			Value:                  cost,
			IsMonotonic:            &isMonotonic,
			AggregationTemporality: &aggregationTemporality,
		}
		metrics = append(metrics, costMetric)
	}

	return metrics
}

// NormalizeCodexModel normalizes a Codex model name for pricing lookup.
// It strips provider prefixes like "openai/".
func NormalizeCodexModel(model string) string {
	trimmed := strings.TrimSpace(model)

	// Strip "openai/" prefix
	if strings.HasPrefix(trimmed, "openai/") {
		trimmed = strings.TrimPrefix(trimmed, "openai/")
	}

	return trimmed
}

// CalculateCodexCost calculates the cost in USD for Codex token usage.
// Returns nil if the model is not in the pricing table.
// Cost formula from CodexBar CostUsagePricing.swift:160-168
func CalculateCodexCost(model string, inputTokens, cachedTokens, outputTokens int64) *float64 {
	normalizedModel := NormalizeCodexModel(model)
	pricing, ok := codexPricingTable[normalizedModel]
	if !ok {
		return nil
	}

	// Clamp values to non-negative
	input := max(0, inputTokens)
	cached := max(0, cachedTokens)
	output := max(0, outputTokens)

	// Cached tokens can't exceed input tokens
	cached = min(cached, input)

	// Non-cached input tokens
	nonCached := input - cached

	// Calculate cost
	cost := float64(nonCached)*pricing.InputCostPerToken +
		float64(cached)*pricing.CacheReadCostPerToken +
		float64(output)*pricing.OutputCostPerToken

	return &cost
}

// parseIntAttr parses an integer attribute from a string map
func parseIntAttr(attrs map[string]string, key string) int64 {
	if val, ok := attrs[key]; ok {
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return 0
}
