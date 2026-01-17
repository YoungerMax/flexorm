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

//go:embed validators_boolean.gots
var validatorsBooleanTmpl string

//go:embed validators_real.gots
var validatorsRealTmpl string

//go:embed validators_double_precision.gots
var validatorsDoublePrecisionTmpl string

//go:embed validators_varchar.gots
var validatorsVarcharTmpl string

//go:embed validators_char.gots
var validatorsCharTmpl string

//go:embed validators_json.gots
var validatorsJsonTmpl string

//go:embed validators_uuid.gots
var validatorsUuidTmpl string

//go:embed validators_date.gots
var validatorsDateTmpl string

//go:embed validators_time.gots
var validatorsTimeTmpl string

//go:embed validators_interval.gots
var validatorsIntervalTmpl string

//go:embed validators_point.gots
var validatorsPointTmpl string

//go:embed validators_line.gots
var validatorsLineTmpl string

//go:embed validators_enum.gots
var validatorsEnumTmpl string

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
		"typescriptName":       typescriptName,
		"typescriptType":       typescriptType,
		"typescriptTypeColumn": typescriptTypeColumn,
		"typescriptTableName":  typescriptTableName,
		"renderValidator":      renderValidator,
	}

	mustParse := func(name, tmpl string) *template.Template {
		return template.Must(template.New(name).Funcs(funcMap).Parse(tmpl))
	}

	typeMappings = map[common.ColumnType]*TypeMapping{
		// String types
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
		common.Varchar: {
			TSType:            "string",
			ValidatorTemplate: mustParse("validators_varchar", validatorsVarcharTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":      colName,
					"Column":       col,
					"HasDefault":   col.Default != "",
					"HasMaxLength": col.MaxLength != nil && *col.MaxLength > 0,
					"MaxLength":    col.MaxLength,
					"HasPattern":   col.Pattern != nil && *col.Pattern != "",
				}
			},
		},
		common.Char: {
			TSType:            "string",
			ValidatorTemplate: mustParse("validators_char", validatorsCharTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
					"HasLength":  col.Length != nil && *col.Length > 0,
					"Length":     col.Length,
				}
			},
		},

		// Numeric types
		common.SmallInt: {
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
		common.BigInt: {
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
		common.Serial: {
			TSType:            "number",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"Column":          col,
					"IsAutoIncrement": true,
					"HasDefault":      true,
				}
			},
		},
		common.SmallSerial: {
			TSType:            "number",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"Column":          col,
					"IsAutoIncrement": true,
					"HasDefault":      true,
				}
			},
		},
		common.BigSerial: {
			TSType:            "number",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"Column":          col,
					"IsAutoIncrement": true,
					"HasDefault":      true,
				}
			},
		},
		common.Numeric: {
			TSType:            "number",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},
		common.Real: {
			TSType:            "number",
			ValidatorTemplate: mustParse("validators_real", validatorsRealTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},
		common.DoublePrecision: {
			TSType:            "number",
			ValidatorTemplate: mustParse("validators_double_precision", validatorsDoublePrecisionTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},

		// Boolean type
		common.Boolean: {
			TSType:            "boolean",
			ValidatorTemplate: mustParse("validators_boolean", validatorsBooleanTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},

		// JSON types
		common.JSON: {
			TSType:            "any",
			ValidatorTemplate: mustParse("validators_json", validatorsJsonTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},
		common.JSONB: {
			TSType:            "any",
			ValidatorTemplate: mustParse("validators_json", validatorsJsonTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},

		// UUID type
		common.UUID: {
			TSType:            "string",
			ValidatorTemplate: mustParse("validators_uuid", validatorsUuidTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},

		// Date/Time types
		common.Date: {
			TSType:            "Date | string",
			ValidatorTemplate: mustParse("validators_date", validatorsDateTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},
		common.Time: {
			TSType:            "string",
			ValidatorTemplate: mustParse("validators_time", validatorsTimeTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
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
		common.Interval: {
			TSType:            "string",
			ValidatorTemplate: mustParse("validators_interval", validatorsIntervalTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},

		// Geometric types
		common.Point: {
			TSType:            "Point",
			ValidatorTemplate: mustParse("validators_point", validatorsPointTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},
		common.Line: {
			TSType:            "Line",
			ValidatorTemplate: mustParse("validators_line", validatorsLineTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"Column":     col,
					"HasDefault": col.Default != "",
				}
			},
		},

		// Enum type
		common.Enum: {
			TSType:            "string",
			ValidatorTemplate: mustParse("validators_enum", validatorsEnumTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":        colName,
					"Column":         col,
					"HasDefault":     col.Default != "",
					"HasEnumOptions": len(col.EnumOptions) > 0,
					"EnumOptions":    col.EnumOptions,
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

func typescriptTypeColumn(column common.Column) string {
	// Special handling for enum types to generate union types
	if column.Type == common.Enum && len(column.EnumOptions) > 0 {
		var opts []string
		for _, opt := range column.EnumOptions {
			opts = append(opts, "'"+opt+"'")
		}
		return strings.Join(opts, " | ")
	}

	// Fall back to regular type mapping for non-enum types
	return typescriptType(column.Type)
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
