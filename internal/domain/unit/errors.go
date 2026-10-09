package unit

import "github.com/superwhys/rent-manager/internal/xerr"

var (
	ErrAccountRequired        = xerr.New(201001, "账号不能为空")
	ErrNameRequired           = xerr.New(201002, "单元名称不能为空")
	ErrAddressRequired        = xerr.New(201003, "单元地址不能为空")
	ErrCityRequired           = xerr.New(201004, "单元城市不能为空")
	ErrProvinceRequired       = xerr.New(201005, "单元省份不能为空")
	ErrNameAlreadyExists      = xerr.New(201006, "单元名称已存在")
	ErrInvalidListScope       = xerr.New(201007, "单元列表范围不合法")
	ErrInvalidMeterItemID     = xerr.New(201008, "抄表项 ID 不合法")
	ErrMeterItemNotFound      = xerr.New(201009, "抄表项不存在")
	ErrNotFound               = xerr.New(201010, "单元不存在")
	ErrUnitHasRooms           = xerr.New(201011, "单元下仍有房间")
	ErrInvalidMeterReadingDay = xerr.New(201012, "抄表日必须是 1-31 的整数")
	ErrInvalidJsonExtend      = xerr.New(201013, "单元扩展配置必须是合法 JSON 对象")
)
