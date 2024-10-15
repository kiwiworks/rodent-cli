package generator

import (
	"context"
	"net/url"
	"os"

	"github.com/cavaliergopher/grab/v3"
	"go.uber.org/zap"

	"github.com/kiwiworks/rodent/errors"
	"github.com/kiwiworks/rodent/logger"
)

type Flags struct {
	OutputDir      string
	GenerateModule bool
}

func DefaultFlags() Flags {
	return Flags{
		OutputDir:      ".",
		GenerateModule: true,
	}
}

func Generate(ctx context.Context, uri url.URL, flags Flags) error {
	log := logger.FromContext(ctx)

	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return errors.Wrapf(err, "failed to create temp dir")
	}
	log.Debug("temporary directory created", zap.String("dir", tmpDir))

	var filename string
	switch uri.Scheme {
	case "http", "https":
		tmpFile, err := os.CreateTemp(tmpDir, "rodent-*")
		defer func(tmpFile *os.File) {
			err := tmpFile.Close()
			if err != nil {
				log.DPanic("failed to close temporary file", zap.Error(err))
			}
		}(tmpFile)
		if err != nil {
			return errors.Wrapf(err, "failed to create temp file")
		}
		log.Debug("temporary file created", zap.String("filename", tmpFile.Name()))
		response, err := grab.Get(tmpFile.Name(), uri.String())
		if err != nil {
			return errors.Wrapf(err, "failed to download spec from %s", uri.String())
		}
		log.Debug("file downloaded", zap.Int64("size", response.Size()))
		filename = tmpFile.Name()
	case "file":
		filename = uri.Path
	default:
		return errors.Newf("unsupported scheme %s", uri.Scheme)
	}
	generator, err := GeneratorFromFile(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to create generator")
	}
	return generator.Build(ctx, flags)
}
