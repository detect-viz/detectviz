package jsonschema

import (
	"embed"
)

var (
	//go:embed schemas/*
	schemaFS embed.FS
)

func LoadSchema(name string) ([]byte, error) {
	return schemaFS.ReadFile("schemas/" + name + ".json")
}
