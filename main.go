package main

import (
	"net/http"
	"os"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
)

// Config struct, matching config.toml file
type AppConfig struct {
	UserAgent          string
	HttpTimeout        int
	AcceptedStatuses   []int
	ConcurrentRequests int
	Endpoints          []Endpoint
}

// Set up a struct for endpoints
type Endpoint struct {
	URL        string
	HeaderHost string
}

// Set up struct for request results
type RequestResult struct {
	URL          string `json:"url"`
	ResponseCode int    `json:"response_code"`
}

// Function for getting and returning http response codes from supplied endpoint
func endpoint_caller(endpoint Endpoint, config AppConfig) RequestResult {
	client := &http.Client{Timeout: time.Duration(time.Second * time.Duration(config.HttpTimeout))}
	defer client.CloseIdleConnections()

	req, err := http.NewRequest(http.MethodGet, endpoint.URL, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Set("User-Agent", config.UserAgent)
	req.Header.Set("Host", endpoint.HeaderHost)
	resp, err := client.Do(req)
	if err != nil {
		return RequestResult{
			URL:          endpoint.URL,
			ResponseCode: 503,
		}
	}
	defer resp.Body.Close()

	return RequestResult{
		URL:          endpoint.URL,
		ResponseCode: resp.StatusCode,
	}
}

func getHealthStatus(c *gin.Context) {
	// Fetch new AppConfig each time for fresh endpoints data
	app_config := getAppConfig()
	endpoints := app_config.Endpoints
	concurrent_requests := app_config.ConcurrentRequests

	// Empty slices for storing results
	var results []RequestResult
	var failed []RequestResult

	// Iterate endpoints and collect http response codes
	queue := make(chan bool, concurrent_requests)
	for i := range endpoints {
		queue <- true
		go func() {
			defer func() { <-queue }()
			result := endpoint_caller(endpoints[i], app_config)
			if !slices.Contains(app_config.AcceptedStatuses, result.ResponseCode) {
				failed = append(failed, result)
			}
			results = append(results, result)
		}()
	}

	for i := 0; i < concurrent_requests; i++ {
		queue <- true
	}

	// Check if any endpoints listed as failed, if so return early
	if len(failed) > 0 {
		c.JSON(http.StatusFailedDependency, failed)
		return
	}

	c.JSON(http.StatusOK, results)
}

// Fetch app config from config.toml file
func getAppConfig() AppConfig {
	config_file, file_err := os.ReadFile("config.toml")
	if file_err != nil {
		panic(file_err)
	}
	var app_config AppConfig
	toml_err := toml.Unmarshal([]byte(config_file), &app_config)
	if toml_err != nil {
		panic(toml_err)
	}

	return app_config
}

// Main function running the endpoint
func main() {
	router := gin.Default()
	router.GET("/health", getHealthStatus)

	router.Run(":1337")
}
