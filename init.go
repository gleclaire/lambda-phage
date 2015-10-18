package main

//import "gopkg.in/yaml.v2"
import "github.com/spf13/cobra"
import "github.com/peterh/liner"
import "fmt"
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
	wd, _ := os.Getwd()
	st, _ := os.Stat(wd)

	prompts := []prompt{
		prompt{
			"Enter a project name",
			st.Name(),
			&cfg.Name,
			nil,
		},
	}

	for _, p := range prompts {
		text := p.text
		if p.def != "" {
			text += " [" + p.def + "]"
		}

		text += ": "
		if s, err := l.Prompt(text); err == nil {
			if p.stringStore != nil {
				*p.stringStore = &s
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
}
