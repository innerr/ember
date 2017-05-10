package cli

import (
	"bufio"
	"os"
	"strings"
)

type Cmd func([]string)
//type Cmd func([]string)error

type Cmds struct {
	path     string
	cmds     map[string]Cmd
	children map[string]*Cmds
	help     map[string]string
}

func NewCmds() *Cmds {
	return &Cmds{"", make(map[string]Cmd), make(map[string]*Cmds), make(map[string]string)}
}

func (p *Cmds) Reg(cmd string, help string, fun func([]string)) {
	p.help[cmd] = help
	p.cmds[cmd] = fun
}

func (p *Cmds) Unreg(cmd string) (help string, fun func([]string)) {
	help = p.help[cmd]
	delete(p.help, cmd)
	fun = p.cmds[cmd]
	delete(p.cmds, cmd)
	return
}

func (p *Cmds) Sub(cmd string, help string) *Cmds {
	p.help[cmd] = help
	path := p.path + "." + cmd
	if len(p.path) == 0 {
		path = cmd
	}
	child := &Cmds{path, make(map[string]Cmd), make(map[string]*Cmds), make(map[string]string)}
	p.children[cmd] = child
	return child
}

func (p *Cmds) Loop() {
	lastest := ""
	for {
		print(p.path + ">> ")
		reader := bufio.NewReader(os.Stdin)
		line, _, _ := reader.ReadLine()
		cmd := strings.Trim(string(line), " ")
		if cmd == "." {
			cmd = lastest
			println(p.path + ">> " + cmd)
		} else {
			lastest = cmd
		}

		if cmd == "" {
			continue
		}
		if cmd == ".." {
			break
		}
		if _, ok := ExitCmds[cmd]; ok {
			os.Exit(0)
		}

		if _, ok := HelpCmds[cmd]; ok {
			p.Help(false)
			println()
			continue
		}

		p.Run(strings.Split(cmd, " "))
	}
}

func (p *Cmds) Help(simple bool) {
	display := func(mark bool, cmd string, help string) {
		if mark {
			print("* ", cmd)
		} else {
			print("  ", cmd)
		}
		println(strings.Repeat(" ", 12 - len(cmd)) + help)
	}

	if !simple {
		display(false, ".", "redo lastest cmd")
		display(false, "..", "back to uplevel")
		display(false, "-help", "display help message. '-?' will also do")
		display(false, "exit", "quit")
		println()
	}
	for cmd, _ := range p.children {
		help, _ := p.help[cmd]
		display(true, cmd, help)
	}
	for cmd, _ := range p.cmds {
		help, _ := p.help[cmd]
		display(false, cmd, help)
	}
}

func (p *Cmds) Run(cmds []string) {
	if len(cmds) == 0 {
		p.Loop()
		return
	}
	cmd, cmds := cmds[0], cmds[1:]
	if fun, ok := p.cmds[cmd]; ok {
		fun(cmds)
		return
	}
	if child, ok := p.children[cmd]; ok {
		child.Run(cmds)
		return
	}
	if _, ok := HelpCmds[cmd]; ok {
		p.Help(true)
		return
	}
	if cmd != "" {
		println("cmd not found:", cmd)
	}
	return
}

var HelpCmds = map[string]string{"--help": "", "--?": "", "-help": "", "-?": "", "help": "", "?": ""}
var ExitCmds = map[string]string{"exit": "", "quit": ""}
