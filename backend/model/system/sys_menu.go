package system

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"k-admin-system/model/common"
)

// MenuMeta 菜单元数据
type MenuMeta struct {
	Icon      string `json:"icon"`
	Title     string `json:"title"`
	Hidden    bool   `json:"hidden"`
	KeepAlive bool   `json:"keep_alive"`
}

// Scan 实现 sql.Scanner 接口
func (m *MenuMeta) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal MenuMeta value")
	}
	return json.Unmarshal(bytes, m)
}

// Value 实现 driver.Valuer 接口
func (m MenuMeta) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// SysMenu 系统菜单模型
type SysMenu struct {
	common.BaseModel
	ParentID  uint      `gorm:"default:0" json:"parentId"`
	Path      string    `gorm:"type:varchar(100)" json:"path"`
	Name      string    `gorm:"type:varchar(50)" json:"name"`
	Component string    `gorm:"type:varchar(100)" json:"component"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Meta      MenuMeta  `gorm:"type:json;serializer:json" json:"meta"`
	BtnPerms  []string  `gorm:"type:json;serializer:json" json:"btn_perms"`
	Children  []SysMenu `gorm:"-" json:"children,omitempty"`
	Roles     []SysRole `gorm:"many2many:sys_role_menus;" json:"-"`
}

// TableName 指定表名
func (SysMenu) TableName() string {
	return "sys_menus"
}
