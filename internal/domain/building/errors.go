package building

import "errors"

var (
	ErrAccountRequired    = errors.New("account is required")
	ErrNameRequired       = errors.New("building name is required")
	ErrAddressRequired    = errors.New("building address is required")
	ErrCityRequired       = errors.New("building city is required")
	ErrProvinceRequired   = errors.New("building province is required")
	ErrInvalidMeterItemID = errors.New("invalid meter item id")
	ErrMeterItemNotFound  = errors.New("meter item not found")
	ErrNotFound           = errors.New("building not found")
	ErrBuildingHasRooms   = errors.New("building has rooms")
)
