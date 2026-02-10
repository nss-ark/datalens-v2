package detection

import (
	"context"
	"sort"
	"time"

	"github.com/complyark/datalens/pkg/types"
)

// ComposableDetector chains multiple detection strategies and merges
// their results using weighted confidence scoring. This is the top-level
// entry point for all PII detection in the system.
//
// Confidence routing:
//
//	≥ 0.95  → AUTO_VERIFY    (no human review)
//	0.80–0.95 → QUICK_VERIFY  (one-click confirm)
//	0.50–0.80 → MANUAL_REVIEW (human inspects)
//	< 0.50  → LOW_CONFIDENCE (flagged)
type ComposableDetector struct {
	strategies []Strategy
}

// taggedResult pairs a detection result with the strategy's weight.
type taggedResult struct {
	result Result
	weight float64
}

// NewComposableDetector creates a detector with the given strategies.
// Strategies are run in the order provided.
func NewComposableDetector(strategies ...Strategy) *ComposableDetector {
	return &ComposableDetector{
		strategies: strategies,
	}
}

// Detect runs all strategies against the input and produces a merged report.
func (d *ComposableDetector) Detect(ctx context.Context, input Input) (*Report, error) {
	start := time.Now()

	report := &Report{
		ColumnName: input.ColumnName,
		Strategies: make([]StrategyOutcome, 0, len(d.strategies)),
	}

	// Collect all raw results from all strategies
	var allResults []taggedResult

	for _, strategy := range d.strategies {
		stratStart := time.Now()
		results, err := strategy.Detect(ctx, input)
		stratDuration := time.Since(stratStart)

		outcome := StrategyOutcome{
			Name:     strategy.Name(),
			Method:   strategy.Method(),
			Found:    len(results) > 0,
			Results:  len(results),
			Duration: stratDuration,
		}
		if err != nil {
			outcome.Error = err.Error()
		}
		report.Strategies = append(report.Strategies, outcome)

		for _, r := range results {
			allResults = append(allResults, taggedResult{result: r, weight: strategy.Weight()})
		}
	}

	if len(allResults) == 0 {
		report.IsPII = false
		report.Duration = time.Since(start)
		return report, nil
	}

	// Merge results: group by PIIType, then compute weighted confidence
	report.Detections = d.mergeResults(allResults)
	report.IsPII = len(report.Detections) > 0
	report.Duration = time.Since(start)

	// Set top match (highest confidence)
	if len(report.Detections) > 0 {
		top := report.Detections[0]
		report.TopMatch = &top
	}

	return report, nil
}

// mergeResults groups raw results by PIIType and computes weighted confidence.
func (d *ComposableDetector) mergeResults(tagged []taggedResult) []MergedDetection {
	// Group by PIIType
	type group struct {
		category    types.PIICategory
		piiType     types.PIIType
		sensitivity types.SensitivityLevel
		methods     map[types.DetectionMethod]bool
		reasoning   []string
		weightedSum float64
		totalWeight float64
	}

	groups := make(map[types.PIIType]*group)

	for _, t := range tagged {
		r := t.result
		g, ok := groups[r.Type]
		if !ok {
			g = &group{
				category:    r.Category,
				piiType:     r.Type,
				sensitivity: r.Sensitivity,
				methods:     make(map[types.DetectionMethod]bool),
			}
			groups[r.Type] = g
		}
		g.methods[r.Method] = true
		g.reasoning = append(g.reasoning, r.Reasoning)
		g.weightedSum += t.weight * r.Confidence
		g.totalWeight += t.weight

		// Upgrade sensitivity if a higher one is found
		if compareSensitivity(r.Sensitivity, g.sensitivity) > 0 {
			g.sensitivity = r.Sensitivity
		}
	}

	// Convert groups to MergedDetections
	var merged []MergedDetection
	for _, g := range groups {
		finalConf := g.weightedSum / g.totalWeight

		// Multi-method boost: if multiple strategies agree, boost confidence
		if len(g.methods) >= 2 {
			finalConf = boostMultiMethod(finalConf, len(g.methods))
		}

		// Cap at 1.0
		if finalConf > 1.0 {
			finalConf = 1.0
		}

		methods := make([]types.DetectionMethod, 0, len(g.methods))
		for m := range g.methods {
			methods = append(methods, m)
		}

		merged = append(merged, MergedDetection{
			Category:        g.category,
			Type:            g.piiType,
			Sensitivity:     g.sensitivity,
			FinalConfidence: finalConf,
			Methods:         methods,
			Reasoning:       mergeReasoning(g.reasoning),
			RequiresReview:  finalConf < 0.80,
		})
	}

	// Sort by confidence descending
	sort.Slice(merged, func(i, j int) bool {
		return merged[i].FinalConfidence > merged[j].FinalConfidence
	})

	return merged
}

// boostMultiMethod increases confidence when multiple strategies agree.
// 2 strategies agreeing: +5%, 3+: +10%.
func boostMultiMethod(confidence float64, methodCount int) float64 {
	switch {
	case methodCount >= 3:
		return confidence * 1.10
	case methodCount == 2:
		return confidence * 1.05
	default:
		return confidence
	}
}

// compareSensitivity returns >0 if a is higher than b.
func compareSensitivity(a, b types.SensitivityLevel) int {
	order := map[types.SensitivityLevel]int{
		types.SensitivityLow:      1,
		types.SensitivityMedium:   2,
		types.SensitivityHigh:     3,
		types.SensitivityCritical: 4,
	}
	return order[a] - order[b]
}

// mergeReasoning combines multiple reasoning strings into one.
func mergeReasoning(reasons []string) string {
	if len(reasons) == 0 {
		return ""
	}
	if len(reasons) == 1 {
		return reasons[0]
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, r := range reasons {
		if !seen[r] {
			seen[r] = true
			unique = append(unique, r)
		}
	}

	result := unique[0]
	for _, r := range unique[1:] {
		result += "; " + r
	}
	return result
}
