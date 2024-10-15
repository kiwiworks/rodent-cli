package generator

import (
	"context"
	"os"
	"os/exec"
	"path"

	"github.com/dave/jennifer/jen"
	"github.com/pb33f/libopenapi"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/kiwiworks/rodent/errors"
	"github.com/kiwiworks/rodent/logger"
)

type Generator struct {
	model      *libopenapi.DocumentModel[v3.Document]
	outputDir  string
	moduleName string
	files      map[string]*jen.File
	imports    []Import
}

func GeneratorFromFile(path string) (*Generator, error) {
	specBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read %s", path)
	}
	document, err := libopenapi.NewDocument(specBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse %s", path)
	}
	model, errs := document.BuildV3Model()
	if errs != nil {
		return nil, errors.Wrapf(multierr.Combine(errs...), "failed to build model from %s", path)
	}
	return NewGenerator(model), nil
}

func NewGenerator(model *libopenapi.DocumentModel[v3.Document]) *Generator {
	return &Generator{
		model:   model,
		files:   make(map[string]*jen.File),
		imports: []Import{},
	}
}

func (g *Generator) Build(ctx context.Context, flags Flags) error {
	outputDir := flags.OutputDir
	g.outputDir = outputDir
	g.moduleName = g.model.Model.Info.Title
	log := logger.FromContext(ctx)

	log.Info("spec metadata",
		zap.String("title", g.model.Model.Info.Title),
		zap.String("version", g.model.Model.Info.Version),
		zap.Int("paths", g.model.Model.Paths.PathItems.Len()),
		zap.Int("components.schemas", g.model.Model.Components.Schemas.Len()),
	)

	if err := g.generateSchemas(g.model.Model.Components.Schemas); err != nil {
		return err
	}

	if err := g.generateClientPackage(g.model.Model); err != nil {
		return err
	}

	for filename, f := range g.files {
		filename = path.Join(outputDir, filename)
		log.Info("saving", zap.String("filename", filename))
		dir := path.Dir(filename)
		stats, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return errors.Wrapf(err, "failed to create directory %s", dir)
				}
			} else {
				return errors.Wrapf(err, "failed to stat directory %s", dir)
			}
		} else if !stats.IsDir() {
			return errors.Newf("path %s is not a directory", dir)
		}
		if err := f.Save(filename); err != nil {
			return errors.Wrapf(err, "failed to save %s", filename)
		}
	}

	if flags.GenerateModule {
		if err := g.generateModule(); err != nil {
			return err
		}

		cmd := exec.Command("go", "mod", "tidy")
		cmd.Dir = outputDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			return errors.Wrapf(err, "failed to run go mod tidy: %s", output)
		}
	}
	return nil
}
