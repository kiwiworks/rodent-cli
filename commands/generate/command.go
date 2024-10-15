package generate

import "github.com/kiwiworks/rodent/command"

func CommandGroup() *command.Command {
	return command.New("generate", "g", "Generates new files")
}
