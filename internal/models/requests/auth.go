package requests

type WechatLoginReq struct {
	Code string `json:"code" validate:"required"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
