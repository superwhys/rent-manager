package models

import "time"

type UnitAccountRole struct {
	ID                 int        `gorm:"column:id;type:bigint;primaryKey;autoIncrement"`
	CreatedAt          time.Time  `gorm:"column:created_at;type:timestamp;not null;autoCreateTime;<-:create"`
	UpdatedAt          time.Time  `gorm:"column:updated_at;type:timestamp;not null;autoUpdateTime"`
	DeletedAt          *time.Time `gorm:"column:deleted_at;type:timestamp"`
	Status             string     `gorm:"column:status;type:enum('active','disabled');not null;default:active"`
	AccountID          int        `gorm:"column:account_id;type:bigint;not null;uniqueIndex:unitaccountrole_unit_id_account_id_role_id,priority:2;index:unitaccountrole_account_id"`
	GrantedByAccountID int        `gorm:"column:granted_by_account_id;type:bigint;not null"`
	RoleID             int        `gorm:"column:role_id;type:bigint;not null;uniqueIndex:unitaccountrole_unit_id_account_id_role_id,priority:3;index:unitaccountrole_role_id"`
	UnitID             int        `gorm:"column:unit_id;type:bigint;not null;uniqueIndex:unitaccountrole_unit_id_account_id_role_id,priority:1;index:unitaccountrole_unit_id"`
	Account            *Account   `gorm:"foreignKey:AccountID;constraint:unit_account_roles_accounts_unit_account_roles,OnDelete:NO ACTION"`
	GrantedByAccount   *Account   `gorm:"foreignKey:GrantedByAccountID;constraint:unit_account_roles_accounts_granted_unit_account_roles,OnDelete:NO ACTION"`
	Role               *Role      `gorm:"foreignKey:RoleID;constraint:unit_account_roles_roles_unit_account_roles,OnDelete:NO ACTION"`
	Unit               *Unit      `gorm:"foreignKey:UnitID;constraint:unit_account_roles_units_unit_account_roles,OnDelete:NO ACTION"`
}

func (UnitAccountRole) TableName() string {
	return "unit_account_roles"
}
