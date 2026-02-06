package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// TSVParser reads tab-separated values files (D2 data format)
type TSVParser struct {
	headers     []string
	headerIndex map[string]int
	rows        []map[string]string
}

// ParseFile reads a TSV file and returns a parser with the data
func ParseFile(filepath string) (*TSVParser, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filepath, err)
	}
	defer file.Close()

	parser := &TSVParser{
		headerIndex: make(map[string]int),
		rows:        make([]map[string]string, 0),
	}

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for large lines

	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, "\t")

		if lineNum == 1 {
			// First line is headers
			parser.headers = fields
			for i, header := range fields {
				parser.headerIndex[header] = i
			}
			continue
		}

		// Parse data row
		row := make(map[string]string)
		for i, field := range fields {
			if i < len(parser.headers) {
				row[parser.headers[i]] = strings.TrimSpace(field)
			}
		}

		// Skip rows with empty first field (usually indicates comment or empty row)
		if len(fields) > 0 && strings.TrimSpace(fields[0]) != "" {
			parser.rows = append(parser.rows, row)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filepath, err)
	}

	return parser, nil
}

// Rows returns all parsed rows
func (p *TSVParser) Rows() []map[string]string {
	return p.rows
}

// Headers returns the column headers
func (p *TSVParser) Headers() []string {
	return p.headers
}

// HasColumn checks if a column exists
func (p *TSVParser) HasColumn(name string) bool {
	_, ok := p.headerIndex[name]
	return ok
}

// Row represents a single row with helper methods
type Row map[string]string

// GetString returns a string value or default
func (r Row) GetString(key, defaultValue string) string {
	if val, ok := r[key]; ok && val != "" {
		return val
	}
	return defaultValue
}

// GetInt returns an integer value or default
func (r Row) GetInt(key string, defaultValue int) int {
	if val, ok := r[key]; ok && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultValue
}

// GetBool returns a boolean value (1 = true, 0 or empty = false)
func (r Row) GetBool(key string) bool {
	if val, ok := r[key]; ok && val != "" {
		return val == "1" || strings.ToLower(val) == "true"
	}
	return false
}

// GetFloat returns a float value or default
func (r Row) GetFloat(key string, defaultValue float64) float64 {
	if val, ok := r[key]; ok && val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return defaultValue
}

// IsEmpty checks if a field is empty
func (r Row) IsEmpty(key string) bool {
	val, ok := r[key]
	return !ok || val == ""
}

// AsRow converts a map to Row for helper methods
func AsRow(m map[string]string) Row {
	return Row(m)
}
