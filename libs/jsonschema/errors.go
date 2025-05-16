package jsonschema

import "errors"

var (
	ErrSchemaNotFound = errors.New("schema not found")
	ErrValidation     = errors.New("data does not conform to schema")
)
