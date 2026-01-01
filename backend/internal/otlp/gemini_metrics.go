package otlp

import (
	"strings"

	"github.com/tobilg/ai-observer/internal/api"
)

// GeminiPricing contains per-token pricing for Gemini models in USD
type GeminiPricing struct {
	InputCostPerToken     float64
	OutputCostPerToken    float64
	CacheReadCostPerToken float64
}

// geminiPricingTable maps model names to their pricing (per token in USD)
// Prices from Google AI pricing (https://ai.google.dev/gemini-api/docs/pricing)
// Converted from USD per 1M tokens to USD per token
var geminiPricingTable = map[string]GeminiPricing{
	// Gemini 3 series
	"gemini-3-pro-preview":   {InputCostPerToken: 2e-6, OutputCostPerToken: 12e-6, CacheReadCostPerToken: 0.2e-6},
	"gemini-3-flash-preview": {InputCostPerToken: 0.5e-6, OutputCostPerToken: 3e-6, CacheReadCostPerToken: 0.05e-6},
	// Gemini 2.5 series
	"gemini-2.5-pro":                          {InputCostPerToken: 1.25e-6, OutputCostPerToken: 10e-6, CacheReadCostPerToken: 0.125e-6},
	"gemini-2.5-flash":                        {InputCostPerToken: 0.3e-6, OutputCostPerToken: 2.5e-6, CacheReadCostPerToken: 0.03e-6},
	"gemini-2.5-flash-preview-09-2025":        {InputCostPerToken: 0.3e-6, OutputCostPerToken: 2.5e-6, CacheReadCostPerToken: 0.03e-6},
	"gemini-2.5-flash-lite":                   {InputCostPerToken: 0.1e-6, OutputCostPerToken: 0.4e-6, CacheReadCostPerToken: 0.01e-6},
	"gemini-2.5-flash-lite-preview-09-2025":   {InputCostPerToken: 0.1e-6, OutputCostPerToken: 0.4e-6, CacheReadCostPerToken: 0.01e-6},
	"gemini-2.5-computer-use-preview-10-2025": {InputCostPerToken: 1.25e-6, OutputCostPerToken: 10e-6, CacheReadCostPerToken: 0},
	// Gemini 2.0 series
	"gemini-2.0-flash":      {InputCostPerToken: 0.1e-6, OutputCostPerToken: 0.4e-6, CacheReadCostPerToken: 0.025e-6},
	"gemini-2.0-flash-lite": {InputCostPerToken: 0.075e-6, OutputCostPerToken: 0.3e-6, CacheReadCostPerToken: 0},
}

// Gemini token type constants matching gemini_cli.token.usage
const (
	GeminiTokenTypeInput   = "input"
	GeminiTokenTypeOutput  = "output"
	GeminiTokenTypeCache   = "cache"
	GeminiTokenTypeThought = "thought"
	GeminiTokenTypeTool    = "tool"
)

// Metric names
const (
	GeminiTokenUsageMetric = "gemini_cli.token.usage"
	GeminiCostUsageMetric  = "gemini_cli.cost.usage"
)

// NormalizeGeminiModel normalizes a Gemini model name for pricing lookup.
func NormalizeGeminiModel(model string) string {
	return strings.TrimSpace(model)
}

// CalculateGeminiCostForTokenType calculates cost for a specific token type.
// Returns nil if the model is not in the pricing table or token count is zero/negative.
func CalculateGeminiCostForTokenType(model string, tokenType string, tokenCount int64) *float64 {
	normalizedModel := NormalizeGeminiModel(model)
	pricing, ok := geminiPricingTable[normalizedModel]
	if !ok {
		return nil
	}

	if tokenCount <= 0 {
		return nil
	}

	var cost float64
	switch tokenType {
	case GeminiTokenTypeInput:
		cost = float64(tokenCount) * pricing.InputCostPerToken
	case GeminiTokenTypeOutput, GeminiTokenTypeThought:
		// Thought tokens are charged at output rate
		cost = float64(tokenCount) * pricing.OutputCostPerToken
	case GeminiTokenTypeCache:
		cost = float64(tokenCount) * pricing.CacheReadCostPerToken
	case GeminiTokenTypeTool:
		// Tool tokens don't have direct cost
		return nil
	default:
		return nil
	}

	return &cost
}

// DeriveGeminiCostMetric creates a cost metric from a token usage metric.
// Returns nil if cost cannot be calculated (wrong metric name, missing attributes, unknown model, etc.).
func DeriveGeminiCostMetric(tokenMetric api.MetricDataPoint) *api.MetricDataPoint {
	// Only process gemini_cli.token.usage metrics
	if tokenMetric.MetricName != GeminiTokenUsageMetric {
		return nil
	}

	// Extract model and type from attributes
	model := tokenMetric.Attributes["model"]
	tokenType := tokenMetric.Attributes["type"]
	if model == "" || tokenType == "" {
		return nil
	}

	// Get token count from Value
	if tokenMetric.Value == nil {
		return nil
	}
	tokenCount := int64(*tokenMetric.Value)

	// Calculate cost
	cost := CalculateGeminiCostForTokenType(model, tokenType, tokenCount)
	if cost == nil {
		return nil
	}

	// Create cost metric
	metricType := "sum"
	isMonotonic := true
	aggregationTemporality := int32(1) // DELTA - each cost is per-event, not cumulative

	costAttrs := map[string]string{
		"model": model,
	}

	return &api.MetricDataPoint{
		Timestamp:              tokenMetric.Timestamp,
		ServiceName:            tokenMetric.ServiceName,
		MetricName:             GeminiCostUsageMetric,
		MetricDescription:      "Total cost in USD for Gemini CLI usage",
		MetricUnit:             "USD",
		ResourceAttributes:     tokenMetric.ResourceAttributes,
		ScopeName:              tokenMetric.ScopeName,
		ScopeVersion:           tokenMetric.ScopeVersion,
		Attributes:             costAttrs,
		MetricType:             metricType,
		Value:                  cost,
		IsMonotonic:            &isMonotonic,
		AggregationTemporality: &aggregationTemporality,
	}
}
