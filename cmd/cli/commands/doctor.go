package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Diagnose the project environment",
	Long:  "Checks installed tools, Docker, and project setup.",
	RunE:  runDoctor,
}

type toolCheck struct {
	name        string
	cmd         string
	versionArgs []string
	installHint string
}

var tools = []toolCheck{
	{"Go", "go", []string{"version"}, "https://go.dev/doc/install"},
	{"Docker", "docker", []string{"info"}, "https://docs.docker.com/get-docker/"},
	{"golangci-lint", "golangci-lint", []string{"version"}, "brew install golangci-lint"},
	{"swag", "swag", []string{"--version"}, "go install github.com/swaggo/swag/cmd/swag@latest"},
	{"goose", "goose", []string{"--version"}, "brew install goose"},
	{"air", "air", []string{"-v"}, "go install github.com/air-verse/air@latest"},
	{"k6", "k6", []string{"version"}, "brew install k6"},
	{"kind", "kind", []string{"version"}, "brew install kind"},
	{"kubectl", "kubectl", []string{"version", "--client", "--short"}, "brew install kubectl"},
}

func runDoctor(_ *cobra.Command, _ []string) error {
	fmt.Println("gopherplate doctor")
	fmt.Println()

	for _, tool := range tools {
		checkTool(tool)
	}

	fmt.Println()
	fmt.Println("Project:")
	checkProject()

	return nil
}

func checkTool(t toolCheck) {
	path, lookErr := exec.LookPath(t.cmd)
	if lookErr != nil {
		fmt.Printf("  [!!] %s - not found (install: %s)\n", t.name, t.installHint)
		return
	}

	// Docker has a different check - running daemon
	if t.cmd == "docker" {
		if runErr := exec.Command("docker", "info").Run(); runErr != nil {
			fmt.Printf("  [!!] %s - installed at %s but daemon is NOT running\n", t.name, path)
			return
		}
		fmt.Printf("  [OK] %s - running\n", t.name)
		return
	}

	// Get version
	output, runErr := exec.Command(t.cmd, t.versionArgs...).CombinedOutput()
	version := "installed"
	if runErr == nil {
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 {
			version = strings.TrimSpace(lines[0])
		}
	}
	fmt.Printf("  [OK] %s - %s\n", t.name, version)
}

func checkProject() {
	if _, statErr := os.Stat("go.mod"); os.IsNotExist(statErr) {
		fmt.Println("  [!!] go.mod not found (not a Go project)")
		return
	}
	fmt.Println("  [OK] go.mod found")

	// Check Docker containers (optional - only if Docker is running)
	if runErr := exec.Command("docker", "info").Run(); runErr == nil {
		psOutput, _ := exec.Command("docker", "ps", "--format", "{{.Names}}").CombinedOutput()
		containers := string(psOutput)
		fmt.Println("  Docker containers:")
		// Check common infra containers
		checkContainer(containers, "postgres")
		checkContainer(containers, "redis")
	}
}

func checkContainer(containers, name string) {
	if strings.Contains(strings.ToLower(containers), name) {
		fmt.Printf("    [OK] %s running\n", name)
	} else {
		fmt.Printf("    [--] %s not running\n", name)
	}
}
