package migrate

import (
	"flexorm/common/v2"
)

// GenerateDiff creates an AST of schema changes between two schemas
func GenerateDiff(oldSchema, newSchema common.Schema) *SchemaDiff {
	var changes []SchemaChange

	// Check for new and modified tables
	for _, newSchemaTable := range newSchema.Tables {
		oldSchemaTable := oldSchema.GetTableByName(newSchemaTable.Name)

		if oldSchemaTable == nil {
			// New table
			changes = append(changes, SchemaChange{
				Type:      ChangeTypeCreateTable,
				TableName: newSchemaTable.Name,
				Details:   TableCreate{Table: newSchemaTable},
			})
		} else {
			// Check for column changes
			colChanges := compareColumns(newSchemaTable.Name, *oldSchemaTable, newSchemaTable)
			changes = append(changes, colChanges...)
		}
	}

	// Check for dropped tables
	for _, oldSchemaTable := range oldSchema.Tables {
		newSchemaTable := newSchema.GetTableByName(oldSchemaTable.Name)

		if newSchemaTable == nil {
			changes = append(changes, SchemaChange{
				Type:      ChangeTypeDropTable,
				TableName: oldSchemaTable.Name,
				Details:   TableDrop{Table: oldSchemaTable},
			})
		}
	}

	return &SchemaDiff{Changes: changes}
}

func compareColumns(tableName string, oldTable, newTable common.Table) []SchemaChange {
	var changes []SchemaChange

	// Check for new columns
	for _, newTableColumn := range newTable.Columns {
		oldTableColumn := oldTable.GetColumnByName(newTableColumn.Name)

		// New column detected
		if oldTableColumn == nil {
			changes = append(changes, SchemaChange{
				Type:      ChangeTypeAddColumn,
				TableName: tableName,
				Details: ColumnAdd{
					ColumnName: newTableColumn.Name,
					Column:     newTableColumn,
				},
			})
		}
	}

	// Check for dropped columns
	for _, oldTableColumn := range oldTable.Columns {
		newTableColumn := newTable.GetColumnByName(oldTableColumn.Name)

		// Dropped column detected
		if newTableColumn == nil {
			changes = append(changes, SchemaChange{
				Type:      ChangeTypeDropColumn,
				TableName: tableName,
				Details: ColumnDrop{
					ColumnName: oldTableColumn.Name,
					Column:     oldTableColumn,
				},
			})
		}
	}

	return changes
}
