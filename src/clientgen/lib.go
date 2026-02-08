package clientgen

import (
	"flexorm/clientgen/v2/java_gson_postgres_jdbc"
	"flexorm/clientgen/v2/typescript_postgresjs"
	"flexorm/common/v2"
	"fmt"
	"io"
)

type Target string

type EmitFunc func(schema common.Schema, w io.Writer) error

const (
	TypeScriptPostgresJS Target = "typescript_postgresjs"
	JavaGsonPostgresJdbc Target = "java_gson_postgres_jdbc"
)

// registry holds all emitter functions mapped to their targets
var registry = map[Target]EmitFunc{
	TypeScriptPostgresJS: typescript_postgresjs.Emit,
	JavaGsonPostgresJdbc: java_gson_postgres_jdbc.Emit,
}

// Emit generates code for the given schema and target
func Emit(schema common.Schema, target Target, w io.Writer) error {
	emitFunc, exists := registry[target]

	if !exists {
		return fmt.Errorf("unsupported target: %s", target)
	}

	return emitFunc(schema, w)
}

// ListTargets returns all registered targets
func ListTargets() []Target {
	targets := make([]Target, 0, len(registry))
	for target := range registry {
		targets = append(targets, target)
	}
	return targets
}
