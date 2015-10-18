package main

import "fmt"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/spf13/cobra"
import "os"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/tj/go-debug"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/lucsky/cuid"
import "runtime"
import "strings"

func init() {
	pkgCmd := &cobra.Command{
		Use:   "pkg",
		Short: "adds all the current folder to a zip file recursively",
		RunE:  pkg,
	}
	flg := pkgCmd.Flags()

	flg.BoolP("verbose", "v", false, "verbosity")
	flg.StringP("output", "o", "", "output file name")

	cmds = append(cmds, pkgCmd)
}

// packages your package up into a zip file
func pkg(c *cobra.Command, _ []string) error {
	var err error
	debug := debug.Debug("cmd.pkg")

	binName := getArchiveName(c)
	zFile, err := newZipFile(binName)

	if err != nil {
		return zipFileFail(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("Error opening directory, %s", wd)
	}

	root, err := os.Open(".")

	if err != nil {
		return fmt.Errorf("Error opening directory, %s", wd)
	}

	var infoCh chan string
	verbose, _ := c.Flags().GetBool("verbose")
	if verbose {
		infoCh = make(chan string, 1000)
	}

	doneCh := make(chan error)

	go func() {
		err := zFile.AddDirectory(root, infoCh)
		if err != nil {
			doneCh <- err
			return
		}

		if infoCh != nil {
			close(infoCh)
		}

		doneCh <- nil
	}()

	good := true
	for good {
		select {
		case i := <-infoCh:
			if i != "" {
				fmt.Println(i)
			}
		case e := <-doneCh:
			if e != nil {
				debug("errored")
				return e
			}
			good = false
		}
	}

	return zFile.Close()
}

// based on whatever has been passed in, this will determine the
// filename for the archive
func getArchiveName(c *cobra.Command) string {
	var binName string

	if cfg != nil {
		binName = *cfg.Name
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

	return binName
}

func zipFileFail(err error) error {
	_, f, l, _ := runtime.Caller(1)
	return fmt.Errorf("[%s:%s]error creating zip file, %s\n", f, l, err.Error())
}
