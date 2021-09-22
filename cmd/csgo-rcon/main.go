package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	readline "github.com/chzyer/readline"
	rcon "github.com/forewing/csgo-rcon"
)

// Flags of the command line
type Flags struct {
	Address     *string  `json:",omitempty"`
	Password    *string  `json:",omitempty"`
	Timeout     *float64 `json:",omitempty"`
	Config      *string  `json:",omitempty"`
	From        *string  `json:",omitempty"`
	Completion  *string  `json:",omitempty"`
	Interactive *bool    `json:",omitempty"`
}

var (
	flags Flags = Flags{
		Address:     flag.String("a", rcon.DefaultAddress, "`address` of the server RCON, in the format of HOST:PORT."),
		Password:    flag.String("p", rcon.DefaultPassword, "`password` of the RCON."),
		Timeout:     flag.Float64("t", rcon.DefaultTimeout.Seconds(), "`timeout` of the connection (seconds)."),
		Config:      flag.String("c", "", "load configs from `file` instead of flags."),
		From:        flag.String("f", "", "read commands from `file`, \"-\" for stdin. From arguments if not set."),
		Completion:  flag.String("m", "", "read completions from `file`"),
		Interactive: flag.Bool("i", false, "interact with the console."),
	}

	client *rcon.Client
)

func init() {
	flag.Parse()
	if len(*flags.Config) == 0 {
		return
	}
	data, err := ioutil.ReadFile(*flags.Config)
	if err != nil {
		fatal(err.Error())
	}
	err = json.Unmarshal(data, &flags)
	if err != nil {
		fatal(err.Error())
	}
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func main() {
	client = rcon.New(*flags.Address, *flags.Password, time.Duration(*flags.Timeout*float64(time.Second)))

	if *flags.Interactive {
		runInteractive()
		return
	}

	if len(*flags.From) > 0 {
		runFile(*flags.From)
		return
	}

	runArgs()
}

func runArgs() {
	cmd := strings.TrimSpace(strings.Join(flag.Args(), " "))
	if len(cmd) == 0 {
		fatal("empty commands")
	}
	message, err := client.Execute(cmd)
	fmt.Println(strings.TrimSpace(message))
	if err != nil {
		fatal(err.Error())
	}
}

func runFile(filename string) {
	file := os.Stdin
	if filename != "-" {
		var err error
		file, err = os.Open(filename)
		if err != nil {
			fatal(err.Error())
		}
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		fatal(err.Error())
	}

	message, err := client.Execute(string(data))
	fmt.Println(strings.TrimSpace(message))
	if err != nil {
		fatal(err.Error())
	}
}

func getCommandCompletion() func(string) []string {
	return func(s string) []string {
		commands := make([]string, 0)
		if len(*flags.Completion) != 0 {
			file, err := os.Open(*flags.Completion)
			if err != nil {
				fatal(err.Error())
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				commands = append(commands, strings.Split(scanner.Text(), " ")[0])
			}
			if err := scanner.Err(); err != nil {
				fatal(err.Error())
			}
		}
		return commands
	}
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItemDynamic(getCommandCompletion()),
	readline.PcItem("bye"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func runInteractive() {
	l, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[34m>>> \033[0m",
		HistoryFile:         "/tmp/csgo-rcon.tmp",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()
	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				println("invalid mode:", line[5:])
			}
		case line == "mode":
			if l.IsVimMode() {
				println("current mode: vim")
			} else {
				println("current mode: emacs")
			}
		case line == "bye":
			goto exit
		case line == "":
		default:
			message, err := client.Execute(string(line))
			fmt.Println(strings.TrimSpace(message))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}
exit:
}
