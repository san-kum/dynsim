package sim

import (
	"context"
	"sync"
)

type Ensemble struct {
	base      *Simulator
	numRuns   int
	seedStart int64
}

func NewEnsemble(s *Simulator, numRuns int, seedStart int64) *Ensemble {
	return &Ensemble{base: s, numRuns: numRuns, seedStart: seedStart}
}

func (e *Ensemble) Run(ctx context.Context, x0 State, cfg Config) ([]*Result, error) {
	results := make([]*Result, e.numRuns)
	errs := make([]error, e.numRuns)

	var wg sync.WaitGroup
	for i := 0; i < e.numRuns; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			cfgCopy := cfg
			cfgCopy.Seed = e.seedStart + int64(idx)

			sim := New(e.base.dyn, e.base.integrator, e.base.controller)
			for _, m := range e.base.metrics {
				sim.AddMetric(m)
			}

			results[idx], errs[idx] = sim.Run(ctx, x0, cfgCopy)
		}(i)
	}

	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
