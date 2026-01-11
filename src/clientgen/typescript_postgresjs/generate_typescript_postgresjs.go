package typescript_postgresjs

import (
	_ "embed"
	"flexorm/common/v2"
	"io"
	"strings"
	"text/template"
)

//go:embed template.gots
var coreTmpl string

//go:embed validators_text.gots
var validatorsTextTmpl string

//go:embed validators_integer.gots
var validatorsIntegerTmpl string

//go:embed validators_timestamp.gots
var validatorsTimestampTmpl string

// TypeMapping holds the TypeScript type and its validator template together
type TypeMapping struct {
	TSType            string
	ValidatorTemplate *template.Template
	BuildValidator    func(colName string, col common.Column) map[string]interface{}
}

var typeMappings map[common.ColumnType]*TypeMapping
var coreTemplate *template.Template

func initializeTypeMappings() {
	funcMap := template.FuncMap{
		"typescriptName":      typescriptName,
		"typescriptType":      typescriptType,
		"typescriptTableName": typescriptTableName,
		"renderValidator":     renderValidator,
	}

	mustParse := func(name, tmpl string) *template.Template {
		return template.Must(template.New(name).Funcs(funcMap).Parse(tmpl))
	}

	typeMappings = map[common.ColumnType]*TypeMapping{
		common.Text: {
			TSType:            "string",
			ValidatorTemplate: mustParse("validators_text", validatorsTextTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":      colName,
					"Column":       col,
					"HasDefault":   col.Default != "",
					"HasLength":    col.Length != nil && *col.Length > 0,
					"Length":       col.Length,
					"HasMinLength": col.MinLength != nil && *col.MinLength > 0,
					"MinLength":    col.MinLength,
					"HasMaxLength": col.MaxLength != nil && *col.MaxLength > 0,
					"MaxLength":    col.MaxLength,
					"HasPattern":   col.Pattern != nil && *col.Pattern != "",
				}
			},
		},
		common.Integer: {
			TSType:            "number",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"Column":          col,
					"IsAutoIncrement": col.Default == "autoincrement",
					"HasDefault":      col.Default != "",
				}
			},
		},
		common.Timestamp: {
			TSType:            "Date",
			ValidatorTemplate: mustParse("validators_timestamp", validatorsTimestampTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},
	}

	coreTemplate = mustParse("core", coreTmpl)
}

func Emit(schema common.Schema, w io.Writer) error {
	initializeTypeMappings()

	return coreTemplate.Execute(w, schema)
}

func typescriptType(dbType common.ColumnType) string {
	if mapping, exists := typeMappings[dbType]; exists {
		return mapping.TSType
	}
	return "any"
}

func renderValidator(column common.Column) string {
	var buf strings.Builder

	mapping, exists := typeMappings[column.Type]
	if !exists || mapping.BuildValidator == nil {
		return ""
	}

	args := mapping.BuildValidator(column.Name, column)
	mapping.ValidatorTemplate.Execute(&buf, args)
	return buf.String()
}

func typescriptName(schemaTableName string) string {
	if schemaTableName == "" {
		return ""
	}

	parts := strings.Split(schemaTableName, "_")
	var result strings.Builder

	for _, part := range parts {
		if part == "" {
			continue
		}
		result.WriteString(strings.ToUpper(part[:1]))
		if len(part) > 1 {
			result.WriteString(strings.ToLower(part[1:]))
		}
	}

	return result.String()
}

func typescriptTableName(schemaTableName string) string {
	return typescriptName(schemaTableName) + "Table"
}
