package roomtype

import "github.com/superwhys/rent-manager/internal/xerr"

var (
	ErrAccountRequired   = xerr.New(203001, "账号不能为空")
	ErrNameRequired      = xerr.New(203002, "房型名称不能为空")
	ErrNameTooLong       = xerr.New(203003, "房型名称过长")
	ErrNotFound          = xerr.New(203004, "房型不存在")
	ErrNameAlreadyExists = xerr.New(203005, "房型名称已存在")
	ErrRoomTypeHasRooms  = xerr.New(203006, "房型下仍有房间")
)
