package system

import (
	"k-admin-system/model/common"
)

// SysMenu 系统菜单模型
type SysMenu struct {
	common.BaseModel
	ParentID  uint      `gorm:"default:0" json:"parentId"`
	Path      string    `gorm:"type:varchar(100)" json:"path"`
	Name      string    `gorm:"type:varchar(50)" json:"name"`
	Component string    `gorm:"type:varchar(100)" json:"component"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Meta      string    `gorm:"type:json" json:"meta"`
	BtnPerms  string    `gorm:"type:json" json:"btnPerms"`
	Children  []SysMenu `gorm:"-" json:"children,omitempty"`
	Roles     []SysRole `gorm:"many2many:sys_role_menus;" json:"-"`
}

// TableName 指定表名
func (SysMenu) TableName() string {
	return "sys_menus"
}
