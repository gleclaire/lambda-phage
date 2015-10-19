package main

import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/gopkg.in/yaml.v2"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/spf13/cobra"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/peterh/liner"
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
	required       bool
	stringStore    **string
	stringSetStore *[]*string
	intStore       **int64
	funcStore      func(interface{})
}

func newPrompt() *prompt {
	return new(prompt)
}

func (p *prompt) isRequired() *prompt {
	p.required = true
	return p
}

func (p *prompt) setDef(d string) *prompt {
	p.def = d
	return p
}

func (p *prompt) setText(t string) *prompt {
	p.text = t
	return p
}

func (p *prompt) withString(s **string) *prompt {
	p.stringStore = s
	return p
}

func (p *prompt) withStringSet(s *[]*string) *prompt {
	p.stringSetStore = s
	return p
}

func (p *prompt) withInt(s **int64) *prompt {
	p.intStore = s
	return p
}

func (p *prompt) withFunc(s func(interface{})) *prompt {
	p.funcStore = s
	return p
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

	prompts := []*prompt{
		newPrompt().
			withString(&cfg.Name).
			isRequired().
			setText("Enter a project name").
			setDef(st.Name()),
		newPrompt().
			withString(&cfg.Description).
			setText("Enter a project description if you'd like").
			setDef(""),
		newPrompt().
			withString(&cfg.Archive).
			setText("Enter a archive name if you'd like").
			setDef(st.Name() + ".zip"),
		newPrompt().
			withString(&cfg.Runtime).
			isRequired().
			setText("What runtime are you using: nodejs, java8, or python 2.7?").
			setDef("nodejs"),
		newPrompt().
			withString(&cfg.EntryPoint).
			isRequired().
			setText("Enter an entry point or handler name").
			setDef("index.handler"),
		newPrompt().
			withInt(&cfg.MemorySize).
			setText("Enter memory size").
			setDef("128"),
		newPrompt().
			withInt(&cfg.Timeout).
			setText("Enter timeout").
			setDef("5"),
		newPrompt().
			withStringSet(&cfg.Regions).
			setText("Enter AWS regions where this function will run").
			setDef("us-east-1"),
		newPrompt().
			withString(&cfg.IamRole.Name).
			isRequired().
			setText("Enter IAM role name").
			setDef(""),
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
