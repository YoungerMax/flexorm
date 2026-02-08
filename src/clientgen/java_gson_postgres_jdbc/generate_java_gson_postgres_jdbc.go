package java_gson_postgres_jdbc

import (
	_ "embed"
	"flexorm/common/v2"
	"io"
	"strings"
	"text/template"
)

//go:embed template.gojava
var coreTmpl string

//go:embed validators_text.gojava
var validatorsTextTmpl string

//go:embed validators_integer.gojava
var validatorsIntegerTmpl string

//go:embed validators_timestamp.gojava
var validatorsTimestampTmpl string

//go:embed validators_boolean.gojava
var validatorsBooleanTmpl string

//go:embed validators_real.gojava
var validatorsRealTmpl string

//go:embed validators_double_precision.gojava
var validatorsDoublePrecisionTmpl string

//go:embed validators_varchar.gojava
var validatorsVarcharTmpl string

//go:embed validators_char.gojava
var validatorsCharTmpl string

//go:embed validators_json.gojava
var validatorsJsonTmpl string

//go:embed validators_uuid.gojava
var validatorsUuidTmpl string

//go:embed validators_date.gojava
var validatorsDateTmpl string

//go:embed validators_time.gojava
var validatorsTimeTmpl string

//go:embed validators_interval.gojava
var validatorsIntervalTmpl string

//go:embed validators_point.gojava
var validatorsPointTmpl string

//go:embed validators_line.gojava
var validatorsLineTmpl string

//go:embed validators_enum.gojava
var validatorsEnumTmpl string

// TypeMapping holds the Java type and its validator template together
type TypeMapping struct {
	JavaType          string
	ValidatorTemplate *template.Template
	BuildValidator    func(colName string, col common.Column) map[string]interface{}
}

var typeMappings map[common.ColumnType]*TypeMapping
var coreTemplate *template.Template

func initializeTypeMappings() {
	funcMap := template.FuncMap{
		"javaClassName":      javaClassName,
		"javaType":           javaType,
		"javaTypeColumn":     javaTypeColumn,
		"javaFieldName":      javaFieldName,
		"javaTableClassName": javaTableClassName,
		"renderValidator":    renderValidator,
		"javaIsPrimitive":    javaIsPrimitive,
		"javaBoxedType":      javaBoxedType,
		"genEnumValues":      genEnumValues,
	}

	mustParse := func(name, tmpl string) *template.Template {
		return template.Must(template.New(name).Funcs(funcMap).Parse(tmpl))
	}

	typeMappings = map[common.ColumnType]*TypeMapping{
		// String types
		common.Text: {
			JavaType:          "String",
			ValidatorTemplate: mustParse("validators_text", validatorsTextTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":      colName,
					"FieldClass":   "String",
					"Nullable":     col.Nullable,
					"HasDefault":   col.Default != "" && col.Default != nil,
					"HasLength":    col.Length != nil && *col.Length > 0,
					"Length":       col.Length,
					"HasMinLength": col.MinLength != nil && *col.MinLength > 0,
					"MinLength":    col.MinLength,
					"HasMaxLength": col.MaxLength != nil && *col.MaxLength > 0,
					"MaxLength":    col.MaxLength,
					"HasPattern":   col.Pattern != nil && *col.Pattern != "",
					"Pattern":      col.Pattern,
				}
			},
		},
		common.Varchar: {
			JavaType:          "String",
			ValidatorTemplate: mustParse("validators_varchar", validatorsVarcharTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":      colName,
					"FieldClass":   "String",
					"Nullable":     col.Nullable,
					"HasDefault":   col.Default != "" && col.Default != nil,
					"HasMaxLength": col.MaxLength != nil && *col.MaxLength > 0,
					"MaxLength":    col.MaxLength,
					"HasPattern":   col.Pattern != nil && *col.Pattern != "",
					"Pattern":      col.Pattern,
				}
			},
		},
		common.Char: {
			JavaType:          "String",
			ValidatorTemplate: mustParse("validators_char", validatorsCharTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "String",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
					"HasLength":  col.Length != nil && *col.Length > 0,
					"Length":     col.Length,
				}
			},
		},

		// Numeric types
		common.SmallInt: {
			JavaType:          "Short",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"FieldClass":      "Short",
					"Nullable":        col.Nullable,
					"IsAutoIncrement": isAutoIncrement(col.Default),
					"HasDefault":      col.Default != "" && col.Default != nil,
				}
			},
		},
		common.Integer: {
			JavaType:          "Integer",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"FieldClass":      "Integer",
					"Nullable":        col.Nullable,
					"IsAutoIncrement": isAutoIncrement(col.Default),
					"HasDefault":      col.Default != "" && col.Default != nil,
				}
			},
		},
		common.BigInt: {
			JavaType:          "Long",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"FieldClass":      "Long",
					"Nullable":        col.Nullable,
					"IsAutoIncrement": isAutoIncrement(col.Default),
					"HasDefault":      col.Default != "" && col.Default != nil,
				}
			},
		},
		common.Serial: {
			JavaType:          "Integer",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"FieldClass":      "Integer",
					"Nullable":        col.Nullable,
					"IsAutoIncrement": true,
					"HasDefault":      true,
				}
			},
		},
		common.SmallSerial: {
			JavaType:          "Short",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"FieldClass":      "Short",
					"Nullable":        col.Nullable,
					"IsAutoIncrement": true,
					"HasDefault":      true,
				}
			},
		},
		common.BigSerial: {
			JavaType:          "Long",
			ValidatorTemplate: mustParse("validators_integer", validatorsIntegerTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":         colName,
					"FieldClass":      "Long",
					"Nullable":        col.Nullable,
					"IsAutoIncrement": true,
					"HasDefault":      true,
				}
			},
		},
		common.Numeric: {
			JavaType:          "java.math.BigDecimal",
			ValidatorTemplate: mustParse("validators_numeric", validatorsRealTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "java.math.BigDecimal",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},
		common.Real: {
			JavaType:          "Float",
			ValidatorTemplate: mustParse("validators_real", validatorsRealTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "Float",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},
		common.DoublePrecision: {
			JavaType:          "Double",
			ValidatorTemplate: mustParse("validators_double_precision", validatorsDoublePrecisionTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "Double",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},

		// Boolean type
		common.Boolean: {
			JavaType:          "Boolean",
			ValidatorTemplate: mustParse("validators_boolean", validatorsBooleanTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "Boolean",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},

		// JSON types
		common.JSON: {
			JavaType:          "com.google.gson.JsonElement",
			ValidatorTemplate: mustParse("validators_json", validatorsJsonTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "com.google.gson.JsonElement",
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},
		common.JSONB: {
			JavaType:          "com.google.gson.JsonElement",
			ValidatorTemplate: mustParse("validators_json", validatorsJsonTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "com.google.gson.JsonElement",
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},

		// UUID type
		common.UUID: {
			JavaType:          "java.util.UUID",
			ValidatorTemplate: mustParse("validators_uuid", validatorsUuidTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "java.util.UUID",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},

		// Date/Time types
		common.Date: {
			JavaType:          "java.time.LocalDate",
			ValidatorTemplate: mustParse("validators_date", validatorsDateTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "java.time.LocalDate",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},
		common.Time: {
			JavaType:          "java.time.LocalTime",
			ValidatorTemplate: mustParse("validators_time", validatorsTimeTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "java.time.LocalTime",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},
		common.Timestamp: {
			JavaType:          "java.time.Instant",
			ValidatorTemplate: mustParse("validators_timestamp", validatorsTimestampTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "java.time.Instant",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},
		common.Interval: {
			JavaType:          "String",
			ValidatorTemplate: mustParse("validators_interval", validatorsIntervalTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "String",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},

		// Geometric types
		common.Point: {
			JavaType:          "Point",
			ValidatorTemplate: mustParse("validators_point", validatorsPointTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "Point",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},
		common.Line: {
			JavaType:          "Line",
			ValidatorTemplate: mustParse("validators_line", validatorsLineTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":    colName,
					"FieldClass": "Line",
					"Nullable":   col.Nullable,
					"HasDefault": col.Default != "" && col.Default != nil,
				}
			},
		},

		// Enum type
		common.Enum: {
			JavaType:          "String",
			ValidatorTemplate: mustParse("validators_enum", validatorsEnumTmpl),
			BuildValidator: func(colName string, col common.Column) map[string]interface{} {
				return map[string]interface{}{
					"ColName":        colName,
					"FieldClass":     "String",
					"Nullable":       col.Nullable,
					"HasDefault":     col.Default != "" && col.Default != nil,
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

func isAutoIncrement(defaultValue interface{}) bool {
	if str, ok := defaultValue.(string); ok {
		return str == "autoincrement"
	}
	return false
}

func javaType(dbType common.ColumnType) string {
	if mapping, exists := typeMappings[dbType]; exists {
		return mapping.JavaType
	}
	return "Object"
}

func javaIsPrimitive(javaType string) bool {
	switch javaType {
	case "short", "int", "long", "double", "float", "boolean":
		return true
	default:
		return false
	}
}

func javaBoxedType(javaType string) string {
	switch javaType {
	case "short":
		return "Short"
	case "int":
		return "Integer"
	case "long":
		return "Long"
	case "double":
		return "Double"
	case "float":
		return "Float"
	case "boolean":
		return "Boolean"
	default:
		return javaType
	}
}

func javaTypeColumn(table common.Table, column common.Column) string {
	if column.Type == common.Enum && len(column.EnumOptions) > 0 {
		return javaClassName(table.Name) + "." + javaClassName(column.Name)
	}
	return javaType(column.Type)
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

func javaClassName(schemaTableName string) string {
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

func javaFieldName(schemaColumnName string) string {
	if schemaColumnName == "" {
		return ""
	}

	parts := strings.Split(schemaColumnName, "_")

	if len(parts) == 0 {
		return ""
	}

	// First part is lowercase
	var result strings.Builder
	result.WriteString(strings.ToLower(parts[0]))

	// Subsequent parts are capitalized
	for i := 1; i < len(parts); i++ {
		part := parts[i]
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

func javaTableClassName(schemaTableName string) string {
	return javaClassName(schemaTableName) + "Table"
}

func genEnumValues(col common.Column) string {
	var result strings.Builder

	for i, opt := range col.EnumOptions {
		result.WriteString(opt)

		if i == len(col.EnumOptions)-1 {
			result.WriteString(";\n")
		} else {
			result.WriteString(",\n")
		}
	}

	return result.String()
}
