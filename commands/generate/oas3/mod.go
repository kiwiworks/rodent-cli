package oas3

import (
	"os"
	"path"
	"text/template"

	"github.com/kiwiworks/rodent/errors"
)

const goModTemplateText = `module {{ .ModuleName }}
go {{ .GoVersion }}

require (
{{ range .Imports }}
	{{ .Package }} {{ .Version }}
{{ end }}
)
`

var goModTemplate = template.Must(template.New("go.mod").Parse(goModTemplateText))

type goModTemplateArgs struct {
	ModuleName string
	GoVersion  string
	Imports    []Import
}

type Import struct {
	Package string
	Version string
}

func (g *Generator) generateModule() error {
	moduleName := g.moduleName
	goModFile, err := os.Create(path.Join(g.outputDir, "go.mod"))
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return errors.Wrapf(err, "failed to create go.mod file")
		}
		goModFile, err = os.OpenFile(path.Join(g.outputDir, "go.mod"), os.O_WRONLY, 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to open go.mod file")
		}
	}
	defer func(goModFile *os.File) {
		err := goModFile.Close()
		if err != nil {
			panic(err)
		}
	}(goModFile)

	err = goModTemplate.Execute(goModFile, goModTemplateArgs{
		ModuleName: moduleName,
		GoVersion:  "1.23.0",
		Imports:    g.imports,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to execute go.mod template")
	}
	return nil
}
