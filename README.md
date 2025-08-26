# Chinchón

A multiplayer Chinchón implementation in Go, featuring:

- [websocket-based client/server architecture](https://github.com/devblac/chinchon-backend/tree/main/server)
- [example terminal-based frontend](https://github.com/devblac/chinchon-backend/blob/main/exampleclient/websocket_client.go)
- [example bot](https://github.com/devblac/chinchon-backend/blob/main/examplebot/bot.go), and a simple, [documented](https://github.com/devblac/chinchon-backend/blob/main/CONTRIBUTING.md#making-your-own-bot) interface for making your own
- WASM build support for browser integration

### Installation

Either install using Go

```bash
$ go install https://github.com/devblac/chinchon-backend@latest
```

Or download the [latest release binary](https://github.com/devblac/chinchon-backend/releases) for your OS.

### Usage

Start a server

```bash
$ chinchon server
```

You may change the port (default is 8080) via environment variable

```bash
$ PORT=1234 chinchon server
```

If you want to play via example terminal-based frontend, start two clients on separate terminals

```bash
$ chinchon player 1
```

```bash
$ chinchon player 2
```

### Playing with someone else over the Internet

Whoever starts the server may expose it to the Internet somehow, e.g. via `cloudflared` tunnels

```bash
$ cloudflared tunnel --url localhost:8080
```

Then, the clients can connect to the address the tunnel provides, e.g. if tunnel says

```bash
...
2024-06-23T18:35:10Z INF +--------------------------------------------------------------------------------------------+
2024-06-23T18:35:10Z INF |  Your quick Tunnel has been created! Visit it at (it may take some time to be reachable):  |
2024-06-23T18:35:10Z INF |  https://retail-curves-bernard-affairs.trycloudflare.com                                   |
2024-06-23T18:35:10Z INF +--------------------------------------------------------------------------------------------+
```

Start the clients with

```bash
$ chinchon player 1 retail-curves-bernard-affairs.trycloudflare.com
```

```bash
$ chinchon player 2 retail-curves-bernard-affairs.trycloudflare.com
```

### Reconnect after issue

If the server dies, state is gone. If client dies, you can simply reconnect to the same server and game goes on.

### I don't like your UI

It's just an example UI. I encourage you to [implement your own frontend](https://github.com/devblac/chinchon-backend/blob/main/CONTRIBUTING.md#making-your-own-frontend). You may [browse the documentation](https://github.com/devblac/chinchon-backend/blob/main/CONTRIBUTING.md) and the [existing terminal UI code](https://github.com/devblac/chinchon-backend/blob/main/exampleclient/ui.go) to guide your implementation.

### I don't like your Bot

It's just an example bot that implements basic chinchon strategy. I encourage you to [implement your own bot](https://github.com/devblac/chinchon-backend/blob/main/CONTRIBUTING.md#making-your-own-bot). You may [browse the documentation](https://github.com/devblac/chinchon-backend/blob/main/CONTRIBUTING.md) and the [existing bot code](https://github.com/devblac/chinchon-backend/blob/main/examplebot/bot.go) to guide your implementation.

## Technology stack

- This chinchon engine is written 100% in Go
- Terminal-based UI uses [Termbox](https://github.com/nsf/termbox-go)
- WASM support uses [TinyGo](https://tinygo.org/) with WASM target to transpile to WebAssembly for browser integration

### Known issues / limitations

- Don't resize your terminal. This is a go-termbox issue. Also, have a terminal with a decent viewport.

### Issues / Improvements

Please do [create issues](https://github.com/devblac/chinchon-backend/issues) and send PRs. Also feel free to reach me for comments / discussions.
