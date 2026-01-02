package main

import (
	"autoglm-go/constants"
	"autoglm-go/phoneagent"
	"autoglm-go/phoneagent/definitions"
	"autoglm-go/utils"
	"bufio"
	"context"
	"fmt"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Config holds all the configuration values from command line arguments
type Config struct {
	BaseURL     string `json:"base_url"`
	Model       string `json:"model"`
	APIKey      string `json:"api_key"`
	MaxSteps    int    `json:"max_steps"`
	DeviceID    string `json:"device_id"`
	Connect     string `json:"connect"`
	Disconnect  string `json:"disconnect"`
	ListDevices bool   `json:"list_devices"`
	EnableTCPIP int    `json:"enable_tcpip"`
	GetDeviceIP string `json:"get_device_ip"`

	WdaUrl     string `json:"wda_url"`
	Pair       bool   `json:"pair"`
	WdaStatus  bool   `json:"wda_status"`
	Quiet      bool   `json:"quiet"`
	ListApps   bool   `json:"list_apps"`
	Lang       string `json:"lang"`
	DeviceType string `json:"device_type"`
	Task       string `json:"task"`
}

var rootCmd = &cobra.Command{
	Use:   "main",
	Short: "Phone Agent - AI-powered phone automation",
	Long: `Phone Agent is an AI-powered tool for automating phone operations.
It supports Android devices via ADB, and iOS devices via WebDriverAgent.`,
	Example: `  # Run with default settings (Android)
  go run main.go

  # Specify model endpoint
  go run main.go --base-url http://localhost:8000/v1

  # Use API key for authentication
  go run main.go --apikey sk-xxxxx

  # Run with specific device
  go run main.go --device-id emulator-5554

  # Connect to remote device
  go run main.go --connect 192.168.1.100:5555

  # List connected devices
  go run main.go --list-devices

  # Enable TCP/IP on USB device and get connection info
  go run main.go --enable-tcpip

  # List supported apps
  go run main.go --list-apps

  # iOS specific examples
  # Run with iOS device
  go run main.go --device-type ios "Open Safari and search for iPhone tips"

  # Use WiFi connection for iOS
  go run main.go --device-type ios --wda-url http://192.168.1.100:8100

  # List connected iOS devices
  go run main.go --device-type ios --list-devices

  # Check WebDriverAgent status
  go run main.go --device-type ios --wda-status

  # Pair with iOS device
  go run main.go --device-type ios --pair`,
	Run: func(cmd *cobra.Command, args []string) {
		// Store the task as a space-joined string
		if len(args) > 0 {
			config.Task = strings.Join(args, " ")
		}

		fmt.Printf("Configuration: %s\n", utils.JsonIndent(config))
		fmt.Printf("Task: %s\n", config.Task)
	},
}

var config = &Config{}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Helper function to get environment variable as int with default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// Helper function to get environment variable as float32 with default value
func getEnvFloat32(key string, defaultValue float32) float32 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 32); err == nil {
			return float32(floatValue)
		}
	}
	return defaultValue
}

func init() {
	// Model options
	rootCmd.PersistentFlags().StringVar(&config.BaseURL, "base-url",
		getEnv("PHONE_AGENT_BASE_URL", "http://localhost:8000/v1"),
		"Model API base URL")

	rootCmd.PersistentFlags().StringVar(&config.Model, "model",
		getEnv("PHONE_AGENT_MODEL", "autoglm-phone"),
		"Model name")

	rootCmd.PersistentFlags().StringVar(&config.APIKey, "apikey",
		getEnv("PHONE_AGENT_API_KEY", "EMPTY"),
		"API key for model authentication")

	rootCmd.PersistentFlags().IntVar(&config.MaxSteps, "max-steps",
		getEnvInt("PHONE_AGENT_MAX_STEPS", 100),
		"Maximum steps per task")

	// Device options
	rootCmd.PersistentFlags().StringVarP(&config.DeviceID, "device-id", "d",
		getEnv("PHONE_AGENT_DEVICE_ID", ""),
		"ADB device ID")

	rootCmd.PersistentFlags().StringVarP(&config.Connect, "connect", "c", "",
		"Connect to remote device (e.g., 192.168.1.100:5555)")

	rootCmd.PersistentFlags().StringVar(&config.Disconnect, "disconnect", "",
		"Disconnect from remote device (or 'all' to disconnect all)")

	rootCmd.PersistentFlags().BoolVar(&config.ListDevices, "list-devices", false,
		"List connected devices and exit")

	// For enable-tcpip, we need custom handling to support optional argument
	rootCmd.PersistentFlags().IntVar(&config.EnableTCPIP, "enable-tcpip", 0,
		"Enable TCP/IP debugging on USB device (default port: 5555, use 0 for default)")

	rootCmd.PersistentFlags().StringVar(&config.GetDeviceIP, "get-device-ip", "",
		"Get device IP ")

	// iOS specific options
	rootCmd.PersistentFlags().StringVar(&config.WdaUrl, "wda-url",
		getEnv("PHONE_AGENT_WDA_URL", "http://localhost:8100"),
		"WebDriverAgent URL for iOS (default: http://localhost:8100)")

	rootCmd.PersistentFlags().BoolVar(&config.Pair, "pair", false,
		"Pair with iOS device (required for some operations)")

	rootCmd.PersistentFlags().BoolVar(&config.WdaStatus, "wda-status", false,
		"Show WebDriverAgent status and exit (iOS only)")

	// Other options
	rootCmd.PersistentFlags().BoolVarP(&config.Quiet, "quiet", "q", false,
		"Suppress verbose output")

	rootCmd.PersistentFlags().BoolVar(&config.ListApps, "list-apps", false,
		"List supported apps and exit")

	rootCmd.PersistentFlags().StringVar(&config.Lang,
		"lang",
		getEnv("PHONE_AGENT_LANG", "cn"),
		"Language for system prompt (cn or en, default: cn)")

	rootCmd.PersistentFlags().StringVar(
		&config.DeviceType,
		"device-type",
		"adb",
		"Device type: adb for Android, ios for iPhone (default: adb)",
	)
}

func main() {
	parseArgs()

	ctx := context.Background()

	// Handle --list-apps (no system check needed)
	if config.ListApps {
		var supportedApps []string
		if config.DeviceType == constants.IOS {
			fmt.Println("Supported iOS apps:")
			fmt.Println("\nNote: For iOS apps, Bundle IDs are configured in:")
			fmt.Println("  constants/apps.go")
			fmt.Println("\nCurrently configured apps:")
			supportedApps = lo.Keys(constants.APP_PACKAGES_IOS)
		} else {
			fmt.Println("Supported Android apps:")
			supportedApps = lo.Keys(constants.APP_PACKAGES_ANDROID)
		}
		sort.Strings(supportedApps)

		for _, app := range supportedApps {
			fmt.Printf("  - %s\n", app)
		}
		return
	}

	device, err := phoneagent.CreateDevice(config.DeviceType)
	if err != nil {
		fmt.Printf("creating device failed, err: %v\n", err)
		return
	}

	// Handle device commands (these may need partial system checks)
	if err := handleDeviceCommands(ctx, device); err != nil {
		fmt.Printf("✗ %s\n", "handle device commands failed")
		return
	}
	if passed := checkSystemRequirements(ctx, config.DeviceType, config.WdaUrl); !passed {
		fmt.Printf("✗ %s\n", "check system requirements failed")
		return
	}

	if passed := checkModelAPI(ctx, config.BaseURL, config.Model, config.APIKey); !passed {
		fmt.Printf("✗ %s\n", "check model api failed")
		return
	}

	modelConfig := &definitions.ModelConfig{
		BaseURL:          config.BaseURL,
		ModelName:        config.Model,
		APIKey:           config.APIKey,
		Lang:             config.Lang,
		MaxTokens:        getEnvInt("PHONE_AGENT_MAX_TOKENS", 3000),
		Temperature:      getEnvFloat32("PHONE_AGENT_TEMPERATURE", 0.0),
		TopP:             getEnvFloat32("PHONE_AGENT_TOP_P", 0.85),
		FrequencyPenalty: getEnvFloat32("PHONE_AGENT_FREQUENCY_PENALTY", 0.2),
	}
	agentConfig := &definitions.AgentConfig{
		MaxSteps: config.MaxSteps,
		DeviceID: config.DeviceID,
		Lang:     config.Lang,
		Verbose:  !config.Quiet,
		WdaUrl:   config.WdaUrl,
	}

	phoneAgent := phoneagent.NewPhoneAgent(device, modelConfig, agentConfig)

	// Print configuration information
	printConfiguration(ctx, phoneAgent)

	// Run with provided task or enter interactive mode
	if config.Task != "" {
		fmt.Printf("\nTask: %s\n", config.Task)
		result, err := phoneAgent.Run(ctx, config.Task)
		if err != nil {
			fmt.Printf("Error running task: %v\n", err)
			return
		}
		fmt.Printf("\nResult: %s\n", result)
	} else {
		// Interactive mode
		fmt.Println("\nEntering interactive mode. Type 'quit' to exit.")

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter your task: ")
			task, err := reader.ReadString('\n')
			if err != nil {
				fmt.Printf("\nError reading input: %v\n", err)
				continue
			}

			task = strings.TrimSpace(task)

			// Check for quit commands
			if strings.ToLower(task) == "quit" || strings.ToLower(task) == "exit" || strings.ToLower(task) == "q" {
				fmt.Println("Goodbye!")
				break
			}

			// Skip empty tasks
			if task == "" {
				continue
			}

			fmt.Println()
			result, err := phoneAgent.Run(ctx, task)
			if err != nil {
				fmt.Printf("\nError: %v\n", err)
				continue
			}
			fmt.Printf("\nResult: %s\n", result)

			// Reset agent for next task
			phoneAgent.Reset(ctx)
		}
	}

}

func parseArgs() *Config {
	// Set pre-run validation
	rootCmd.PersistentPreRunE = validateArgs

	// Execute the command
	cobra.CheckErr(rootCmd.Execute())

	return config
}
func validateArgs(cmd *cobra.Command, args []string) error {
	// Handle enable-tcpip with optional argument
	if cmd.Flags().Changed("enable-tcpip") && config.EnableTCPIP == 0 {
		config.EnableTCPIP = 5555 // Default port
	}

	// Validate lang and device-type choices
	if config.Lang != "cn" && config.Lang != "en" {
		return fmt.Errorf("invalid language option: %s. Must be 'cn' or 'en'", config.Lang)
	}

	if config.DeviceType != "adb" && config.DeviceType != "ios" {
		return fmt.Errorf("invalid device type: %s. Must be 'adb' or 'ios'", config.DeviceType)
	}

	return nil
}

func handleDeviceCommands(ctx context.Context, device phoneagent.Device) error {
	deviceType := config.DeviceType

	// 处理iOS特定命令
	if deviceType == constants.IOS {
		return handleIOSDeviceCommands(ctx)
	}

	// 处理 --list-devices
	if config.ListDevices {
		devices, err := device.ListDevices(ctx)
		if err != nil {
			fmt.Printf("listing devices failed, err: %v\n", err)
			return err
		}
		if len(devices) == 0 {
			fmt.Println("No devices connected.")
		} else {
			fmt.Println("Connected devices:")
			fmt.Println(strings.Repeat("-", 60))
			for _, d := range devices {
				statusIcon := "✓"
				if d.Status != "device" {
					statusIcon = "✗"
				}
				connType := d.ConnectionType
				modelInfo := ""
				if d.Model != "" {
					modelInfo = fmt.Sprintf(" (%s)", d.Model)
				}
				fmt.Printf("  %s %-30s [%s]%s\n", statusIcon, d.DeviceID, connType, modelInfo)
			}
		}
		return nil
	}

	// 处理 --connect
	if config.Connect != "" {
		fmt.Printf("Connecting to %s...\n", config.Connect)
		message, err := device.Connect(ctx, config.Connect)
		if err != nil {
			fmt.Printf("✗ %s\n", message)
			return err
		}
		fmt.Printf("✓ %s\n", message)
		return nil
	}

	// 处理 --enable-tcpip
	if config.EnableTCPIP > 0 {
		port := config.EnableTCPIP
		fmt.Printf("Enabling TCP/IP debugging on port %d...\n", port)

		err := device.EnableTCPIP(ctx, port, config.DeviceID)
		if err != nil {
			fmt.Printf("✗ %s\n", "enable tcpip failed")
			return err
		}
		fmt.Printf("✓ %s\n", "enable tcpip success")
		return nil
	}

	// 处理 --get-device-ip
	if len(config.GetDeviceIP) > 0 {
		ip, err := device.GetDeviceIP(ctx, config.GetDeviceIP)
		if err != nil {
			fmt.Printf("✗ %s\n", "get device ip failed")
			return err
		}
		fmt.Printf("✓ %s: %s\n", "device ip", ip)
		return nil
	}

	// 处理 --disconnect
	if config.Disconnect != "" {
		var (
			message string
			err     error
		)

		if config.Disconnect == "all" {
			fmt.Println("Disconnecting all remote devices...")
			message, err = device.Disconnect(ctx, "") // 断开所有连接
		} else {
			fmt.Printf("Disconnecting from %s...\n", config.Disconnect)
			message, err = device.Disconnect(ctx, config.Disconnect)
		}
		fmt.Printf("Disconnecting result: %s\n", message)

		var statusSymbol string
		if err != nil {
			statusSymbol = "✗"
		} else {
			statusSymbol = "✓"
		}
		fmt.Printf("%s %s\n", statusSymbol, message)
		return err
	}

	return nil
}

func handleIOSDeviceCommands(ctx context.Context) error {
	// todo
	return nil
}
func checkSystemRequirements(ctx context.Context, deviceType string, wdaURL string) bool {
	fmt.Println("🔍 Checking system requirements...")
	fmt.Println(strings.Repeat("-", 50))

	allPassed := true

	// Determine tool name and command
	var toolName, toolCmd string
	if deviceType == constants.ADB {
		toolName = "ADB"
		toolCmd = "adb"
	} else {
		toolName = "libimobiledevice"
		toolCmd = "idevice_id"
	}

	// Check 1: Tool installed
	fmt.Printf("1. Checking %s installation... ", toolName)
	_, err := exec.LookPath(toolCmd)
	if err != nil {
		fmt.Println("❌ FAILED")
		fmt.Printf("   Error: %s is not installed or not in PATH.\n", toolName)
		fmt.Printf("   Solution: Install %s:\n", toolName)
		if deviceType == constants.ADB {
			fmt.Println("     - macOS: brew install android-platform-tools")
			fmt.Println("     - Linux: sudo apt install android-tools-adb")
			fmt.Println("     - Windows: Download from https://developer.android.com/studio/releases/platform-tools")
		} else { // IOS
			fmt.Println("     - macOS: brew install libimobiledevice")
			fmt.Println("     - Linux: sudo apt-get install libimobiledevice-utils")
		}
		allPassed = false
	} else {
		// Double check by running version command
		var versionCmd *exec.Cmd
		if deviceType == constants.ADB {
			versionCmd = exec.Command(toolCmd, "version")
		} else { // IOS
			versionCmd = exec.Command(toolCmd, "-ln")
		}

		output, err := versionCmd.Output()
		if err != nil {
			fmt.Println("❌ FAILED")
			fmt.Printf("   Error: %s command failed to run: %v\n", toolName, err)
			allPassed = false
		}
		lines := strings.Split(string(output), "\n")
		versionLine := ""
		if len(lines) > 0 {
			versionLine = strings.TrimSpace(lines[0])
		}
		if versionLine == "" {
			versionLine = "installed"
		}
		fmt.Printf("✅ OK (%s)\n", versionLine)
	}

	// If tool is not installed, skip remaining checks
	if !allPassed {
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println("❌ System check failed. Please fix the issues above.")
		return false
	}

	// Check 2: Device connected
	fmt.Print("2. Checking connected devices... ")
	var devices []string
	var deviceIDs []string

	if deviceType == constants.ADB {
		cmd := exec.Command("adb", "devices")
		output, err := cmd.CombinedOutput() // 捕获 stdout + stderr
		if err != nil {
			fmt.Println("❌ FAILED")
			fmt.Printf("   Error: %s command timed out: %v\n", toolName, err)
			allPassed = false
		}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines[1:] { // 跳过第一行（header）
			line = strings.TrimSpace(line)
			if line != "" && strings.Contains(line, "\tdevice") {
				devices = append(devices, line)
				parts := strings.Split(line, "\t")
				if len(parts) > 0 {
					deviceIDs = append(deviceIDs, parts[0])
				}
			}
		}
	} else { // IOS
		// todo
	}

	if len(devices) == 0 {
		fmt.Println("❌ FAILED")
		fmt.Println("   Error: No devices connected.")
		fmt.Println("   Solution:")
		if deviceType == constants.ADB {
			fmt.Println("     1. Enable USB debugging on your Android device")
			fmt.Println("     2. Connect via USB and authorize the connection")
			fmt.Println("     3. Or connect remotely: go run main.go --connect <ip>:<port>")
		} else { // IOS
			fmt.Println("     1. Connect your iOS device via USB")
			fmt.Println("     2. Unlock device and tap 'Trust This Computer'")
			fmt.Println("     3. Verify: idevice_id -l")
			fmt.Println("     4. Or connect via WiFi using device IP")
		}
		allPassed = false
	} else {
		displayIDs := deviceIDs
		if len(displayIDs) > 2 {
			displayIDs = deviceIDs[:2]
			displayIDs = append(displayIDs, "...")
		}
		fmt.Printf("✅ OK (%d device(s): %s)\n", len(devices), strings.Join(displayIDs, ", "))
	}

	// If no device connected, skip ADB Keyboard check
	if !allPassed {
		fmt.Println(strings.Repeat("-", 50))
		fmt.Println("❌ System check failed. Please fix the issues above.")
		return false
	}

	// Check 3: ADB Keyboard installed (only for ADB) or WebDriverAgent (for iOS)
	if deviceType == constants.ADB {
		fmt.Print("3. Checking ADB Keyboard... ")
		cmd := exec.Command("adb", "shell", "ime", "list", "-s")
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println("❌ FAILED")
			fmt.Println("   Error: ADB command timed out:", err)
			allPassed = false
		}

		imeList := string(output)
		if strings.Contains(imeList, "com.android.adbkeyboard/.AdbIME") {
			fmt.Println("✅ OK")
		} else {
			fmt.Println("❌ FAILED")
			fmt.Println("   Error: ADB Keyboard is not installed on the device.")
			fmt.Println("   Solution:")
			fmt.Println("     1. Download ADB Keyboard APK from:")
			fmt.Println("        https://github.com/senzhk/ADBKeyBoard/blob/master/ADBKeyboard.apk")
			fmt.Println("     2. Install it on your device: adb install ADBKeyboard.apk")
			fmt.Println("     3. Enable it in Settings > System > Languages & Input > Virtual Keyboard")
			allPassed = false
		}

	} else { // IOS
		// todo
		// // Check WebDriverAgent
		// fmt.Printf("3. Checking WebDriverAgent (%s)... ", wdaURL)
		// conn := NewXCTestConnection(wdaURL)
		// if conn.IsWDAResponsive() {
		// 	fmt.Println("✅ OK")
		// 	// Get WDA status for additional info
		// 	status, err := conn.GetWDAStatus()
		// 	if err == nil && status != nil {
		// 		sessionID, ok := status["sessionId"]
		// 		if ok {
		// 			fmt.Printf("   Session ID: %v\n", sessionID)
		// 		}
		// 	}
		// } else {
		// 	fmt.Println("❌ FAILED")
		// 	fmt.Println("   Error: WebDriverAgent is not running or not accessible.")
		// 	fmt.Println("   Solution:")
		// 	fmt.Println("     1. Run WebDriverAgent on your iOS device via Xcode")
		// 	fmt.Println("     2. For USB: Set up port forwarding: iproxy 8100 8100")
		// 	fmt.Println("     3. For WiFi: Use device IP, e.g., --wda-url http://192.168.1.100:8100")
		// 	fmt.Println("     4. Verify in browser: open http://localhost:8100/status")
		// 	allPassed = false
		// }
	}

	fmt.Println(strings.Repeat("-", 50))

	if allPassed {
		fmt.Println("✅ All system checks passed!")
	} else {
		fmt.Println("❌ System check failed. Please fix the issues above.")
	}

	return allPassed
}

// printConfiguration prints the configuration information
func printConfiguration(ctx context.Context, phoneAgent *phoneagent.PhoneAgent) {

	modelConfig := phoneAgent.ModelConfig
	agentConfig := phoneAgent.AgentConfig
	device := phoneAgent.Device

	fmt.Println(strings.Repeat("=", 50))
	if config.DeviceType == constants.ADB {
		fmt.Println("Phone Agent - AI-powered phone automation")
	} else {
		fmt.Println("Phone Agent iOS - AI-powered iOS automation")
	}

	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Model: %s\n", modelConfig.ModelName)
	fmt.Printf("Base URL: %s\n", modelConfig.BaseURL)
	fmt.Printf("Max Steps: %d\n", agentConfig.MaxSteps)
	fmt.Printf("Language: %s\n", agentConfig.Lang)
	fmt.Printf("Device Type: %s\n", strings.ToUpper(config.DeviceType))

	// Show iOS-specific config
	if config.DeviceType == constants.IOS {
		fmt.Printf("WDA URL: %s\n", config.WdaUrl)
	}

	devices, err := device.ListDevices(ctx)
	if err != nil {
		fmt.Println("❌ Failed to list devices:", err)
		return
	}

	// Show device info
	if config.DeviceType == constants.IOS {
		if agentConfig.DeviceID != "" {
			fmt.Printf("Device: %s\n", agentConfig.DeviceID)
		} else if len(devices) > 0 {
			device := devices[0]
			deviceName := device.DeviceID
			if len(deviceName) > 16 {
				deviceName = deviceName[:16]
			}
			fmt.Printf("Device: %s\n", deviceName)
			if device.Model != "" {
				fmt.Printf("        %s\n", device.Model)
			}
		}
	} else {
		if agentConfig.DeviceID != "" {
			fmt.Printf("Device: %s\n", agentConfig.DeviceID)
		} else if len(devices) > 0 {
			fmt.Printf("Device: %s (auto-detected)\n", devices[0].DeviceID)
		}
	}
	fmt.Println(strings.Repeat("=", 50))
}

func checkModelAPI(ctx context.Context, baseURL, modelName, apiKey string) bool {
	fmt.Println("🔍 Checking model API...")
	fmt.Println(strings.Repeat("-", 50))
	allPassed := true
	// Check 1: Network connectivity using chat API
	fmt.Printf("1. Checking API connectivity (%s)... ", baseURL)

	// Create OpenAI client
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)
	// Set timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	// Use chat completion to test connectivity
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: modelName,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "please return hello world",
				},
			},
			MaxCompletionTokens: 5,
			Temperature:         0,
			Stream:              false,
		},
	)
	// Check response
	if err != nil {
		fmt.Println("❌ FAILED")
		errorMsg := err.Error()
		// Provide more specific error messages
		switch {
		case strings.Contains(errorMsg, "connection refused") || strings.Contains(errorMsg, "connection error"):
			fmt.Printf("   Error: Cannot connect to %s\n", baseURL)
			fmt.Println("   Solution:")
			fmt.Println("     1. Check if the model server is running")
			fmt.Println("     2. Verify the base URL is correct")
			fmt.Printf("     3. Try: curl %s/chat/completions\n", baseURL)
		case strings.Contains(strings.ToLower(errorMsg), "timed out") || strings.Contains(errorMsg, "timeout"):
			fmt.Printf("   Error: Connection to %s timed out\n", baseURL)
			fmt.Println("   Solution:")
			fmt.Println("     1. Check your network connection")
			fmt.Println("     2. Verify the server is responding")
		case strings.Contains(errorMsg, "no such host") || strings.Contains(errorMsg, "name resolution"):
			fmt.Println("   Error: Cannot resolve hostname")
			fmt.Println("   Solution:")
			fmt.Println("     1. Check the URL is correct")
			fmt.Println("     2. Verify DNS settings")
		default:
			fmt.Printf("   Error: %s\n", errorMsg)
		}
		allPassed = false
	} else if len(resp.Choices) == 0 {
		fmt.Println("❌ FAILED")
		fmt.Println("   Error: Received empty response from API")
		allPassed = false
	} else {
		fmt.Printf("✅ OK, Response: %s\n", utils.JsonString(resp))
	}
	fmt.Println(strings.Repeat("-", 50))
	if allPassed {
		fmt.Println("✅ Model API checks passed!")
	} else {
		fmt.Println("❌ Model API check failed. Please fix the issues above.")
	}
	return allPassed
}
