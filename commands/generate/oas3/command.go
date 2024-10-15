package oas3

import "github.com/kiwiworks/rodent/command"

func GenerateOpenapiCommandGroup() *command.Command {
	return command.New(
		"generate.openapi", "oas3", "Generate resources from OpenAPI 3.0 specifications",
	)
}
