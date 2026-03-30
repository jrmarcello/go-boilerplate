package commands

import "github.com/spf13/cobra"

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add components to an existing project",
	Long:  "Add domains, endpoints, or other components to an existing project.",
}

func init() {
	addCmd.AddCommand(addDomainCmd)
}
