package main

import "github.com/spf13/cobra"

func main() {
	pkgCmd := &cobra.Command{
		Use:   "pkg",
		Short: "adds all the current folder to a zip file recursively",
		Run:   pkg,
	}
	pkgCmd.Flags().BoolP("verbose", "v", false, "verbosity")

	root := &cobra.Command{Use: "lambda-phage"}
	root.PersistentFlags().BoolP("verbose", "v", false, "verbosity")
	root.AddCommand(pkgCmd)
	root.Execute()
}
