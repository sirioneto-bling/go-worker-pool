package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sirioneto/go-worker-pool/work"
)

func main() {
	start := time.Now()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := work.NewPool(50, 200)
	if err != nil {
		log.Fatal(err)
	}

	logger := log.Default()

	go func() {
		jobWithResult := work.NewJobWithResult(sendRequestWithResult, handleRequestError)
		pool.Start(ctx)
		for range 500 {
			pool.AddJob(jobWithResult)
		}

		pool.WaitJobs()
		logger.Println("stopping pool after results")
		pool.Stop()
	}()

	totalJobsProcessed, jobsWithError := 0, 0

	for {
		select {
		case <-ctx.Done():
			return
		case <-pool.Done():
			logger.Println("total jobs processed:", totalJobsProcessed)
			logger.Println("jobs processed succefully:", totalJobsProcessed-jobsWithError)
			logger.Println("jobs processed with error:", jobsWithError)
			logger.Println("execution time:", time.Since(start).Seconds(), "segundos")
			return
		default:
			for {
				result, ok := <-pool.Result()
				if !ok {
					break
				}

				totalJobsProcessed++

				value := fmt.Sprintf("[SUCCESS] %s", result.Value)
				if result.Error != nil {
					jobsWithError++
					value = fmt.Sprintf("[ERROR] %s", result.Error.Error())
				}
				serializedResult := fmt.Sprintf("result=%v\n", value)
				logger.Println(serializedResult)
			}
		}
	}
}

func sendRequestWithResult() (any, error) {
	const url = "https://www.google.com/"
	time.Sleep(1 * time.Second) // simulate process load

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("%s returned status code %d \n", url, res.StatusCode)
	return result, nil
}

func handleRequestError(err error) {
	logger := log.Default()
	logger.Println("request error:", err.Error())
}
