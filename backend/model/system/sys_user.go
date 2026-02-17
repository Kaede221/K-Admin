package system

import (
	"k-admin-system/model/common"
)

// SysUser 系统用户模型
type SysUser struct {
	common.BaseModel
	Username  string   `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Password  string   `gorm:"type:varchar(255);not null" json:"-"`
	Nickname  string   `gorm:"type:varchar(50)" json:"nickname"`
	HeaderImg string   `gorm:"type:varchar(255)" json:"headerImg"`
	Phone     string   `gorm:"type:varchar(20)" json:"phone"`
	Email     string   `gorm:"type:varchar(100)" json:"email"`
	RoleID    uint     `gorm:"not null" json:"roleId"`
	Role      *SysRole `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Active    bool     `gorm:"default:true" json:"active"`
}

// TableName 指定表名
func (SysUser) TableName() string {
	return "sys_users"
}
