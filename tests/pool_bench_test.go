package tests

import (
	"context"
	"testing"
	"time"

	"github.com/sirioneto/go-worker-pool/work"
)

func BenchmarkWorkerPool(b *testing.B) {
	pool, _ := work.NewPool(50, 100)
	ctx := context.Background()

	go func() {
		pool.Start(ctx)

		job := work.NewJobWithResult(func() (any, error) {
			time.Sleep(10 * time.Millisecond)
			return "ok", nil
		}, func(err error) {})

		for i := 0; i < b.N; i++ {
			pool.AddJob(job)
		}
	}()

	pool.WaitJobs()
	pool.Stop()
}
