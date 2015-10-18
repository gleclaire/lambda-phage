package main

import "gopkg.in/yaml.v2"
import "github.com/spf13/cobra"
import "github.com/peterh/liner"
import "strconv"
import "strings"
import "fmt"
import "io/ioutil"
import "os"

func init() {
	initCmd := &cobra.Command{
		Use:   "init",
		Short: "initializes a config for your function",
		Run:   initPhage,
	}

	cmds = append(cmds, initCmd)
}

type prompt struct {
	text           string
	def            string
	stringStore    **string
	stringSetStore *[]*string
	intStore       **int64
}

// helps you build a config file
func initPhage(c *cobra.Command, _ []string) {
	l := liner.NewLiner()

	fmt.Println(`
		HELLO AND WELCOME

		This command will help you set up your code for deployment to lambda!
		Please answer the prompts as they appear below:
	`)

	//reqMsg := "Sorry, that field is required. Try again."

	cfg := new(Config)
	cfg.IamRole = new(IamRole)
	cfg.Location = new(Location)
	wd, _ := os.Getwd()
	st, _ := os.Stat(wd)

	prompts := []prompt{
		prompt{
			"Enter a project name",
			st.Name(),
			&cfg.Name,
			nil,
			nil,
		},
		prompt{
			"Enter a project description if you'd like",
			"",
			&cfg.Description,
			nil,
			nil,
		},
		prompt{
			"Enter an archive name if you'd like",
			st.Name() + ".zip",
			&cfg.Archive,
			nil,
			nil,
		},
		prompt{
			"What runtime are you using: nodejs, java8, or python 2.7?",
			"nodejs",
			&cfg.Runtime,
			nil,
			nil,
		},

		prompt{
			"Enter an entry point or handler name",
			"index.handler",
			&cfg.EntryPoint,
			nil,
			nil,
		},

		prompt{
			"Enter memory size",
			"128",
			nil,
			nil,
			&cfg.MemorySize,
		},
		prompt{
			"Enter timeout",
			"5",
			nil,
			nil,
			&cfg.Timeout,
		},
		prompt{
			"Enter AWS regions where this function will run",
			"us-east-1",
			nil,
			&cfg.Regions,
			nil,
		},
		prompt{
			"Enter IAM role name",
			"us-east-1",
			&cfg.IamRole.Name,
			nil,
			nil,
		},
	}

	for _, cPrompt := range prompts {
		p := cPrompt
		text := p.text
		if p.def != "" {
			text += " [" + p.def + "]"
		}

		text += ": "
		if s, err := l.Prompt(text); err == nil {
			input := s
			hasInput := input != ""
			if p.stringStore != nil {
				if hasInput {
					*p.stringStore = &input
				} else {
					*p.stringStore = &p.def
				}

			} else if p.stringSetStore != nil {
				var splitMe string
				if hasInput {
					splitMe = input
				} else {
					splitMe = p.def
				}

				spl := strings.Split(splitMe, ",")
				pspl := make([]*string, len(spl))
				for i, v := range spl {
					// we need to set the value
					// in a variable local to this block
					// because the pointed-to value in
					// `v` will change on the next
					// loop iteration
					realVal := v
					pspl[i] = &realVal
				}

				*p.stringSetStore = pspl

			} else if p.intStore != nil {
				var tParse string
				if hasInput {
					tParse = input
				} else {
					tParse = p.def
				}

				i, _ := strconv.ParseInt(tParse, 10, 64)
				*p.intStore = &i
			}
		} else if err == liner.ErrPromptAborted {
			fmt.Println("Aborted")
			return
		} else {
			fmt.Println("Error reading line: ", err)
			return
		}
	}

	l.Close()

	d, err := yaml.Marshal(cfg)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile("l-p.yml", d, os.FileMode(0644))
}
