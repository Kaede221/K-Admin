package tools

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gorm.io/gorm"
)

type CodeGeneratorService struct {
	db *gorm.DB
}

func NewCodeGeneratorService(db *gorm.DB) *CodeGeneratorService {
	return &CodeGeneratorService{
		db: db,
	}
}

// FieldConfig represents a field configuration for code generation
type FieldConfig struct {
	ColumnName   string `json:"column_name"`
	FieldName    string `json:"field_name"`
	FieldType    string `json:"field_type"`
	JSONTag      string `json:"json_tag"`
	GormTag      string `json:"gorm_tag"`
	Comment      string `json:"comment"`
	TSType       string `json:"ts_type"`
	Label        string `json:"label"`
	FormType     string `json:"form_type"`
	Searchable   bool   `json:"searchable"`
	Nullable     bool   `json:"nullable"`
	IsPrimaryKey bool   `json:"is_primary_key"`
}

// GenerateConfig represents the configuration for code generation
type GenerateConfig struct {
	TableName    string          `json:"table_name"`
	StructName   string          `json:"struct_name"`
	PackageName  string          `json:"package_name"`
	FrontendPath string          `json:"frontend_path"`
	ModulePath   string          `json:"module_path"`
	Fields       []FieldConfig   `json:"fields"`
	Options      GenerateOptions `json:"options"`
	TableComment string          `json:"table_comment"`
	RouterPath   string          `json:"router_path"`
}

// GenerateOptions represents options for code generation
type GenerateOptions struct {
	GenerateModel         bool `json:"generate_model"`
	GenerateService       bool `json:"generate_service"`
	GenerateAPI           bool `json:"generate_api"`
	GenerateRouter        bool `json:"generate_router"`
	GenerateFrontendAPI   bool `json:"generate_frontend_api"`
	GenerateFrontendTypes bool `json:"generate_frontend_types"`
	GenerateFrontendPage  bool `json:"generate_frontend_page"`
}

// TableMetadata represents metadata extracted from a database table
type TableMetadata struct {
	TableName    string              `json:"table_name"`
	TableComment string              `json:"table_comment"`
	Columns      []CodeGenColumnInfo `json:"columns"`
}

// CodeGenColumnInfo represents information about a database column
type CodeGenColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Key      string `json:"key"`
	Default  string `json:"default"`
	Extra    string `json:"extra"`
	Comment  string `json:"comment"`
}

// GetTableMetadata extracts metadata from a database table
func (s *CodeGeneratorService) GetTableMetadata(tableName string) (*TableMetadata, error) {
	var columns []CodeGenColumnInfo

	query := `
		SELECT 
			COLUMN_NAME as name,
			COLUMN_TYPE as type,
			IS_NULLABLE = 'YES' as nullable,
			COLUMN_KEY as ` + "`key`" + `,
			COLUMN_DEFAULT as ` + "`default`" + `,
			EXTRA as extra,
			COLUMN_COMMENT as comment
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	if err := s.db.Raw(query, tableName).Scan(&columns).Error; err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("table %s not found", tableName)
	}

	// Get table comment
	var tableComment string
	commentQuery := `
		SELECT TABLE_COMMENT
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = ?
	`
	s.db.Raw(commentQuery, tableName).Scan(&tableComment)

	return &TableMetadata{
		TableName:    tableName,
		TableComment: tableComment,
		Columns:      columns,
	}, nil
}

// GenerateCode generates code based on the configuration
func (s *CodeGeneratorService) GenerateCode(config GenerateConfig) (map[string]string, error) {
	files := make(map[string]string)

	// Add helper fields to config
	config.RouterPath = strings.ToLower(strings.ReplaceAll(config.StructName, "_", "-"))

	// Generate backend files
	if config.Options.GenerateModel {
		content, err := s.generateFromTemplate("backend/model.tpl", config)
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("backend/model/%s/%s.go", config.PackageName, strings.ToLower(config.StructName))] = content
	}

	if config.Options.GenerateService {
		content, err := s.generateFromTemplate("backend/service.tpl", config)
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("backend/service/%s/%s_service.go", config.PackageName, strings.ToLower(config.StructName))] = content
	}

	if config.Options.GenerateAPI {
		content, err := s.generateFromTemplate("backend/api.tpl", config)
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("backend/api/v1/%s/%s.go", config.PackageName, strings.ToLower(config.StructName))] = content
	}

	if config.Options.GenerateRouter {
		content, err := s.generateFromTemplate("backend/router.tpl", config)
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("backend/router/%s/%s.go", config.PackageName, strings.ToLower(config.StructName))] = content
	}

	// Generate frontend files
	if config.Options.GenerateFrontendTypes {
		content, err := s.generateFromTemplate("frontend/types.tpl", config)
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("%s/api/%s/types.ts", config.FrontendPath, strings.ToLower(config.StructName))] = content
	}

	if config.Options.GenerateFrontendAPI {
		content, err := s.generateFromTemplate("frontend/api.tpl", config)
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("%s/api/%s/index.ts", config.FrontendPath, strings.ToLower(config.StructName))] = content
	}

	if config.Options.GenerateFrontendPage {
		// Generate page
		pageContent, err := s.generateFromTemplate("frontend/page.tpl", config)
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("%s/views/%s/index.tsx", config.FrontendPath, strings.ToLower(config.StructName))] = pageContent

		// Generate modal
		modalContent, err := s.generateFromTemplate("frontend/modal.tpl", config)
		if err != nil {
			return nil, err
		}
		files[fmt.Sprintf("%s/views/%s/components/%sModal.tsx", config.FrontendPath, strings.ToLower(config.StructName), config.StructName)] = modalContent
	}

	return files, nil
}

// PreviewCode generates code without writing to files
func (s *CodeGeneratorService) PreviewCode(config GenerateConfig) (map[string]string, error) {
	return s.GenerateCode(config)
}

// WriteGeneratedCode writes generated code to disk
func (s *CodeGeneratorService) WriteGeneratedCode(files map[string]string) error {
	for path, content := range files {
		// Create directory if it doesn't exist
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write file
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}
	}

	return nil
}

// CreateTable creates a new table from field definitions
func (s *CodeGeneratorService) CreateTable(tableName string, fields []FieldConfig) error {
	var sqlBuilder strings.Builder
	sqlBuilder.WriteString(fmt.Sprintf("CREATE TABLE `%s` (\n", tableName))
	sqlBuilder.WriteString("  `id` bigint unsigned NOT NULL AUTO_INCREMENT,\n")

	for i, field := range fields {
		sqlBuilder.WriteString(fmt.Sprintf("  `%s` %s", field.ColumnName, field.FieldType))

		if !field.Nullable {
			sqlBuilder.WriteString(" NOT NULL")
		}

		if field.Comment != "" {
			sqlBuilder.WriteString(fmt.Sprintf(" COMMENT '%s'", field.Comment))
		}

		if i < len(fields)-1 || true {
			sqlBuilder.WriteString(",\n")
		}
	}

	sqlBuilder.WriteString("  `created_at` datetime(3) DEFAULT NULL,\n")
	sqlBuilder.WriteString("  `updated_at` datetime(3) DEFAULT NULL,\n")
	sqlBuilder.WriteString("  `deleted_at` datetime(3) DEFAULT NULL,\n")
	sqlBuilder.WriteString("  PRIMARY KEY (`id`),\n")
	sqlBuilder.WriteString("  KEY `idx_deleted_at` (`deleted_at`)\n")
	sqlBuilder.WriteString(") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;")

	return s.db.Exec(sqlBuilder.String()).Error
}

// generateFromTemplate generates code from a template file
func (s *CodeGeneratorService) generateFromTemplate(templatePath string, config GenerateConfig) (string, error) {
	// Add helper field for lowercase struct name
	type TemplateData struct {
		GenerateConfig
		LowerStructName string
	}

	data := TemplateData{
		GenerateConfig:  config,
		LowerStructName: strings.ToLower(config.StructName[:1]) + config.StructName[1:],
	}

	// Read template file
	templateFile := filepath.Join("backend/resource/template", templatePath)
	templateContent, err := os.ReadFile(templateFile)
	if err != nil {
		return "", fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Parse and execute template
	tmpl, err := template.New(templatePath).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}

	return buf.String(), nil
}

// ConvertColumnToField converts a database column to a field configuration
func ConvertColumnToField(col CodeGenColumnInfo) FieldConfig {
	field := FieldConfig{
		ColumnName:   col.Name,
		FieldName:    toCamelCase(col.Name),
		JSONTag:      col.Name,
		Comment:      col.Comment,
		Nullable:     col.Nullable,
		IsPrimaryKey: col.Key == "PRI",
	}

	// Map database type to Go type
	field.FieldType = mapDBTypeToGoType(col.Type)
	field.TSType = mapDBTypeToTSType(col.Type)
	field.FormType = mapDBTypeToFormType(col.Type)
	field.Label = toLabel(col.Name)

	// Build Gorm tag
	gormTags := []string{fmt.Sprintf("column:%s", col.Name)}
	if col.Key == "PRI" {
		gormTags = append(gormTags, "primaryKey")
	}
	if !col.Nullable {
		gormTags = append(gormTags, "not null")
	}
	field.GormTag = strings.Join(gormTags, ";")

	// Determine if searchable (string types are searchable)
	field.Searchable = strings.Contains(col.Type, "varchar") || strings.Contains(col.Type, "text")

	return field
}

// Helper functions
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func toLabel(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, " ")
}

func mapDBTypeToGoType(dbType string) string {
	dbType = strings.ToLower(dbType)

	// Check for boolean first (before int check)
	if strings.Contains(dbType, "bool") || strings.Contains(dbType, "tinyint(1)") {
		return "bool"
	}
	if strings.Contains(dbType, "int") {
		if strings.Contains(dbType, "unsigned") {
			return "uint"
		}
		return "int"
	}
	if strings.Contains(dbType, "varchar") || strings.Contains(dbType, "text") || strings.Contains(dbType, "char") {
		return "string"
	}
	if strings.Contains(dbType, "decimal") || strings.Contains(dbType, "float") || strings.Contains(dbType, "double") {
		return "float64"
	}
	if strings.Contains(dbType, "datetime") || strings.Contains(dbType, "timestamp") {
		return "time.Time"
	}
	if strings.Contains(dbType, "json") {
		return "string"
	}

	return "string"
}

func mapDBTypeToTSType(dbType string) string {
	dbType = strings.ToLower(dbType)

	// Check for boolean first (before int check)
	if strings.Contains(dbType, "bool") || strings.Contains(dbType, "tinyint(1)") {
		return "boolean"
	}
	if strings.Contains(dbType, "int") || strings.Contains(dbType, "decimal") || strings.Contains(dbType, "float") || strings.Contains(dbType, "double") {
		return "number"
	}

	return "string"
}

func mapDBTypeToFormType(dbType string) string {
	dbType = strings.ToLower(dbType)

	// Check for boolean first (before int check)
	if strings.Contains(dbType, "bool") || strings.Contains(dbType, "tinyint(1)") {
		return "switch"
	}
	if strings.Contains(dbType, "int") || strings.Contains(dbType, "decimal") || strings.Contains(dbType, "float") || strings.Contains(dbType, "double") {
		return "number"
	}
	if strings.Contains(dbType, "text") {
		return "textarea"
	}

	return "input"
}
