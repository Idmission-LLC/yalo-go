package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	yalo "github.com/Idmission-LLC/yalo-go"
	"github.com/spf13/viper"
)

func init() {
	// Set the file name of the configurations file
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")

	// Enable reading from environment variables
	viper.AutomaticEnv()

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Could not read config file: %v", err)
		log.Println("Using environment variables or defaults")
	}
}

func main() {
	// Check for command line arguments
	if len(os.Args) < 4 {
		log.Fatal("Usage: go run main.go <notification_type> <phone_number> <json_params> [priority]")
	}

	notificationType := os.Args[1]
	phoneNumber := os.Args[2]
	jsonParams := os.Args[3]
	priority := "1"

	var notificationOptions []yalo.NotificationOption
	if len(os.Args) >= 5 {
		priority = os.Args[4]
		notificationOptions = append(notificationOptions, yalo.WithPriority(priority))
	}

	// Load configuration from .env file or environment variables
	accountID := viper.GetString("YALO_ACCOUNT_ID")
	botID := viper.GetString("YALO_BOT_ID")
	token := viper.GetString("YALO_TOKEN")
	debug := viper.GetBool("YALO_DEBUG")

	// Set default base URL if not provided
	viper.SetDefault("YALO_BASE_URL", "https://api-global.yalochat.com")
	baseURL := viper.GetString("YALO_BASE_URL")

	// Validate required configuration
	if accountID == "" {
		log.Fatal("YALO_ACCOUNT_ID is required")
	}
	if botID == "" {
		log.Fatal("YALO_BOT_ID is required")
	}
	if token == "" {
		log.Fatal("YALO_TOKEN is required")
	}

	// Initialize the Yalo client
	client := yalo.NewClient(
		yalo.WithBaseURL(baseURL),
		yalo.WithAccount(accountID, botID),
		yalo.WithToken(token),
		yalo.WithDebug(debug),
	)

	if debug {
		fmt.Println("=== Debug mode enabled ===")
	}

	// Parse JSON params from command line
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(jsonParams), &params); err != nil {
		log.Fatalf("Error parsing JSON params: %v", err)
	}

	// Send the notification
	fmt.Println("=== Sending Notification ===")
	fmt.Printf("Notification Type: %s\n", notificationType)
	fmt.Printf("Phone: %s\n", phoneNumber)
	fmt.Printf("Params: %s\n", jsonParams)
	fmt.Printf("Priority: %s\n", priority)

	result, err := client.SendNotification(
		context.Background(),
		notificationType,
		phoneNumber,
		params,
		notificationOptions...,
	)
	if err != nil {
		log.Fatalf("Error sending notification: %v", err)
	}

	fmt.Printf("\nSuccess: %t\n", result.Success)
	fmt.Printf("ID: %s\n", result.ID)
	fmt.Printf("Message IDs: %v\n", result.MessageIDs)
	prettyBytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("\nFull Response:")
	fmt.Println(string(prettyBytes))
}
