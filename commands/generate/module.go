package generate

import (
	"github.com/kiwiworks/rodent-cli/commands/generate/oas3"
	"github.com/kiwiworks/rodent/app"
	"github.com/kiwiworks/rodent/app/module"
	"github.com/kiwiworks/rodent/command"
)

func Module() app.Module {
	return app.NewModule(
		command.Commands(
			CommandGroup,
		),
		module.SubModules(
			oas3.Module,
		),
	)
}
