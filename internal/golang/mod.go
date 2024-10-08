package golang

import (
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"

	"github.com/kiwiworks/rodent/errors"
)

// FindModulePath returns the name of the golang module in which the path currently lies, or an error if any.
func FindModulePath(path string) (string, error) {
	if path == "" {
		return "", errors.Newf("path cannot be empty")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get absolute path for %s", path)
	}
	for {
		modFilePath := filepath.Join(absPath, "go.mod")
		if _, err := os.Stat(modFilePath); err == nil {
			modFileContent, err := os.ReadFile(modFilePath)
			if err != nil {
				return "", errors.Wrapf(err, "failed to read go.mod file at %s", modFilePath)
			}
			modFile, err := modfile.Parse("go.mod", modFileContent, nil)
			if err != nil {
				return "", errors.Wrapf(err, "failed to parse go.mod file at %s", modFilePath)
			}
			if modFile.Module != nil {
				return modFile.Module.Mod.Path, nil
			}
			return "", errors.Newf("go.mod file found at %s but module declaration is missing", modFilePath)
		}
		parentPath := filepath.Dir(absPath)
		if parentPath == absPath {
			break
		}
		absPath = parentPath
	}
	return "", errors.Newf("no go.mod file found in the path hierarchy for %s", path)
}
