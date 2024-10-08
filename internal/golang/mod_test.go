package golang

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParentModFromPath(t *testing.T) {
	r := require.New(t)

	mod, err := FindModulePath("./internal/golang/mod_test.go")
	r.NoError(err)
	r.Equal("github.com/kiwiworks/rodent-cli", mod)
}
