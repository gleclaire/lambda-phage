package main

import "fmt"
import "github.com/spf13/cobra"
import "os"
import "github.com/tj/go-debug"
import "github.com/lucsky/cuid"
import "runtime"
import "strings"

func init() {
	pkgCmd := &cobra.Command{
		Use:   "pkg",
		Short: "adds all the current folder to a zip file recursively",
		Run:   pkg,
	}
	flg := pkgCmd.Flags()

	flg.BoolP("verbose", "v", false, "verbosity")
	flg.StringP("output", "o", "", "output file name")

	cmds = append(cmds, pkgCmd)
}

// packages your package up into a zip file
func pkg(c *cobra.Command, _ []string) {
	var err error
	debug := debug.Debug("cmd.pkg")
	var binName string

	if cfg != nil {
		binName = cfg.Pkg.Name
	}

	flagName, _ := c.Flags().GetString("output")
	if flagName != "" {
		binName = flagName
	}

	if binName == "" {
		binName = "lambda-phage-" + cuid.New()
	}

	// add ".zip" to the filename if one is not found
	if strings.Index(binName, ".zip") != (len(binName) - 4) {
		binName += ".zip"
	}

	zFile, err := newZipFile(binName)

	if err != nil {
		zipFileFail(err)
		return
	}

	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error opening directory, %s", wd)
		return
	}

	root, err := os.Open(".")

	if err != nil {
		fmt.Printf("Error opening directory, %s", wd)
		return
	}

	var infoCh chan string
	verbose, _ := c.Flags().GetBool("verbose")
	if verbose {
		infoCh = make(chan string, 1000)
	}

	errCh := make(chan error)

	go func() {
		err := zFile.AddDirectory(root, infoCh)
		if err != nil {
			errCh <- err
			return
		}

		if infoCh != nil {
			close(infoCh)
		}

		close(errCh)
	}()

	for {
		select {
		case i := <-infoCh:
			if i != "" {
				fmt.Println(i)
			}

		case e := <-errCh:
			if e != nil {
				fmt.Println(e)
			}
			debug("errored")
			return
		}
	}

	err = zFile.Close()

}

func zipFileFail(err error) {
	_, f, l, _ := runtime.Caller(1)
	fmt.Printf("[%s:%s]error creating zip file, %s\n", f, l, err.Error())
	return
}
