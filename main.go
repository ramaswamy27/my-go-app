package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"
)

// PipelineConfig maps directly to your custom Jenkins YAML schema
type PipelineConfig struct {
	Stack        string            `yaml:"stack"`
	StackVersion string            `yaml:"stackVersion"`
	ImageName    string            `yaml:"imageName"`
	ImageVer     interface{}       `yaml:"imageVer"` // Dynamic handling for numerical/string versions
	Resources    map[string]string `yaml:"resources"`
	BuildCmd     string            `yaml:"buildCmd"`
	TestCmd      string            `yaml:"testCmd"`
}

func main() {
	// 1. Load pipeline-config.yaml
	yamlFile, err := os.ReadFile("pipeline-config.yaml")
	if err != nil {
		log.Fatalf("❌ Failed to read pipeline-config.yaml: %v", err)
	}

	var config PipelineConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Fatalf("❌ Failed to parse YAML: %v", err)
	}

	// 2. Resolve the exact Docker base image
	baseImage := fmt.Sprintf("%s:%s", config.Stack, config.StackVersion)
	fmt.Printf("🐳 Target Environment Image: %s\n", baseImage)
	fmt.Printf("📦 Internal App Name:       %s:%v\n", config.ImageName, config.ImageVer)

	// 3. Fail-fast validation for placeholders
	if strings.Contains(config.BuildCmd, "<required") || strings.Contains(config.TestCmd, "<test") {
		log.Fatalf("❌ Error: Please update placeholder build/test commands in your yaml configuration.")
	}

	// 4. Run Build and Test commands inside the designated Docker environment
	commands := []struct {
		name string
		cmd  string
	}{
		{"Build Stage", config.BuildCmd},
		{"Test Stage", config.TestCmd},
	}

	currentDir, _ := os.Getwd()

	for _, step := range commands {
		fmt.Printf("\n--- Running %s ---\n", step.name)
		fmt.Printf("Command: %s\n\n", step.cmd)

		// Spawns an ephemeral container matching the 'stack' environment, mounting your local working directory
		dockerArgs := []string{
			"run", "--rm",
			"-v", fmt.Sprintf("%s:/app", currentDir),
			"-w", "/app",
		}

		// Apply memory constraints if present in the YAML configurations
		if mem, exists := config.Resources["memory"]; exists {
			dockerArgs = append(dockerArgs, "-m", mem)
		}

		// Append the target engine image and runtime command execution logic
		dockerArgs = append(dockerArgs, baseImage, "sh", "-c", step.cmd)

		execCmd := exec.Command("docker", dockerArgs...)
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr

		if err := execCmd.Run(); err != nil {
			log.Fatalf("❌ Pipeline execution stopped: %s encountered a failure: %v", step.name, err)
		}
	}

	fmt.Println("\n✅ Verification Successful! Local environment passed all definitions.")
}

