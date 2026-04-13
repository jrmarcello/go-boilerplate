package commands

import "github.com/spf13/cobra"

var version = "dev" // set via ldflags at build time

var rootCmd = &cobra.Command{
	Use:   "gopherplate",
	Short: "Go microservice template scaffolding tool",
	Long: `Boilerplate CLI scaffolds new Go microservices and domains
following Clean Architecture patterns.

Commands:
  new              Create a new microservice from the template
  add domain       Add a new domain to an existing project
  add endpoint     Add a custom endpoint to an existing domain
  remove domain    Remove a domain from an existing project
  remove endpoint  Remove a custom endpoint from an existing domain
  wiring           Regenerate server.go, router.go, container.go from detected domains
  doctor           Diagnose project setup (tools, Docker, go.mod)
  version          Show CLI version`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(removeCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(wiringCmd)
	removeCmd.AddCommand(removeEndpointCmd)
}
