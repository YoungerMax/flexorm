package migrate

import (
	"flexorm/common/v2"
	"fmt"
	"strings"
)

// typeMap maps common.ColumnType to PostgreSQL type strings
var typeMap = map[common.ColumnType]string{
	common.SmallInt:        "SMALLINT",
	common.Integer:         "INTEGER",
	common.BigInt:          "BIGINT",
	common.Serial:          "SERIAL",
	common.SmallSerial:     "SMALLSERIAL",
	common.BigSerial:       "BIGSERIAL",
	common.Char:            "CHAR",
	common.Varchar:         "VARCHAR",
	common.Text:            "TEXT",
	common.Numeric:         "NUMERIC",
	common.Real:            "REAL",
	common.DoublePrecision: "DOUBLE PRECISION",
	common.Boolean:         "BOOLEAN",
	common.JSON:            "JSON",
	common.JSONB:           "JSONB",
	common.UUID:            "UUID",
	common.Date:            "DATE",
	common.Time:            "TIME",
	common.Timestamp:       "TIMESTAMP",
	common.Interval:        "INTERVAL",
	common.Point:           "POINT",
	common.Line:            "LINE",
	common.Enum:            "TEXT",
}

// numericTypes are types that support numeric defaults
var numericTypes = []common.ColumnType{
	common.Integer,
	common.SmallInt,
	common.BigInt,
	common.Numeric,
	common.Real,
	common.DoublePrecision,
}

// stringTypes are types that require quoted string defaults
var stringTypes = []common.ColumnType{
	common.Text,
	common.Char,
	common.Varchar,
	common.Enum,
	common.Date,
	common.Time,
	common.Interval,
	common.UUID,
}

// PostgreSQLGenerator generates SQL for PostgreSQL databases
type PostgreSQLGenerator struct{}

func (g PostgreSQLGenerator) GenerateSQL(diff *SchemaDiff) (upSQL, downSQL string) {
	var upStatements []string
	var downStatements []string

	for _, change := range diff.Changes {
		switch change.Type {
		case ChangeTypeCreateTable:
			details := change.Details.(TableCreate)
			upStatements = append(upStatements, g.generateCreateTable(change.TableName, details.Table))
			downStatements = append([]string{g.generateDropTable(change.TableName)}, downStatements...)

		case ChangeTypeDropTable:
			details := change.Details.(TableDrop)
			upStatements = append(upStatements, g.generateDropTable(change.TableName))
			downStatements = append([]string{g.generateCreateTable(change.TableName, details.Table)}, downStatements...)

		case ChangeTypeAddColumn:
			details := change.Details.(ColumnAdd)
			upStatements = append(upStatements, g.generateAddColumn(change.TableName, details.ColumnName, details.Column))
			downStatements = append([]string{g.generateDropColumn(change.TableName, details.ColumnName)}, downStatements...)

		case ChangeTypeDropColumn:
			details := change.Details.(ColumnDrop)
			upStatements = append(upStatements, g.generateDropColumn(change.TableName, details.ColumnName))
			downStatements = append([]string{g.generateAddColumn(change.TableName, details.ColumnName, details.Column)}, downStatements...)

		case ChangeTypeAlterColumn:
			details := change.Details.(ColumnAlter)
			upStatements = append(upStatements, g.generateAlterColumn(change.TableName, details.ColumnName, details.OldColumn, details.NewColumn))
			downStatements = append([]string{g.generateAlterColumn(change.TableName, details.ColumnName, details.NewColumn, details.OldColumn)}, downStatements...)
		}
	}

	return strings.Join(upStatements, "\n\n"), strings.Join(downStatements, "\n\n")
}

func (g PostgreSQLGenerator) generateCreateTable(tableName string, table common.Table) string {
	var columns []string

	for _, col := range table.Columns {
		colDef := fmt.Sprintf("  %s %s", col.Name, g.sqlType(col))

		if col.PrimaryKey {
			colDef += " PRIMARY KEY"
		}

		if col.Default != nil && !col.PrimaryKey {
			colDef += g.formatDefault(col)
		}

		columns = append(columns, colDef)
	}

	// Add enum CHECK constraints after column definitions
	var constraints []string
	for _, col := range table.Columns {
		if col.Type == common.Enum && len(col.EnumOptions) > 0 {
			constraint := g.generateEnumCheckConstraint(tableName, col.Name, col.EnumOptions)
			constraints = append(constraints, constraint)
		}
	}

	// Combine columns and constraints
	allDefs := columns
	if len(constraints) > 0 {
		allDefs = append(allDefs, constraints...)
	}

	return fmt.Sprintf("CREATE TABLE %s (\n%s\n);", tableName, strings.Join(allDefs, ",\n"))
}

func (g PostgreSQLGenerator) generateDropTable(tableName string) string {
	return fmt.Sprintf("DROP TABLE %s;", tableName)
}

func (g PostgreSQLGenerator) generateAddColumn(tableName, columnName string, column common.Column) string {
	colDef := fmt.Sprintf("%s %s", columnName, g.sqlType(column))
	if column.Default != nil {
		colDef += g.formatDefault(column)
	}
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s;", tableName, colDef)
}

func (g PostgreSQLGenerator) generateDropColumn(tableName, columnName string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;", tableName, columnName)
}

func (g PostgreSQLGenerator) generateAlterColumn(tableName, columnName string, oldColumn, newColumn common.Column) string {
	var statements []string

	if oldColumn.Type != newColumn.Type {
		statements = append(statements,
			fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;",
				tableName, columnName, g.sqlType(newColumn)))
	}

	if oldColumn.Default != newColumn.Default {
		if newColumn.Default != nil {
			statements = append(statements,
				fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET%s;",
					tableName, columnName, g.formatDefault(newColumn)))
		} else {
			statements = append(statements,
				fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;",
					tableName, columnName))
		}
	}

	return strings.Join(statements, "\n")
}

func (g PostgreSQLGenerator) sqlType(col common.Column) string {
	baseType, exists := typeMap[col.Type]
	if !exists {
		panic("bad column type: " + string(col.Type))
	}

	// Handle length-based types
	switch col.Type {
	case common.Char:
		if col.Length != nil && *col.Length > 0 {
			return fmt.Sprintf("CHAR(%d)", *col.Length)
		}
	case common.Varchar:
		if col.MaxLength != nil && *col.MaxLength > 0 {
			return fmt.Sprintf("VARCHAR(%d)", *col.MaxLength)
		}
	case common.Integer:
		if col.PrimaryKey && col.Default == "autoincrement" {
			return "SERIAL"
		}
	}

	return baseType
}

func (g PostgreSQLGenerator) formatDefault(col common.Column) string {
	if col.Default == nil {
		return ""
	}

	switch col.Type {
	case common.Timestamp:
		if col.Default == "now" {
			return " DEFAULT CURRENT_TIMESTAMP"
		}
	case common.Integer, common.SmallInt, common.BigInt:
		if col.Default == "autoincrement" {
			return ""
		}
		return fmt.Sprintf(" DEFAULT %v", col.Default)
	case common.Serial, common.SmallSerial, common.BigSerial:
		return ""
	}

	// Handle quoted string types
	if g.isInSlice(col.Type, stringTypes) {
		if str, ok := col.Default.(string); ok {
			return fmt.Sprintf(" DEFAULT '%s'", str)
		}
		// Handle POINT and LINE defaults (they're objects in JSON but need special formatting)
		if col.Type == common.Point || col.Type == common.Line {
			return fmt.Sprintf(" DEFAULT '%s'", g.formatGeometricDefault(col))
		}
	}

	// Handle numeric types
	if g.isInSlice(col.Type, numericTypes) {
		return fmt.Sprintf(" DEFAULT %v", col.Default)
	}

	// Handle JSON types
	if col.Type == common.JSON || col.Type == common.JSONB {
		return fmt.Sprintf(" DEFAULT '%v'", col.Default)
	}

	// Handle boolean
	if col.Type == common.Boolean {
		return fmt.Sprintf(" DEFAULT %v", col.Default)
	}

	// Handle POINT and LINE types with special formatting
	if col.Type == common.Point || col.Type == common.Line {
		return fmt.Sprintf(" DEFAULT '%s'", g.formatGeometricDefault(col))
	}

	return fmt.Sprintf(" DEFAULT %v", col.Default)
}

func (g PostgreSQLGenerator) generateEnumCheckConstraint(tableName, columnName string, enumOptions []string) string {
	// Format enum options as quoted strings for SQL
	var quotedOptions []string
	for _, option := range enumOptions {
		quotedOptions = append(quotedOptions, fmt.Sprintf("'%s'", option))
	}

	constraintName := fmt.Sprintf("%s_%s_check", tableName, columnName)
	optionsList := strings.Join(quotedOptions, ", ")

	return fmt.Sprintf("  CONSTRAINT %s CHECK (%s IN (%s))", constraintName, columnName, optionsList)
}

func (g PostgreSQLGenerator) formatGeometricDefault(col common.Column) string {
	// Try to convert to map[string]interface{} for POINT and LINE
	if defaultMap, ok := col.Default.(map[string]interface{}); ok {
		if col.Type == common.Point {
			x, xOk := defaultMap["x"]
			y, yOk := defaultMap["y"]
			if xOk && yOk {
				return fmt.Sprintf("(%v,%v)", x, y)
			}
		} else if col.Type == common.Line {
			a, aOk := defaultMap["a"]
			b, bOk := defaultMap["b"]
			c, cOk := defaultMap["c"]
			if aOk && bOk && cOk {
				return fmt.Sprintf("{%v,%v,%v}", a, b, c)
			}
		}
	}

	// Fallback: try to convert to string and return as-is
	return fmt.Sprintf("%v", col.Default)
}

func (g PostgreSQLGenerator) isInSlice(colType common.ColumnType, slice []common.ColumnType) bool {
	for _, t := range slice {
		if t == colType {
			return true
		}
	}
	return false
}
