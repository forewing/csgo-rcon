# csgo-rcon

[![Go Report Card](https://goreportcard.com/badge/github.com/forewing/csgo-rcon?style=flat-square)](https://goreportcard.com/report/github.com/forewing/csgo-rcon)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/forewing/csgo-rcon?style=flat-square)](https://github.com/forewing/csgo-rcon/releases/latest)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/forewing/csgo-rcon)](https://pkg.go.dev/github.com/forewing/csgo-rcon)

Golang package for CS:GO RCON Protocol client. Also support other games using the protocol (Source Engine games, Minecraft, etc.)

> For the protocol specification, go to [Source RCON Protocol from Valve](http://developer.valvesoftware.com/wiki/Source_RCON_Protocol)

> Need a web-based admin panel? Check it out at [forewing/webrcon-server](https://github.com/forewing/webrcon-server)!

## Usage

1. Import the package

```go
import "github.com/forewing/csgo-rcon"
```

2. Create a client with `rcon.New(address, password string, timeout time.Duration)`, assuming your server rcon are hosted at `10.114.51.41:27015`, with password `password`, and you want the connection timeout to be 2 seconds.

```go
c := rcon.New("10.114.51.41:27015", "password", time.Second * 2)
```

3. Execute commands use `*Client.Execute(cmd string)`. Execute once if no "\n" provided. Return result message and nil on success, empty string and an error on failure.

```go
// Execute a single command
msg, err := c.Execute("bot_add_ct")

// Execute multiple commands at once
// Source engine games treat `;` as command separator
// May not work in other games, test before use
msg, err := c.Execute("game_mode 1; game_type 0; changelevel de_dust2")
```

4. Note: If `cmd` includes "\n", it is treated as a script file. Splitted and trimmed into lines. Line starts with "//" will be treated as comment and ignored. When all commands seccess, concatted messages and nil will be returned. Once failed, concatted previous succeeded messages and an error will be returned.

```go
cmd := `game_mode 1
game_type 0
// run_game half_life_3 (ignored)
changelevel de_dust2`

// Execute multiple commands separately
msg, err := c.Execute(cmd)
```


## Command Line Tool

### Install

```
go get -u github.com/forewing/csgo-rcon/cmd/csgo-rcon
```

Or download from [release page](https://github.com/forewing/csgo-rcon/releases/latest).

### Usage

```
Usage of csgo-rcon:
  -a address
        address of the server RCON, in the format of HOST:PORT. (default "127.0.0.1:27015")
  -c file
        load configs from file instead of flags.
  -f file
        read commands from file, "-" for stdin. From arguments if not set.
  -i    interact with the console.
  -m  file
        read completions from file
  -p password
        password of the RCON.
  -t timeout
        timeout of the connection (seconds). (default 1)
```

1. From arguments

```
$ csgo-rcon -c config.json mp_warmuptime 999
L **/**/20** - **:**:**: rcon from "**.**.**.**:***": command "mp_warmuptime 999"
```

2. From file (`-` for stdin)

```
$ csgo-rcon -c config.json -f commands.cfg
```

3. Interactive

```
$ csgo-rcon -c config.json -i
>>> bot_add_ct
L **/**/20** - **:**:**: "Derek<4><BOT><>" connected, address ""
L **/**/20** - **:**:**: "Derek<4><BOT>" switched from team <Unassigned> to <CT>
L **/**/20** - **:**:**: "Derek<4><BOT><>" entered the game
L **/**/20** - **:**:**: rcon from "**.**.**.**:***": command "bot_add_ct"
>>> users
<slot:userid:"name">
0 users
L **/**/20** - **:**:**: rcon from "**.**.**.**:***": command "users"
>>> ^C
```

4 .Completion

``` sh
# First download the completion file from your server
csgo-rcon -c config.json cvarlist > cmds.txt
# and remove top 2 and bottom 2 lines
tail -n +3 cmds.txt | head -n -2 > cmds.txt.bak && mv cmds.txt.bak cmds.txt
# then use -m flag to specify the completion file
csgo-rcon -c config.json -i -m cmds.txt
```

