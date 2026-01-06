package optim

import (
	"context"
	"math"

	"github.com/san-kum/dynsim/internal/experiment"
)

type GridSearch struct {
	paramNames []string
	ranges     [][]float64
}

func NewGridSearch(params []string, ranges [][]float64) *GridSearch {
	return &GridSearch{paramNames: params, ranges: ranges}
}

func (g *GridSearch) Search(
	ctx context.Context,
	buildExperiment func(params map[string]float64) (*experiment.Experiment, error),
	metricName string,
) (map[string]float64, float64, error) {

	best := math.Inf(1)
	var bestParams map[string]float64

	g.searchRecursive(ctx, 0, make(map[string]float64), buildExperiment, metricName, &best, &bestParams)

	return bestParams, best, nil
}

func (g *GridSearch) searchRecursive(
	ctx context.Context,
	depth int,
	current map[string]float64,
	buildExperiment func(map[string]float64) (*experiment.Experiment, error),
	metricName string,
	best *float64,
	bestParams *map[string]float64,
) {
	if depth == len(g.paramNames) {
		exp, err := buildExperiment(current)
		if err != nil {
			return
		}

		result, err := exp.Run(ctx)
		if err != nil {
			return
		}

		val := result.Metrics[metricName]
		if val < *best {
			*best = val
			*bestParams = make(map[string]float64)
			for k, v := range current {
				(*bestParams)[k] = v
			}
		}
		return
	}

	paramName := g.paramNames[depth]
	for _, val := range g.ranges[depth] {
		newParams := make(map[string]float64)
		for k, v := range current {
			newParams[k] = v
		}
		newParams[paramName] = val

		g.searchRecursive(ctx, depth+1, newParams, buildExperiment, metricName, best, bestParams)
	}
}
