package meteritem

import "errors"

var (
	ErrAccountRequired   = errors.New("account is required")
	ErrCodeRequired      = errors.New("meter item code is required")
	ErrUnitRequired      = errors.New("meter item unit is required")
	ErrInvalidCalcType   = errors.New("invalid meter item calc type")
	ErrNotFound          = errors.New("meter item not found")
	ErrCodeAlreadyExists = errors.New("meter item code already exists")
)
