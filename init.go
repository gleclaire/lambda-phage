package main

import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/gopkg.in/yaml.v2"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/spf13/cobra"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/peterh/liner"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/service/iam"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/aws/aws-sdk-go/aws"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/tj/go-debug"
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
	funcStore      func(string)
	completer      func(string) []string
}

func newPrompt() *prompt {
	return new(prompt)
}

func (p *prompt) withCompleter(f liner.Completer) *prompt {
	p.completer = f
	return p
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

func (p *prompt) withFunc(s func(string)) *prompt {
	p.funcStore = s
	return p
}

// helps you build a config file
func initPhage(c *cobra.Command, _ []string) {
	l := liner.NewLiner()
	l.SetCtrlCAborts(true)
	fmt.Println(`
		HELLO AND WELCOME

		This command will help you set up your code for deployment to lambda!
		Please answer the prompts as they appear below:
	`)

	//reqMsg := "Sorry, that field is required. Try again."

	// set this callback we can use to call all the stuff
	var realCompleter liner.Completer
	l.SetCompleter(func(line string) []string {
		if realCompleter != nil {
			return realCompleter(line)
		}
		return nil
	})

	cfg := new(Config)
	cfg.IamRole = new(IamRole)
	cfg.Location = new(Location)
	prompts := getPrompts(cfg)

	for _, cPrompt := range prompts {
		p := cPrompt
		text := p.text
		if p.def != "" {
			text += " [" + p.def + "]"
		}

		text += ": "

		realCompleter = nil
		if p.completer != nil {
			realCompleter = p.completer
		}

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
			} else if p.funcStore != nil {
				// just call the function, man
				p.funcStore(s)
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

	err = ioutil.WriteFile("l-p.yml", d, os.FileMode(0644))
	if err != nil {
		panic(err)
	}
}

// returns all the prompts needed for the `init` command
func getPrompts(cfg *Config) []*prompt {
	wd, _ := os.Getwd()
	st, _ := os.Stat(wd)

	iamRoles, roleMap := getIamRoles()

	return []*prompt{
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
			withFunc(
			func(s string) {
				// if this looks like an ARN,
				// we'll assume it is... for now
				if strings.Index(s, "arn:aws:iam::") == 0 {
					cfg.IamRole.Arn = &s
				} else {
					// check to see if this name is inside the role map
					if arn, ok := roleMap[s]; ok {
						cfg.IamRole.Arn = arn
					} else {
						// if not in role map, set the name
						cfg.IamRole.Name = &s
					}
				}
			},
		).
			isRequired().
			setText("Enter IAM role name").
			setDef("").
			withCompleter(
			func(l string) []string {
				c := make([]string, 0)
				for _, role := range iamRoles {
					if strings.HasPrefix(*role.Name, l) {
						c = append(c, *role.Name)
					}
				}

				return c
			},
		),
	}
}

// pulls all the IAM roles from your account
func getIamRoles() ([]*IamRole, map[string]*string) {
	debug := debug.Debug("core.getIamRoles")
	i := iam.New(nil)
	r, err := i.ListRoles(&iam.ListRolesInput{
		// try loading up to 1000 roles now
		MaxItems: aws.Int64(1000),
	})

	if err != nil {
		debug("getting IAM roles failed! maybe you don't have permission to do that?")
		return []*IamRole{}, map[string]*string{}
	}

	roles := make([]*IamRole, len(r.Roles))
	roleMap := make(map[string]*string)
	for i, r := range r.Roles {
		roles[i] = &IamRole{
			Arn:  r.Arn,
			Name: r.RoleName,
		}
		roleMap[*r.RoleName] = r.Arn
	}

	return roles, roleMap
}
