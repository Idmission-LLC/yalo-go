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
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run main.go <phone_number> <json_params> [priority]")
	}

	phoneNumber := os.Args[1]
	jsonParams := os.Args[2]
	priority := "1"

	var notificationOptions []yalo.NotificationOption
	if len(os.Args) >= 4 {
		priority = os.Args[3]
		notificationOptions = append(notificationOptions, yalo.WithPriority(priority))
	}

	// Load configuration from .env file or environment variables
	accountID := viper.GetString("YALO_ACCOUNT_ID")
	botID := viper.GetString("YALO_BOT_ID")
	token := viper.GetString("YALO_TOKEN")
	notificationType := viper.GetString("YALO_NOTIFICATION_TYPE")
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
	if notificationType == "" {
		log.Fatal("YALO_NOTIFICATION_TYPE is required")
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

	fmt.Printf("\nStatus: %s\n", result.Status)
	fmt.Printf("Message: %s\n", result.Message)
	prettyBytes, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println("\nFull Response:")
	fmt.Println(string(prettyBytes))

}
