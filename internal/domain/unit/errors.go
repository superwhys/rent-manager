package unit

import "errors"

var (
	ErrAccountRequired    = errors.New("账号不能为空")
	ErrNameRequired       = errors.New("单元名称不能为空")
	ErrAddressRequired    = errors.New("单元地址不能为空")
	ErrCityRequired       = errors.New("单元城市不能为空")
	ErrProvinceRequired   = errors.New("单元省份不能为空")
	ErrNameAlreadyExists  = errors.New("单元名称已存在")
	ErrInvalidListScope   = errors.New("单元列表范围不合法")
	ErrInvalidMeterItemID = errors.New("抄表项 ID 不合法")
	ErrMeterItemNotFound  = errors.New("抄表项不存在")
	ErrNotFound           = errors.New("单元不存在")
	ErrUnitHasRooms       = errors.New("单元下仍有房间")
)
