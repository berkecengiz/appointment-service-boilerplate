// API Key Generator for hms-service
//
// # AUTHENTICATION
//
// Protected endpoints require X-API-Key header. The middleware validates the key
// against configured service keys.
//
// GENERATE API KEYS
//
//	go run cmd/keygen/main.go -service <name>
//
// Options:
//
//	-service  service name (required)
//	-length   key length in bytes (default: 32)
//	-count    number of keys (default: 1)
//	-env      output in .env format
//
// # CONFIGURATION
//
// Add to .env:
//
//	API_KEYS=service1:key1,service2:key2,service3:key3
//
// Format: serviceName:apiKey pairs, comma-separated.
//
// USAGE
//
//	curl -H "X-API-Key: your-key-here" http://localhost:8080/appointments
//
// RESPONSES
//
//	401 - Missing or invalid API key
//	200/201 - Valid key, request processed
//
// # KEY ROTATION
//
// Update API_KEYS environment variable and restart service.
package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	serviceName := flag.String("service", "", "Service name (required)")
	length := flag.Int("length", 32, "API key length in bytes (default: 32)")
	count := flag.Int("count", 1, "Number of keys to generate (default: 1)")
	envFormat := flag.Bool("env", false, "Output in .env format")
	flag.Parse()

	if *serviceName == "" {
		fmt.Println("Usage: keygen -service <service-name> [-length 32] [-count 1] [-env]")
		fmt.Println("\nExamples:")
		fmt.Println("  keygen -service appointment-ui")
		fmt.Println("  keygen -service customer-portal -length 48")
		fmt.Println("  keygen -service admin-dashboard -count 3")
		fmt.Println("  keygen -service app1 -env")
		os.Exit(1)
	}

	if *length < 16 {
		fmt.Println("Error: length must be at least 16 bytes for security")
		os.Exit(1)
	}

	var keys []string
	for i := 0; i < *count; i++ {
		key, err := generateAPIKey(*length)
		if err != nil {
			fmt.Printf("Error generating key: %v\n", err)
			os.Exit(1)
		}
		keys = append(keys, key)
	}

	if *envFormat {
		// Output in .env format
		fmt.Printf("# API keys for service: %s\n", *serviceName)
		fmt.Printf("API_KEYS=%s:%s", *serviceName, keys[0])
		for i := 1; i < len(keys); i++ {
			fmt.Printf(",%s-%d:%s", *serviceName, i+1, keys[i])
		}
		fmt.Println()
	} else {
		// Output in readable format
		fmt.Printf("Service: %s\n", *serviceName)
		fmt.Println(strings.Repeat("-", 60))
		for i, key := range keys {
			if *count > 1 {
				fmt.Printf("Key %d: %s\n", i+1, key)
			} else {
				fmt.Printf("API Key: %s\n", key)
			}
		}
		fmt.Println(strings.Repeat("-", 60))
		fmt.Printf("\n.env format:\n")
		fmt.Printf("API_KEYS=%s:%s", *serviceName, keys[0])
		for i := 1; i < len(keys); i++ {
			fmt.Printf(",%s-%d:%s", *serviceName, i+1, keys[i])
		}
		fmt.Println()
	}
}

func generateAPIKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Use URL-safe base64 encoding and remove padding for cleaner keys
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
