package roomtype

import "errors"

var (
	ErrAccountRequired   = errors.New("账号不能为空")
	ErrNameRequired      = errors.New("房型名称不能为空")
	ErrNameTooLong       = errors.New("房型名称过长")
	ErrNotFound          = errors.New("房型不存在")
	ErrNameAlreadyExists = errors.New("房型名称已存在")
	ErrRoomTypeHasRooms  = errors.New("房型下仍有房间")
)
