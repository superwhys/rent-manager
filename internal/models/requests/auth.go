package requests

type WechatLoginReq struct {
	Code string `json:"code"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}
