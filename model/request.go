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
