package main

import "github.com/hopkinsth/lambda-phage/Godeps/_workspace/src/github.com/peterh/liner"
import "strconv"
import "strings"
import "fmt"
import "sync"

type prompter interface {
	run() error
	add() *prompt
}

// describes a set of prompts
type promptSet struct {
	prompts []*prompt
	pmu     sync.Mutex
}

// creates and adds a new prompt to the set
func (ps *promptSet) add() *prompt {
	pr := newPrompt(ps)
	ps.pmu.Lock()
	ps.prompts = append(ps.prompts, pr)
	ps.pmu.Unlock()

	return pr
}

// uses prompt to run through your prompts in the order
// youv'e added them
func (ps *promptSet) run() error {
	ps.pmu.Lock()
	defer ps.pmu.Unlock()
	for _, p := range ps.prompts {
		p := p
		err := p.run()
		if err != nil {
			return err
		}
	}
	return nil
}

type prompt struct {
	ps             *promptSet
	text           string
	description    string
	def            string
	required       bool
	stringStore    **string
	stringSetStore *[]*string
	intStore       **int64
	funcStore      func(string)
	completer      func(string) []string
}

func newPrompt(ps *promptSet) *prompt {
	p := new(prompt)
	p.ps = ps
	return p
}

// this method is for use in chaining prompt calls
// and will return either:
// 1. the prompt set that this prompt belongs to OR
// 2. the prompt this method was called on
//
// both a prompt and a prompt set implement the `prompter` interface
// so you can call run() on either of them when your prompts are done
func (p *prompt) done() prompter {
	if p.ps != nil {
		return p.ps
	} else {
		return p
	}
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

func (p *prompt) setDescription(t string) *prompt {
	p.description = t
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

// for prompts, this is an identity function
// which makes it conform to the prompter interface
// ...so done() works the way we want
func (p *prompt) add() *prompt {
	return p
}

// runs this prompt and collects its result
func (p *prompt) run() error {
	l := liner.NewLiner()
	defer l.Close()
	l.SetCtrlCAborts(true)
	var realCompleter liner.Completer
	l.SetCompleter(func(line string) []string {
		if realCompleter != nil {
			return realCompleter(line)
		}
		return nil
	})

	text := p.text
	if p.def != "" {
		text += " [" + p.def + "]"
	}

	text += ": "

	realCompleter = nil
	if p.completer != nil {
		realCompleter = p.completer
	}

	if p.description != "" {
		fmt.Print(p.description)
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
		// TODO: return error
		return nil
	} else {
		fmt.Println("Error reading line: ", err)
		// TODO: return error
		return nil
	}

	return nil
}
