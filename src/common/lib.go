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
	Text      ColumnType = "text"
	Integer   ColumnType = "integer"
	Timestamp ColumnType = "timestamp"
)

type Column struct {
	Name       string      `json:"name"`
	Type       ColumnType  `json:"type"`
	PrimaryKey bool        `json:"primaryKey"`
	Default    interface{} `json:"default"`
	Length     *int        `json:"length"`
	MinLength  *int        `json:"minLength"`
	MaxLength  *int        `json:"maxLength"`
	Pattern    *string     `json:"pattern"`
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
