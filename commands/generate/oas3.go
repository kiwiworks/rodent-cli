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
	var filename string
	var fileUrl string
	var outputDir string

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
				return oas3.Generate(cmd.Context(), u, outputDir)
			}
			if fileUrl != "" {
				u, err := url.Parse(fileUrl)
				if err != nil {
					return err
				}
				return oas3.Generate(cmd.Context(), *u, outputDir)
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
		}, &outputDir),
	)
}
