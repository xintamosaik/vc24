# README

# Setup Go Environment
1. Install Go: Download and install Go from the [official website](https://go.dev/dl/).
2. Optionally get IDE support by installing the [Go extension for Visual Studio Code](https://marketplace.visualstudio.com/items?itemName=golang.Go) or use another IDE of your choice that supports Go development.
3. Install templ as a tool:
```bash
   go get -tool github.com/a-h/templ/cmd/templ@latest
 ```

# Development
When using templ with this project, ensure that any changes to the templates are
reflected by restarting the server. You can quickly restart the server (main.go)
using the following one-liner:


First, generate the templates by running:
```bash
    go tool tmpl generate
```

CTRL-C to stop the server and then run the command again to start it up again:
```bash

    go run main.go
```

Or in one line:
```bash
    go tool tmpl generate && go run main.go
```

I only support unix-like systems, so if you're using Windows, you might need to adapt the command accordingly.

