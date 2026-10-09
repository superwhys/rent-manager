package models

import "time"

type UnitAccountRole struct {
	ID                 int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt          time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt          *time.Time `gorm:"column:deleted_at;type:timestamp"`
	Status             string     `gorm:"column:status;type:enum('active','disabled');not null;default:active"`
	AccountID          string     `gorm:"column:account_id;type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null;uniqueIndex:unitaccountrole_unit_id_account_id_role_id,priority:2;index:unitaccountrole_account_id"`
	GrantedByAccountID string     `gorm:"column:granted_by_account_id;type:varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;not null"`
	RoleID             int        `gorm:"column:role_id;type:bigint;not null;uniqueIndex:unitaccountrole_unit_id_account_id_role_id,priority:3;index:unitaccountrole_role_id"`
	UnitID             int        `gorm:"column:unit_id;type:bigint;not null;uniqueIndex:unitaccountrole_unit_id_account_id_role_id,priority:1;index:unitaccountrole_unit_id"`
	Role               *Role      `gorm:"foreignKey:RoleID;constraint:unit_account_roles_roles_unit_account_roles,OnDelete:NO ACTION"`
	Unit               *Unit      `gorm:"foreignKey:UnitID;constraint:unit_account_roles_units_unit_account_roles,OnDelete:NO ACTION"`
}

func (UnitAccountRole) TableName() string {
	return "unit_account_roles"
}
