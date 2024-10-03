package main

import (
	"github.com/kiwiworks/rodent-cli/commands"
	"github.com/kiwiworks/rodent/app"
)

func main() {
	app.New("rodent-cli", "0.1.0", app.Modules(commands.Module)).Run()
}
