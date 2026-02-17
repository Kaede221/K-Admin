package system

// SysCasbinRule Casbin规则模型
// 用于存储Casbin的RBAC策略规则
// 表结构遵循Casbin Gorm Adapter的标准格式
type SysCasbinRule struct {
	ID    uint   `gorm:"primarykey;autoIncrement" json:"id"`
	Ptype string `gorm:"size:100;uniqueIndex:unique_index" json:"ptype"`
	V0    string `gorm:"size:100;uniqueIndex:unique_index" json:"v0"`
	V1    string `gorm:"size:100;uniqueIndex:unique_index" json:"v1"`
	V2    string `gorm:"size:100;uniqueIndex:unique_index" json:"v2"`
	V3    string `gorm:"size:100;uniqueIndex:unique_index" json:"v3"`
	V4    string `gorm:"size:100;uniqueIndex:unique_index" json:"v4"`
	V5    string `gorm:"size:100;uniqueIndex:unique_index" json:"v5"`
}

// TableName 指定表名
func (SysCasbinRule) TableName() string {
	return "sys_casbin_rules"
}
