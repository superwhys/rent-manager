package unit

import "errors"

var (
	ErrAccountRequired    = errors.New("account is required")
	ErrNameRequired       = errors.New("unit name is required")
	ErrAddressRequired    = errors.New("unit address is required")
	ErrCityRequired       = errors.New("unit city is required")
	ErrProvinceRequired   = errors.New("unit province is required")
	ErrInvalidMeterItemID = errors.New("invalid meter item id")
	ErrMeterItemNotFound  = errors.New("meter item not found")
	ErrNotFound           = errors.New("unit not found")
	ErrUnitHasRooms       = errors.New("unit has rooms")
)
