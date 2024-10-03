package commands

import (
	"github.com/kiwiworks/rodent-cli/commands/generate"
	"github.com/kiwiworks/rodent/command"
	"github.com/kiwiworks/rodent/module"
)

func Module() module.Module {
	return module.New(
		"rodent-cli",
		module.SubModules(command.Module),
		module.Commands(generate.Oas3Client),
	)
}
