package auth

type SendEmailCodeRequest struct {
	Email string `json:"email" form:"-" header:"-" validate:"required"`
}

type EmailLoginRequest struct {
	Email string `json:"email" form:"-" header:"-" validate:"required"`
	Code  string `json:"code"  form:"-" header:"-" validate:"required,len=6,number"`
}

type PasswordLoginRequest struct {
	Email    string `json:"email"    form:"-" header:"-" validate:"required"`
	Password string `json:"password" form:"-" header:"-" validate:"required,max=128"`
}

type RegisterRequest struct {
	Email    string `json:"email"    form:"-" header:"-" validate:"required"`
	Code     string `json:"code"     form:"-" header:"-" validate:"required,len=6,number"`
	Password string `json:"password" form:"-" header:"-" validate:"required,min=8,max=128"`
}

type SendEmailCodeResponse struct {
	ResendAfterSeconds int `json:"resend_after_seconds"`
	ExpiresInSeconds   int `json:"expires_in_seconds"`
}
