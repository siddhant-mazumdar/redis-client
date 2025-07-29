package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const baseURL = "http://localhost:10001/apis/redis"

func main() {
	// Wait for server to start
	fmt.Println("Waiting for server to start...")
	time.Sleep(2 * time.Second)

	// Test health check
	// testHealthCheck()

	// Test Redis operations
	testRedisOperations()
}

// func testHealthCheck() {
// 	fmt.Println("\n=== Testing Health Check ===")

// 	resp, err := http.Get(baseURL + "/health")
// 	if err != nil {
// 		log.Printf("Health check failed: %v", err)
// 		return
// 	}
// 	defer resp.Body.Close()

// 	body, _ := io.ReadAll(resp.Body)
// 	fmt.Printf("Health check response: %s\n", string(body))
// }

func testRedisOperations() {
	fmt.Println("\n=== Testing Redis Operations ===")

	// Test SET operation using parse-and-store endpoint
	testData := map[string]interface{}{
		"resp_data": "*3\r\n$3\r\nSET\r\n$8\r\ntest_key\r\n$10\r\ntest_value\r\n",
	}

	jsonData, _ := json.Marshal(testData)
	resp, err := http.Post(baseURL+"/v1/redis/parse-and-store", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("SET command failed: %v", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("SET response: %s\n", string(body))

	// Test GET operation using get-data endpoint
	getData := map[string]interface{}{
		"key": "test_key",
	}

	getJsonData, _ := json.Marshal(getData)
	getResp, err := http.Post(baseURL+"/v1/redis/get-data", "application/json", bytes.NewBuffer(getJsonData))
	if err != nil {
		log.Printf("GET command failed: %v", err)
		return
	}
	defer getResp.Body.Close()

	getBody, _ := io.ReadAll(getResp.Body)
	fmt.Printf("GET response: %s\n", string(getBody))
}
