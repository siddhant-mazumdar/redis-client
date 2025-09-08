package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

const (
	rootURL = "http://localhost:10001"
	baseURL = rootURL + "/apis/redis"
)

func main() {
	// Wait for server to start
	fmt.Println("Waiting for server to start...")
	time.Sleep(2 * time.Second)

	runAllTests()
}

func runAllTests() {
	fmt.Println("\n=== Health Checks ===")
	testHealthCheck()

	fmt.Println("\n=== RESP Parsing ===")
	testParseResp()

	fmt.Println("\n=== Parse and Store (SET) ===")
	testParseAndStore("test_key", "test_value")

	fmt.Println("\n=== Get Data (multiple request styles) ===")
	testGetDataAll("test_key")

	fmt.Println("\n=== HMSET / HMGET / HMGET-MULTI ===")
	testHashOps()

	fmt.Println("\n=== Command History ===")
	testCommandHistory()

	fmt.Println("\n=== Delete Data (multiple request styles) ===")
	testDeleteDataAll("test_key")
}

func testHealthCheck() {
	// These endpoints are not behind the SERVICE_ENDPOINT prefix
	for _, path := range []string{"/health", "/health-check", "/"} {
		status, body := doGet(rootURL + path)
		fmt.Printf("GET %s -> %d: %s\n", path, status, body)
	}
}

func testParseResp() {
	payload := map[string]any{
		"resp_data": "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
	}
	status, body := doPost(baseURL+"/v1/redis/parse-resp", payload)
	fmt.Printf("POST /v1/redis/parse-resp -> %d: %s\n", status, body)
}

func testParseAndStore(key, val string) {
	respStr := fmt.Sprintf("*3\r\n$3\r\nSET\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(key), key, len(val), val)
	payload := map[string]any{"resp_data": respStr}
	status, body := doPost(baseURL+"/v1/redis/parse-and-store", payload)
	fmt.Printf("POST /v1/redis/parse-and-store -> %d: %s\n", status, body)
}

func testGetDataAll(key string) {
	// Try path-param style (if supported)
	status1, body1 := doGet(baseURL + "/v1/redis/get-data/" + url.PathEscape(key))
	fmt.Printf("GET /v1/redis/get-data/%s -> %d: %s\n", key, status1, body1)

	// Try POST with key_name
	status2, body2 := doPost(baseURL+"/v1/redis/get-data", map[string]any{"key_name": key})
	fmt.Printf("POST /v1/redis/get-data {key_name} -> %d: %s\n", status2, body2)

	// Try POST with key (alternative payload some clients used)
	status3, body3 := doPost(baseURL+"/v1/redis/get-data", map[string]any{"key": key})
	fmt.Printf("POST /v1/redis/get-data {key} -> %d: %s\n", status3, body3)
}

func testDeleteDataAll(key string) {
	// Try path-param style (if supported)
	status1, body1 := doGet(baseURL + "/v1/redis/delete-data/" + url.PathEscape(key))
	fmt.Printf("GET /v1/redis/delete-data/%s -> %d: %s\n", key, status1, body1)

	// Try POST with key_name
	status2, body2 := doPost(baseURL+"/v1/redis/delete-data", map[string]any{"key_name": key})
	fmt.Printf("POST /v1/redis/delete-data {key_name} -> %d: %s\n", status2, body2)

	// Verify deletion by attempting GET again
	status3, body3 := doPost(baseURL+"/v1/redis/get-data", map[string]any{"key_name": key})
	fmt.Printf("POST /v1/redis/get-data after delete -> %d: %s\n", status3, body3)
}

func testCommandHistory() {
	status1, body1 := doGet(baseURL + "/v1/redis/command-history")
	fmt.Printf("GET /v1/redis/command-history -> %d: %s\n", status1, body1)

	status2, body2 := doGet(baseURL + "/v1/redis/command-history?limit=5")
	fmt.Printf("GET /v1/redis/command-history?limit=5 -> %d: %s\n", status2, body2)
}

func testHashOps() {
	// HMSET
	status1, body1 := doPost(baseURL+"/v1/redis/hmset", map[string]any{
		"key":    "hkey1",
		"fields": map[string]string{"f1": "v1", "f2": "v2"},
	})
	fmt.Printf("POST /v1/redis/hmset -> %d: %s\n", status1, body1)

	// HMGET
	status2, body2 := doPost(baseURL+"/v1/redis/hmget", map[string]any{
		"key":    "hkey1",
		"fields": []string{"f1", "f2", "f3"},
	})
	fmt.Printf("POST /v1/redis/hmget -> %d: %s\n", status2, body2)

	// HMGET-MULTI
	status3, body3 := doPost(baseURL+"/v1/redis/hmget-multi", map[string]any{
		"keys": map[string][]string{
			"hkey1": {"f1", "f2"},
			"hkey2": {"x"},
		},
	})
	fmt.Printf("POST /v1/redis/hmget-multi -> %d: %s\n", status3, body3)
}

// Helpers
func doPost(u string, payload any) (int, string) {
	b, _ := json.Marshal(payload)
	resp, err := http.Post(u, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Printf("POST %s error: %v", u, err)
		return 0, err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}

func doGet(u string) (int, string) {
	resp, err := http.Get(u)
	if err != nil {
		log.Printf("GET %s error: %v", u, err)
		return 0, err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body)
}
