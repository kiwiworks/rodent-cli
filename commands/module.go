package commands

import (
	"github.com/kiwiworks/rodent-cli/commands/generate"
	"github.com/kiwiworks/rodent/app"
	"github.com/kiwiworks/rodent/app/module"
	"github.com/kiwiworks/rodent/command"
)

func Module() app.Module {
	return app.NewModule(
		module.SubModules(
			// root command and cli core
			command.Module,
			// `rodent-cli generate` command group
			generate.Module,
		),
	)
}
