package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	rcon "github.com/forewing/csgo-rcon"
)

// Flags of the command line
type Flags struct {
	Address     *string  `json:",omitempty"`
	Password    *string  `json:",omitempty"`
	Timeout     *float64 `json:",omitempty"`
	Config      *string  `json:",omitempty"`
	From        *string  `json:",omitempty"`
	Interactive *bool    `json:",omitempty"`
}

var (
	flags Flags = Flags{
		Address:     flag.String("a", rcon.DefaultAddress, "`address` of the server RCON, in the format of HOST:PORT."),
		Password:    flag.String("p", rcon.DefaultPassword, "`password` of the RCON."),
		Timeout:     flag.Float64("t", rcon.DefaultTimeout.Seconds(), "`timeout` of the connection (seconds)."),
		Config:      flag.String("c", "", "load configs from `file` instead of flags."),
		From:        flag.String("f", "", "read commands from `file`, \"-\" for stdin. From arguments if not set."),
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

func runInteractive() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">>> ")
		data, _, err := reader.ReadLine()
		if err != nil {
			return
		}
		message, err := client.Execute(string(data))
		fmt.Println(strings.TrimSpace(message))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
