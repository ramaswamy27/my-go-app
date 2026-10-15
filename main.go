package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents your exact pipeline-config.yaml schema
type Config struct {
	Stack        string            `yaml:"stack"`
	StackVersion string            `yaml:"stackVersion"`
	ImageName    string            `yaml:"imageName"`
	ImageVer     interface{}       `yaml:"imageVer"` // Uses interface{} to safely handle float/string formats
	Resources    map[string]string `yaml:"resources"`
	BuildCmd     string            `yaml:"buildCmd"`
	TestCmd      string            `yaml:"testCmd"`
}

func main() {
	configFile := "pipeline-config.yaml"

	// 1. Read the custom YAML file
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("❌ Error: Could not find or read %s: %v\nMake sure this file exists in the current directory.", configFile, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("❌ Error: Failed to parse YAML structure: %v", err)
	}

	// 2. Validate configuration keys
	fmt.Println("🔍 Validating Pipeline Configuration...")
	if strings.TrimSpace(config.BuildCmd) == "" || config.BuildCmd == "<required build command>" {
		log.Fatal("❌ Validation Failed: 'buildCmd' is empty or still contains the placeholder text.")
	}
	if strings.TrimSpace(config.TestCmd) == "" || config.TestCmd == "<test_command>" {
		log.Fatal("❌ Validation Failed: 'testCmd' is empty or still contains the placeholder text.")
	}

	fmt.Printf("📦 Stack: %s (%s) | Image: %s:%v\n", config.Stack, config.StackVersion, config.ImageName, config.ImageVer)
	if cpu, ok := config.Resources["cpu"]; ok {
		fmt.Printf("⚙️  Resource Allocations -> CPU: %s, Memory: %s\n", cpu, config.Resources["memory"])
	}

	// 3. Execute Build Step
	fmt.Printf("\n🔨 Running Build Stage: [%s]\n", config.BuildCmd)
	if err := runCommand(config.BuildCmd); err != nil {
		log.Fatalf("❌ Build Stage Failed: %v", err)
	}
	fmt.Println("✅ Build Completed Successfully!")

	// 4. Execute Test Step
	fmt.Printf("\n🧪 Running Test Stage: [%s]\n", config.TestCmd)
	if err := runCommand(config.TestCmd); err != nil {
		log.Fatalf("❌ Test Stage Failed: %v", err)
	}
	fmt.Println("✅ All Test Suites Passed Successfully!")
	fmt.Println("\n🚀 Local Pipeline Run: SUCCESS")
}

// runCommand helper routes output streams cleanly to your console in real-time
func runCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
