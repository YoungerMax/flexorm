package studio

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed template.html
var templateHTML string

type Studio struct {
	db       *sql.DB
	driver   string
	template *template.Template
}

type TableInfo struct {
	Name     string       `json:"name"`
	RowCount int          `json:"rowCount"`
	Columns  []ColumnInfo `json:"columns"`
}

type ColumnInfo struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Nullable   bool   `json:"nullable"`
	PrimaryKey bool   `json:"primaryKey"`
}

type TableData struct {
	Columns []string                 `json:"columns"`
	Rows    []map[string]interface{} `json:"rows"`
	Total   int                      `json:"total"`
	Page    int                      `json:"page"`
	PerPage int                      `json:"perPage"`
}

func RunWebServer(host string, port int, databaseUrl string, driver string) error {
	db, err := sql.Open(driver, databaseUrl)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	tmpl, err := template.New("index").Parse(templateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	studio := &Studio{
		db:       db,
		driver:   driver,
		template: tmpl,
	}

	http.HandleFunc("/", studio.handleIndex)
	http.HandleFunc("/api/tables", studio.handleGetTables)
	http.HandleFunc("/api/table/", studio.handleGetTableData)

	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("🚀 FlexORM Studio running at http://%s\n", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Studio) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := s.template.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Studio) handleGetTables(w http.ResponseWriter, r *http.Request) {
	tables, err := s.getTables()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tables)
}

func (s *Studio) handleGetTableData(w http.ResponseWriter, r *http.Request) {
	tableName := strings.TrimPrefix(r.URL.Path, "/api/table/")
	if tableName == "" {
		http.Error(w, "table name required", http.StatusBadRequest)
		return
	}

	page := 1
	perPage := 50

	if p := r.URL.Query().Get("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	data, err := s.getTableData(tableName, page, perPage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (s *Studio) getTables() ([]TableInfo, error) {
	var query string

	if s.driver == "postgres" || s.driver == "postgresql" {
		query = `
			SELECT table_name 
			FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_type = 'BASE TABLE'
			ORDER BY table_name`
	} else {
		query = `
			SELECT name 
			FROM sqlite_master 
			WHERE type='table' 
			AND name NOT LIKE 'sqlite_%'
			ORDER BY name`
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}

		columns, err := s.getTableColumns(tableName)
		if err != nil {
			return nil, err
		}

		rowCount, err := s.getRowCount(tableName)
		if err != nil {
			rowCount = 0
		}

		tables = append(tables, TableInfo{
			Name:     tableName,
			RowCount: rowCount,
			Columns:  columns,
		})
	}

	return tables, nil
}

func (s *Studio) getTableColumns(tableName string) ([]ColumnInfo, error) {
	var query string

	if s.driver == "postgres" || s.driver == "postgresql" {
		query = fmt.Sprintf(`
			SELECT 
				c.column_name,
				c.data_type,
				c.is_nullable,
				CASE WHEN pk.constraint_type = 'PRIMARY KEY' THEN true ELSE false END as is_primary
			FROM information_schema.columns c
			LEFT JOIN (
				SELECT kcu.column_name, tc.constraint_type
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu
					ON tc.constraint_name = kcu.constraint_name
					AND tc.table_schema = kcu.table_schema
				WHERE tc.table_name = '%s' 
				AND tc.constraint_type = 'PRIMARY KEY'
			) pk ON pk.column_name = c.column_name
			WHERE c.table_name = '%s'
			ORDER BY c.ordinal_position`, tableName, tableName)
	} else {
		query = fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	}

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo

	if s.driver == "postgres" || s.driver == "postgresql" {
		for rows.Next() {
			var col ColumnInfo
			var nullable string
			var isPK *bool

			if err := rows.Scan(&col.Name, &col.Type, &nullable, &isPK); err != nil {
				return nil, err
			}

			col.Nullable = nullable == "YES"
			if isPK != nil {
				col.PrimaryKey = *isPK
			}
			columns = append(columns, col)
		}
	} else {
		for rows.Next() {
			var cid int
			var col ColumnInfo
			var notNull int
			var dfltValue *string
			var pk int

			if err := rows.Scan(&cid, &col.Name, &col.Type, &notNull, &dfltValue, &pk); err != nil {
				return nil, err
			}

			col.Nullable = notNull == 0
			col.PrimaryKey = pk > 0
			columns = append(columns, col)
		}
	}

	return columns, nil
}

func (s *Studio) getRowCount(tableName string) (int, error) {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := s.db.QueryRow(query).Scan(&count)
	return count, err
}

func (s *Studio) getTableData(tableName string, page, perPage int) (*TableData, error) {
	// Get columns first
	columns, err := s.getTableColumns(tableName)
	if err != nil {
		return nil, err
	}

	columnNames := make([]string, len(columns))
	for i, col := range columns {
		columnNames[i] = col.Name
	}

	// Get total count
	total, err := s.getRowCount(tableName)
	if err != nil {
		return nil, err
	}

	// Get paginated data
	offset := (page - 1) * perPage
	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d OFFSET %d", tableName, perPage, offset)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		data = append(data, row)
	}

	return &TableData{
		Columns: columnNames,
		Rows:    data,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}, nil
}
