package jsonschema

func Validate(schemaName string, data any) error {
	schema, err := LoadSchema(schemaName)
	if err != nil {
		return ErrSchemaNotFound
	}
	// 將 data 轉為 JSON bytes 然後與 schema 驗證
	// 使用 github.com/xeipuuv/gojsonschema
}
