package web

type SignUpReq struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type LoginTokenReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserEditReq struct {
	Nickname string `json:"nickname"`
	Birthday string `json:"birthday"`
	Bio      string `json:"bio"`
}

type SendSMSCodeReq struct {
	Phone string `json:"phone"`
}

type LoginSMSReq struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}
