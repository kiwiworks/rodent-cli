package oas3

import (
	"github.com/kiwiworks/rodent-cli/commands/generate/oas3/client"
	"github.com/kiwiworks/rodent/app"
	"github.com/kiwiworks/rodent/command"
)

func Module() app.Module {
	return app.NewModule(
		command.Commands(
			GenerateOpenapiCommandGroup,
			client.GenerateOpenAPIClient,
		),
	)
}
