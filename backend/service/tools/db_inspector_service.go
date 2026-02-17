package tools

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"k-admin-system/global"
)

// DBInspectorService 数据库检查器服务
type DBInspectorService struct{}

// ColumnInfo 列信息
type ColumnInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Key      string `json:"key"`
	Default  string `json:"default"`
	Extra    string `json:"extra"`
	Comment  string `json:"comment"`
}

// GetTables 获取所有表名
func (s *DBInspectorService) GetTables() ([]string, error) {
	var tables []string

	// 检测数据库类型
	dbType := global.DB.Dialector.Name()

	if dbType == "sqlite" {
		// SQLite: 从 sqlite_master 查询表
		query := `SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
		if err := global.DB.Raw(query).Scan(&tables).Error; err != nil {
			return nil, fmt.Errorf("failed to get tables: %w", err)
		}
	} else {
		// MySQL: 使用 information_schema
		var dbName string
		if err := global.DB.Raw("SELECT DATABASE()").Scan(&dbName).Error; err != nil {
			return nil, fmt.Errorf("failed to get database name: %w", err)
		}

		query := `SELECT table_name FROM information_schema.tables 
		          WHERE table_schema = ? AND table_type = 'BASE TABLE'
		          ORDER BY table_name`

		if err := global.DB.Raw(query, dbName).Scan(&tables).Error; err != nil {
			return nil, fmt.Errorf("failed to get tables: %w", err)
		}
	}

	return tables, nil
}

// GetTableSchema 获取表结构
func (s *DBInspectorService) GetTableSchema(tableName string) ([]CodeGenColumnInfo, error) {
	// 验证表名（防止SQL注入）
	if !isValidTableName(tableName) {
		return nil, errors.New("invalid table name")
	}

	var columns []CodeGenColumnInfo

	// 检测数据库类型
	dbType := global.DB.Dialector.Name()

	if dbType == "sqlite" {
		// SQLite: 使用 PRAGMA table_info
		type sqliteColumn struct {
			CID       int    `gorm:"column:cid"`
			Name      string `gorm:"column:name"`
			Type      string `gorm:"column:type"`
			NotNull   int    `gorm:"column:notnull"`
			DfltValue string `gorm:"column:dflt_value"`
			PK        int    `gorm:"column:pk"`
		}

		var sqliteColumns []sqliteColumn
		query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
		if err := global.DB.Raw(query).Scan(&sqliteColumns).Error; err != nil {
			return nil, fmt.Errorf("failed to get table schema: %w", err)
		}

		if len(sqliteColumns) == 0 {
			return nil, errors.New("table not found")
		}

		// 转换为 ColumnInfo
		for _, col := range sqliteColumns {
			key := ""
			if col.PK > 0 {
				key = "PRI"
			}
			columns = append(columns, CodeGenColumnInfo{
				Name:     col.Name,
				Type:     col.Type,
				Nullable: col.NotNull == 0,
				Key:      key,
				Default:  col.DfltValue,
				Extra:    "",
				Comment:  "",
			})
		}
	} else {
		// MySQL: 使用 information_schema
		var dbName string
		if err := global.DB.Raw("SELECT DATABASE()").Scan(&dbName).Error; err != nil {
			return nil, fmt.Errorf("failed to get database name: %w", err)
		}

		query := `SELECT 
		            column_name as name,
		            column_type as type,
		            is_nullable = 'YES' as nullable,
		            column_key as ` + "`key`" + `,
		            COALESCE(column_default, '') as ` + "`default`" + `,
		            extra,
		            COALESCE(column_comment, '') as comment
		          FROM information_schema.columns
		          WHERE table_schema = ? AND table_name = ?
		          ORDER BY ordinal_position`

		if err := global.DB.Raw(query, dbName, tableName).Scan(&columns).Error; err != nil {
			return nil, fmt.Errorf("failed to get table schema: %w", err)
		}

		if len(columns) == 0 {
			return nil, errors.New("table not found")
		}
	}

	return columns, nil
}

// GetTableData 获取表数据（支持分页）
func (s *DBInspectorService) GetTableData(tableName string, page, pageSize int) ([]map[string]interface{}, int64, error) {
	// 验证表名
	if !isValidTableName(tableName) {
		return nil, 0, errors.New("invalid table name")
	}

	var total int64
	var data []map[string]interface{}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	if err := global.DB.Raw(countQuery).Scan(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count records: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf("SELECT * FROM `%s` LIMIT ? OFFSET ?", tableName)
	if err := global.DB.Raw(dataQuery, pageSize, offset).Scan(&data).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to query table data: %w", err)
	}

	return data, total, nil
}

// ExecuteSQL 执行SQL语句
func (s *DBInspectorService) ExecuteSQL(sql string, readOnly bool) (interface{}, error) {
	// 验证SQL
	if err := s.ValidateSQL(sql, readOnly); err != nil {
		return nil, err
	}

	// 判断是查询还是执行
	sqlUpper := strings.ToUpper(strings.TrimSpace(sql))
	if strings.HasPrefix(sqlUpper, "SELECT") ||
		strings.HasPrefix(sqlUpper, "SHOW") ||
		strings.HasPrefix(sqlUpper, "DESCRIBE") ||
		strings.HasPrefix(sqlUpper, "DESC") {
		// 查询操作
		var results []map[string]interface{}
		if err := global.DB.Raw(sql).Scan(&results).Error; err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		return results, nil
	} else {
		// 执行操作
		result := global.DB.Exec(sql)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to execute SQL: %w", result.Error)
		}
		return map[string]interface{}{
			"rows_affected": result.RowsAffected,
		}, nil
	}
}

// CreateRecord 创建记录
func (s *DBInspectorService) CreateRecord(tableName string, data map[string]interface{}) error {
	// 验证表名
	if !isValidTableName(tableName) {
		return errors.New("invalid table name")
	}

	if len(data) == 0 {
		return errors.New("no data provided")
	}

	// 构建INSERT语句
	var columns []string
	var placeholders []string
	var values []interface{}

	for col, val := range data {
		columns = append(columns, fmt.Sprintf("`%s`", col))
		placeholders = append(placeholders, "?")
		values = append(values, val)
	}

	query := fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	if err := global.DB.Exec(query, values...).Error; err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	return nil
}

// UpdateRecord 更新记录
func (s *DBInspectorService) UpdateRecord(tableName string, id interface{}, data map[string]interface{}) error {
	// 验证表名
	if !isValidTableName(tableName) {
		return errors.New("invalid table name")
	}

	if len(data) == 0 {
		return errors.New("no data provided")
	}

	// 构建UPDATE语句
	var setClauses []string
	var values []interface{}

	for col, val := range data {
		setClauses = append(setClauses, fmt.Sprintf("`%s` = ?", col))
		values = append(values, val)
	}
	values = append(values, id)

	query := fmt.Sprintf("UPDATE `%s` SET %s WHERE id = ?",
		tableName,
		strings.Join(setClauses, ", "))

	result := global.DB.Exec(query, values...)
	if result.Error != nil {
		return fmt.Errorf("failed to update record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

// DeleteRecord 删除记录
func (s *DBInspectorService) DeleteRecord(tableName string, id interface{}) error {
	// 验证表名
	if !isValidTableName(tableName) {
		return errors.New("invalid table name")
	}

	query := fmt.Sprintf("DELETE FROM `%s` WHERE id = ?", tableName)

	result := global.DB.Exec(query, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return errors.New("record not found")
	}

	return nil
}

// ValidateSQL 验证SQL语句
func (s *DBInspectorService) ValidateSQL(sql string, readOnly bool) error {
	if strings.TrimSpace(sql) == "" {
		return errors.New("SQL statement is empty")
	}

	sqlUpper := strings.ToUpper(strings.TrimSpace(sql))

	// 只读模式下的限制
	if readOnly {
		// 只允许SELECT、SHOW、DESCRIBE、DESC
		if !strings.HasPrefix(sqlUpper, "SELECT") &&
			!strings.HasPrefix(sqlUpper, "SHOW") &&
			!strings.HasPrefix(sqlUpper, "DESCRIBE") &&
			!strings.HasPrefix(sqlUpper, "DESC") {
			return errors.New("only SELECT, SHOW, DESCRIBE, DESC statements are allowed in read-only mode")
		}
	}

	// 危险操作检查（需要超级管理员权限）
	dangerousKeywords := []string{
		"DROP",
		"TRUNCATE",
		"ALTER DATABASE",
		"CREATE DATABASE",
		"DROP DATABASE",
	}

	for _, keyword := range dangerousKeywords {
		if strings.Contains(sqlUpper, keyword) {
			return fmt.Errorf("dangerous operation '%s' is not allowed", keyword)
		}
	}

	return nil
}

// isValidTableName 验证表名是否合法
func isValidTableName(tableName string) bool {
	// 只允许字母、数字、下划线
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, tableName)
	return matched
}
