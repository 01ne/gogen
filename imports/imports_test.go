package imports_test

import (
	"go/token"
	"go/types"
	"testing"

	"github.com/01ne/gogen/imports"
	"github.com/01ne/testify/require"
)

// TestDevendorizeImportPaths checks if vendored
// import paths are devendorized correctly.
func TestDevendorizeImportPaths(t *testing.T) {
	i := imports.New("github.com/01ne/gogen/imports")
	pkg := types.NewPackage("github.com/01ne/gogen/vendor/github.com/01ne/testify/mock", "mock")
	named := types.NewNamed(types.NewTypeName(token.Pos(0), pkg, "", &types.Array{}), &types.Array{}, nil)
	i.AddImportsFrom(named)
	require.Equal(t, map[string]string{"github.com/01ne/testify/mock": "mock"}, i.Imports())
}
