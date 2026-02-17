package system

import (
	"k-admin-system/model/common"
)

// SysRole 系统角色模型
type SysRole struct {
	common.BaseModel
	RoleName  string    `gorm:"type:varchar(50);not null" json:"roleName"`
	RoleKey   string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"roleKey"`
	DataScope string    `gorm:"type:varchar(20);default:'all'" json:"dataScope"`
	Sort      int       `gorm:"default:0" json:"sort"`
	Status    bool      `gorm:"default:true" json:"status"`
	Remark    string    `gorm:"type:varchar(255)" json:"remark"`
	Users     []SysUser `gorm:"foreignKey:RoleID" json:"-"`
	Menus     []SysMenu `gorm:"many2many:sys_role_menus;" json:"-"`
}

// TableName 指定表名
func (SysRole) TableName() string {
	return "sys_roles"
}
