package generate

import (
	"net/url"
	"path"

	"github.com/spf13/cobra"

	"github.com/kiwiworks/rodent-cli/commands/generate/oas3"
	"github.com/kiwiworks/rodent/command"
	"github.com/kiwiworks/rodent/errors"
)

func Oas3Client() *command.Command {
	var (
		filename string
		fileUrl  string
	)
	flags := oas3.DefaultFlags()

	return command.New("oas3", "Generate OAS3 client", "todo",
		command.Runner(func(cmd *cobra.Command, args []string) error {
			if filename == "" && fileUrl == "" {
				return errors.Newf("either filename or url must be provided")
			}
			if filename != "" {
				if path.IsAbs(filename) {
					// make filename absolute
					filename = path.Clean(filename)
				}
				u := url.URL{
					Scheme: "file",
					Path:   filename,
				}
				return oas3.Generate(cmd.Context(), u, flags)
			}
			if fileUrl != "" {
				u, err := url.Parse(fileUrl)
				if err != nil {
					return err
				}
				return oas3.Generate(cmd.Context(), *u, flags)
			}
			return errors.Newf("not implemented")
		}),
		command.StringFlag(command.Flag{
			Name:        "filename",
			Shorthand:   "f",
			OneRequired: true,
			Exclusive:   true,
			Usage:       "input filename, accepts either yaml or json",
		}, &filename),
		command.StringFlag(command.Flag{
			Name:        "url",
			Shorthand:   "u",
			OneRequired: true,
			Exclusive:   true,
			Usage:       "input url to download the spec from, accepts either yaml or json",
		}, &fileUrl),
		command.StringFlag(command.Flag{
			Name:      "output",
			Shorthand: "o",
			Required:  true,
			Usage:     "output directory for the generated code",
		}, &flags.OutputDir),
		command.BoolFlag(command.Flag{
			Name:      "module",
			Shorthand: "m",
			Required:  false,
			Usage:     "generate module, if set to false, will generate a simple package without an associated go.mod file",
		}, &flags.GenerateModule),
	)
}
