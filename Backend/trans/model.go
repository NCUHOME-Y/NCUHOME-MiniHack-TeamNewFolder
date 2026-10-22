package trans

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=8"`
}
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type Token struct {
	Token string `json:"token"`
}

type UpdateBioReq struct {
	Bio string `json:"bio"`
}
type UpdateAvatarReq struct {
	AvatarURL string `json:"avatar_url"`
}
