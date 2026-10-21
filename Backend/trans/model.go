package trans

type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
