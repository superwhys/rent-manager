package model

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`

	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}

type LogoutRequest struct {
	Username string `json:"username"`
}

type ChangePasswordRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type GetUserInfoRequest struct {
	Username string `json:"username"`
}

type UpdateUserInfoRequest struct {
	Username string `json:"username"`
	Info     User   `json:"info"`
}

type DeleteUserRequest struct {
	Username string `json:"username"`
}

type CreateTenantRequest struct {
	Username   string `json:"username"`
	Phone      string `json:"phone"`
	IDCard     string `json:"id_card"`
	RoomID     uint   `json:"room_id"`
	PropertyID uint   `json:"property_id"`
	MoveInDate string `json:"move_in_date"`
}

type UpdateTenantRequest struct {
	ID         uint   `json:"id"`
	Username   string `json:"username"`
	Phone      string `json:"phone"`
	IDCard     string `json:"id_card"`
	RoomID     uint   `json:"room_id"`
	PropertyID uint   `json:"property_id"`
	MoveInDate string `json:"move_in_date"`
}

type DeleteTenantRequest struct {
	ID uint `json:"id"`
}
