package main

import "github.com/spf13/cobra"
import "gopkg.in/yaml.v2"
import "io/ioutil"
import "github.com/tj/go-debug"
import "os"
import "fmt"

var cmds = []*cobra.Command{}
var cfg *Config

/*

# lambda-phage config file sample

name: my-first-lambda-function
description: provides some sample stuff
pkg:
  name: my-first-lambda-function.zip
deploy:
  type: s3
  s3-bucket: test-bucket
  use-versioning: true
*/

type Config struct {
	Name        string
	Description string
	Pkg         struct {
		Name string
	}
	Deploy struct {
		Type          string
		S3Bucket      string
		UseVersioning *bool
	}
}

func main() {
	defaultCfgName := "l-p.yml"
	debug := debug.Debug("main")

	root := &cobra.Command{Use: "lambda-phage"}
	root.PersistentFlags().BoolP("verbose", "v", false, "verbosity")
	root.PersistentFlags().StringP("config", "c", defaultCfgName, "config file location")
	root.ParseFlags(os.Args)
	for _, cmd := range cmds {
		root.AddCommand(cmd)
	}

	cf, _ := root.Flags().GetString("config")
	cfgExists := false

	if _, err := os.Stat(cf); err != nil &&
		(!os.IsNotExist(err) || cf != defaultCfgName) {
		fmt.Println("Error reading config file: %s", err.Error())
		return
	} else if err == nil {
		cfgExists = true
	}

	if cfgExists {
		debug("reading config file %s", cf)
		f, err := ioutil.ReadFile("l-p.yml")
		if err != nil {
			fmt.Println("Error reading config file: %s", err.Error())
			return
		}

		debug("decoding config")
		err = yaml.Unmarshal(f, &cfg)
		if err != nil {
			fmt.Println("Error reading config file: %s", err.Error())
			return
		}
	}

	root.Execute()
}
