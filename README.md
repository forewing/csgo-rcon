# csgo-rcon

[![Go Report Card](https://goreportcard.com/badge/github.com/forewing/csgo-rcon?style=flat-square)](https://goreportcard.com/report/github.com/forewing/csgo-rcon)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/forewing/csgo-rcon?style=flat-square)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/forewing/csgo-rcon)](https://pkg.go.dev/github.com/forewing/csgo-rcon)

Golang package for CS:GO (or other Source Dedicated Server) RCON Protocol client.

> For the protocol specification, go to [Source RCON Protocol from Valve](http://developer.valvesoftware.com/wiki/Source_RCON_Protocol)

## Usage

1. Import the package

```go
import "github.com/forewing/csgo-rcon"
```

2. Create a client with `rcon.New(address, password string, timeout time.Duration)`, assuming your server rcon are hosted at `10.114.51.41:27015`, with password `password`, and you want the connection timeout to be 2 seconds.

```go
c := rcon.New("10.114.51.41:27015", "password", time.Seconds * 2)
```

3. Execute commands use `*Client.Execute(cmd string)`. On success, a message and nil error will be returned. On failure, an empty message and error will be returned.

## Command Line Tool

### Install

```
go get -u github.com/forewing/csgo-rcon/cmd/csgo-rcon
```

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