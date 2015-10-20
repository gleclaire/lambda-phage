package main

import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/spf13/cobra"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/gopkg.in/yaml.v2"
import "io/ioutil"
import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/tj/go-debug"
import "os"
import "os/user"
import "fmt"
import "crypto/sha1"

func init() {
	prjCmd := &cobra.Command{
		Use:   "project [command]",
		Short: "does project stuff",
	}

	//flg := prjCmd.Flags()

	prjCmd.AddCommand(&cobra.Command{
		Use:   "create [name]",
		Short: "creates a new project with the specified name & adds the current folder to it",
		RunE:  createProjectCmd,
	})

	prjCmd.AddCommand(&cobra.Command{
		Use:   "add [project]",
		Short: "adds the current folder to a project",
		RunE:  addToProjectCmd,
	})

	cmds = append(cmds, prjCmd)
}

// loads a project file or creates a new one
func getProject(pName string) (*Project, error) {
	var err error
	prjDir, err := createProjectDir()
	if err != nil {
		return nil, err
	}

	pFilePath := fmt.Sprintf("%s/%s.yml", prjDir, pName)

	pCfg, err := openProject(pFilePath)

	if err != nil {
		return nil, fmt.Errorf("Error getting project file:\n%s\n", err)
	}

	pCfg.fName = pFilePath
	return pCfg, nil
}

// defines the configuration for a project
type Project struct {
	// YAML filename for project
	fName     string
	Functions map[string]ProjectFunction
}

type ProjectFunction struct {
	Id     string
	Config *Config
	Path   string
}

// writes this to a yaml file
func (p *Project) writeToFile() error {
	return writeToYamlFile(p, p.fName)
}

// adds a lambda function to this project
// or maybe not if the config file path
// is
func (p *Project) addFunction(c *Config) *Project {
	debug := debug.Debug("project.addFunction")
	var nHash string

	if c.Name == nil {
		debug("bummer, your lambda function doesn't have a name")
		return p
	}

	nHash = fmt.Sprintf("%x", sha1.Sum([]byte(*c.Name+c.fName)))

	p.Functions[*c.Name] = ProjectFunction{
		Id:     nHash,
		Config: c,
		Path:   c.fName,
	}

	return p
}

// cobra command for creating project
func createProjectCmd(c *cobra.Command, args []string) error {
	if len(args) == 0 || args[0] == "" {
		return fmt.Errorf("You didn't give us a project name! Please give us one.")
	}
	pName := args[0]
	pCfg, err := getProject(pName)

	if err != nil {
		fmt.Printf("Error creating or opening project:\n%s\n", err)
		return nil
	}

	// add the current function to the project
	pCfg.addFunction(cfg)
	err = pCfg.writeToFile()

	if err != nil {
		fmt.Printf("Error saving project:\n%s\n", err)
		return nil
	}

	var action string

	switch c.Name() {
	case "add":
		action = fmt.Sprintf("added %s to", *cfg.Name)
	default:
		action = "created"
	}

	fmt.Printf("%s project %s\n", action, pName)

	return nil
}

// adds the current config file to the project
// uses createProjectCmd now, but will be separate from that
// so we reserve the right to change it whenever
func addToProjectCmd(c *cobra.Command, args []string) error {
	if cfg == nil || (cfg != nil && cfg.Name == nil) {
		return fmt.Errorf("No project configuration found! Please run `lambda-phage init` first!")
	}
	// just use create project for now
	return createProjectCmd(c, args)
}

// parses or creates a project config file
func openProject(fName string) (*Project, error) {
	debug := debug.Debug("core.openProject")
	// read the project file
	b, err := ioutil.ReadFile(fName)

	if err != nil && err == os.ErrNotExist {
		return nil, fmt.Errorf("Error reading project file: %s\n", err)
	}

	var pCfg *Project

	if b != nil {
		debug("have project data from %s, trying to parse it", fName)
		// if we got data from the file, parse it
		err = yaml.Unmarshal(b, &pCfg)
		if err != nil {
			return nil, fmt.Errorf("Error parsing project file %s:\n %s\n", fName, err)
		}
	} else {
		debug("no project found, making empty struct")
		pCfg = new(Project)
		pCfg.fName = fName
		pCfg.Functions = make(map[string]ProjectFunction)
	}

	return pCfg, nil
}

// creates a directory for storing projects
// returns the path to that directory
// or an error if doing something failed
//
// TODO: wtf will this do on windows?
func createProjectDir() (string, error) {
	debug := debug.Debug("project.createProjectDir")
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	debug("got user info; home directory is %s", u.HomeDir)

	cfgDir := u.HomeDir + `/.lambda_phage`
	prjDir := cfgDir + "/projects"

	f, err := os.Stat(prjDir)
	if f != nil && f.IsDir() {
		return prjDir, nil
	}

	err = os.Mkdir(prjDir, 0755)
	if err != nil {
		return "", err
	}

	prjDir += "/projects"
	err = os.Mkdir(prjDir, 0755)
	if err != nil {
		return "", err
	}

	return prjDir, nil
}
