package main

import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/spf13/cobra"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/gopkg.in/yaml.v2"
import "io/ioutil"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/tj/go-debug"
import "os"
import "fmt"
import "path/filepath"

var cmds = []*cobra.Command{}
var cfg *Config

var lambdaphage = `
     ,-^-.
     |\/\|
     '-V-'
       H
       H
       H
    .-;":-.
   ,'|  '; \
`

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

	wd, _ := os.Getwd()

	if _, err := os.Stat(cf); err != nil &&
		(!os.IsNotExist(err) || cf != defaultCfgName) {
		fmt.Println("Error reading config file: %s", err.Error())
		return
	} else if err == nil {
		cfgExists = true
	}

	if cfgExists {
		debug("reading config file %s", cf)
		f, err := ioutil.ReadFile(cf)
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

		if filepath.IsAbs(cf) {
			cfg.fName = cf
		} else {
			// if the path is not absolute, we need to make it absolute
			cfg.fName = wd + string(filepath.Separator) + cf
		}

	}

	fmt.Println(lambdaphage)
	root.Execute()
}
