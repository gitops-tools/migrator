package celpatch

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	celext "github.com/google/cel-go/ext"
	"github.com/tidwall/sjson"
)

// Change is used to describe a way to modify a key in the resource using a
// CEL expression.
type Change struct {
	Key      string `json:"key"`
	NewValue string `json:"newValue"`
}

// WithCELLib adds a CEL library to the underlying CEL parser.
func WithCELLib(l cel.EnvOption) func(*celOptions) {
	return func(opt *celOptions) {
		opt.Libraries = append(opt.Libraries, l)
	}
}

type celOptions struct {
	Libraries []cel.EnvOption
}

type celOptionFunc func(*celOptions)

// ApplyChanges applies a set of Changes to a JSON representation of a resource
// and returns a copy with the changes applied.
func ApplyChanges(b []byte, changes []Change, opts ...celOptionFunc) ([]byte, error) {
	changed := map[string]string{}

	options := &celOptions{}
	for _, opt := range opts {
		opt(options)
	}

	env, err := makeCELEnv(options)
	if err != nil {
		return nil, fmt.Errorf("failed to setup CEL environment: %w", err)
	}

	var data map[string]any
	if err := json.Unmarshal(b, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal resource: %w", err)
	}

	for _, change := range changes {
		newValue, err := evaluate(change.NewValue, env, map[string]any{
			"resource": data})
		if err != nil {
			// TODO: multi-error
			return nil, err
		}

		changed[change.Key] = newValue
	}

	size := len(b)
	updated := make([]byte, size)
	if n := copy(updated, b); n != size {
		return nil, errors.New("failed to copy body for updating")
	}

	for k, v := range changed {
		updated, err = sjson.SetBytes(updated, k, v)
		if err != nil {
			return nil, err // TODO
		}
	}

	return updated, nil
}

func evaluate(expr string, env *cel.Env, data map[string]any) (string, error) {
	parsed, issues := env.Parse(expr)
	if issues != nil && issues.Err() != nil {
		return "", fmt.Errorf("failed to parse expression %q: %w", expr, issues.Err())
	}

	checked, issues := env.Check(parsed)
	if issues != nil && issues.Err() != nil {
		return "", fmt.Errorf("expression %v check failed: %w", expr, issues.Err())
	}

	prg, err := env.Program(checked, cel.EvalOptions(cel.OptOptimize))
	if err != nil {
		return "", fmt.Errorf("expression %v failed to create a Program: %w", expr, err)
	}

	out, _, err := prg.Eval(data)
	if err != nil {
		return "", fmt.Errorf("expression %v failed to evaluate: %w", expr, err)
	}

	if v, ok := out.(types.String); ok {
		return v.Value().(string), nil
	}

	return "", fmt.Errorf("expression %q did not evaluate to a string", expr)
}

func makeCELEnv(options *celOptions) (*cel.Env, error) {
	mapStrDyn := decls.NewMapType(decls.String, decls.Dyn)

	envOptions := []cel.EnvOption{
		celext.Strings(),
		celext.Encoders(),
		cel.Declarations(
			decls.NewVar("resource", mapStrDyn),
		),
	}
	envOptions = append(envOptions, options.Libraries...)

	return cel.NewEnv(
		envOptions...,
	)
}
