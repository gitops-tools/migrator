package ext

import (
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/functions"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// Demo returns a cel.EnvOption to configure namespaced helper macros and
// functions.
func Demo() cel.EnvOption {
	r, err := types.NewRegistry()
	if err != nil {
		panic(err) // TODO: Do something better?
	}

	return cel.Lib(&demoLib{registry: r})
}

type demoLib struct {
	// TODO: store the LDAP client?
	registry *types.Registry
}

// LibraryName implements the SingletonLibrary interface method.
func (*demoLib) LibraryName() string {
	return "gitops.tools.lib.ext.demo"
}

// CompileOptions implements the Library interface method.
func (l *demoLib) CompileOptions() []cel.EnvOption {
	mapStrDyn := cel.MapType(cel.StringType, cel.DynType)

	opts := []cel.EnvOption{
		cel.Function("ldap_lookup",
			cel.Overload("ldap_lookup_string", []*cel.Type{cel.StringType}, mapStrDyn, cel.UnaryBinding(makeLDAPLookup(l)))),
	}

	return opts
}

// ProgramOptions implements the Library interface method.
func (*demoLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

// This should accept something that can do LDAP Lookups.
func makeLDAPLookup(lib *demoLib) functions.UnaryOp {
	return func(lhs ref.Val) ref.Val {
		decodedVal := map[string]any{
			"guid": "27f7c407-99d7-4c5a-8ebd-206a5c2e3f3d",
		}
		return types.NewDynamicMap(lib.registry, decodedVal)
	}
}
