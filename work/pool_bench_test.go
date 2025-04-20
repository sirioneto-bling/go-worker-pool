package work

import (
	"context"
	"testing"
	"time"
)

func BenchmarkWorkerPool(b *testing.B) {
	benchmarks := []struct {
		name       string
		numWorkers int
		bufferSize int
	}{
		{"Small_Pool", 10, 20},
		{"Medium_Pool", 50, 100},
		{"Large_Pool", 100, 200},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			pool, _ := NewPool(bm.numWorkers, bm.bufferSize)
			ctx := context.Background()

			b.ResetTimer()
			go func() {
				pool.Start(ctx)
				job := NewJobWithResult(func() (any, error) {
					time.Sleep(10 * time.Millisecond)
					return "ok", nil
				}, func(err error) {})

				for i := 0; i < b.N; i++ {
					pool.AddJob(job)
				}
			}()

			pool.WaitJobs()
			pool.Stop()
		})
	}
}
