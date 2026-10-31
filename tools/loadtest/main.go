package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	baseURL      = "http://localhost:3000"
	supervisorID = 3
	requestCount = 30
	firstStudent = 6
)

type result struct {
	status int
	err    error
}

func main() {
	token := os.Getenv("LOADTEST_TOKEN")
	if token == "" {
		fmt.Println("ERROR: environment variable LOADTEST_TOKEN belum diisi.")
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	start := make(chan struct{})
	results := make(chan result, requestCount)

	var wg sync.WaitGroup

	for i := 0; i < requestCount; i++ {
		studentID := firstStudent + i

		wg.Add(1)
		go func(studentID int) {
			defer wg.Done()
			<-start

			body, _ := json.Marshal(map[string]int{
				"student_id": studentID,
			})

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf("%s/api/v1/supervisors/%d/assign",
					baseURL, supervisorID),
				bytes.NewReader(body),
			)
			if err != nil {
				results <- result{err: err}
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := client.Do(req)
			if err != nil {
				results <- result{err: err}
				return
			}
			defer resp.Body.Close()

			results <- result{status: resp.StatusCode}
		}(studentID)
	}

	close(start)
	wg.Wait()
	close(results)

	counts := map[int]int{}
	errorsCount := 0

	for r := range results {
		if r.err != nil {
			errorsCount++
			fmt.Printf("NETWORK ERROR: %v\n", r.err)
			continue
		}

		counts[r.status]++
		fmt.Printf("HTTP status: %d\n", r.status)
	}

	fmt.Println("\n=== HASIL LOAD TEST D.3 ===")
	fmt.Printf("Total request : %d\n", requestCount)
	fmt.Printf("HTTP 201      : %d\n", counts[201])
	fmt.Printf("HTTP 409      : %d\n", counts[409])
	fmt.Printf("HTTP lainnya  : %d\n", requestCount-counts[201]-counts[409]-errorsCount)
	fmt.Printf("Network error : %d\n", errorsCount)

	if counts[201] == 3 &&
		counts[409] == 27 &&
		errorsCount == 0 {
		fmt.Println("PASS: tepat 3 penetapan berhasil.")
	} else {
		fmt.Println("FAIL: hasil tidak sesuai harapan 3 sukses dan 27 konflik.")
		os.Exit(1)
	}
}
