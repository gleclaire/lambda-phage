package main

import "github.com/spf13/cobra"

var cmds = []*cobra.Command{}

func main() {

	root := &cobra.Command{Use: "lambda-phage"}
	root.PersistentFlags().BoolP("verbose", "v", false, "verbosity")

	for _, cmd := range cmds {
		root.AddCommand(cmd)
	}

	root.Execute()
}
