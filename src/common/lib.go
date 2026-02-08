package common

import (
	"encoding/json"
	"os"
)

type Schema struct {
	Tables []Table `json:"tables"`
}

func (s Schema) GetTableByName(name string) *Table {
	for i := range s.Tables {
		if s.Tables[i].Name == name {
			return &s.Tables[i]
		}
	}

	return nil
}

type Table struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
}

func (t Table) GetColumnByName(name string) *Column {
	for i := range t.Columns {
		if t.Columns[i].Name == name {
			return &t.Columns[i]
		}
	}

	return nil
}

type ColumnType string

var (
	// Numeric types
	SmallInt        ColumnType = "smallint"
	Integer         ColumnType = "integer"
	BigInt          ColumnType = "bigint"
	Serial          ColumnType = "serial"
	SmallSerial     ColumnType = "smallserial"
	BigSerial       ColumnType = "bigserial"
	Numeric         ColumnType = "numeric"
	Real            ColumnType = "real"
	DoublePrecision ColumnType = "double precision"

	// String types
	Char    ColumnType = "char"
	Varchar ColumnType = "varchar"
	Text    ColumnType = "text"

	// Boolean type
	Boolean ColumnType = "boolean"

	// JSON types
	JSON  ColumnType = "json"
	JSONB ColumnType = "jsonb"

	// UUID type
	UUID ColumnType = "uuid"

	// Date/Time types
	Date      ColumnType = "date"
	Time      ColumnType = "time"
	Timestamp ColumnType = "timestamp"
	Interval  ColumnType = "interval"

	// Geometric types
	Point ColumnType = "point"
	Line  ColumnType = "line"

	// Enum type
	Enum ColumnType = "enum"
)

type Column struct {
	Name        string      `json:"name"`
	Type        ColumnType  `json:"type"`
	PrimaryKey  bool        `json:"primaryKey"`
	Default     interface{} `json:"default"`
	Length      *int        `json:"length"`
	MinLength   *int        `json:"minLength"`
	MaxLength   *int        `json:"maxLength"`
	Pattern     *string     `json:"pattern"`
	EnumOptions []string    `json:"enumOptions"`
	Nullable    bool        `json:"nullable"`
}

// Helper Functions
var EmptySchema = Schema{Tables: []Table{}}

func LoadSchema(filepath string) (*Schema, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var schema Schema
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&schema); err != nil {
		return nil, err
	}

	return &schema, nil
}
