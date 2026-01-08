package dynamo

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

			s := New(e.base.dyn, e.base.integrator, e.base.controller)
			for _, m := range e.base.metrics {
				s.AddMetric(m)
			}

			results[idx], errs[idx] = s.Run(ctx, x0, cfgCopy)
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

// ParallelFor executes a function in parallel over a range [0, n)
func ParallelFor(n, minChunk int, fn func(start, end int)) {
	numWorkers := 4 // Default
	if n <= minChunk || numWorkers <= 1 {
		fn(0, n)
		return
	}

	workers := numWorkers
	if n/minChunk < workers {
		workers = n / minChunk
	}
	if workers < 1 {
		workers = 1
	}

	chunkSize := (n + workers - 1) / workers

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		start := w * chunkSize
		end := start + chunkSize
		if end > n {
			end = n
		}

		go func(s, e int) {
			defer wg.Done()
			fn(s, e)
		}(start, end)
	}

	wg.Wait()
}
