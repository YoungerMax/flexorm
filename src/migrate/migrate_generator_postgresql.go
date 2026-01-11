package migrate

import (
	"flexorm/common/v2"
	"fmt"
	"strings"
)

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

	return fmt.Sprintf("CREATE TABLE %s (\n%s\n);", tableName, strings.Join(columns, ",\n"))
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
	switch col.Type {
	case "integer":
		if col.PrimaryKey && col.Default == "autoincrement" {
			return "SERIAL"
		}
		return "INTEGER"
	case "text":
		return "TEXT"
	case "timestamp":
		return "TIMESTAMP"
	default:
		// TODO: dont panic
		panic("bad column type: " + col.Type)
	}
}

func (g PostgreSQLGenerator) formatDefault(col common.Column) string {
	if col.Default == nil {
		return ""
	}

	switch col.Type {
	case "timestamp":
		if col.Default == "now" {
			return " DEFAULT CURRENT_TIMESTAMP"
		}
	case "integer":
		if col.Default == "autoincrement" {
			return ""
		}
		return fmt.Sprintf(" DEFAULT %v", col.Default)
	case "text":
		if str, ok := col.Default.(string); ok {
			return fmt.Sprintf(" DEFAULT '%s'", str)
		}
	}

	return fmt.Sprintf(" DEFAULT %v", col.Default)
}
