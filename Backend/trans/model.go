package trans

import (
	"time"
)

type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type LoginResp struct {
	AccountID string    `json:"accountid"` // 登录凭证（如 JWT）
	User      *UserResp `json:"user"`      // 用户信息（绝对不包含密码）
}
type Token struct {
	AccountID string `json:"accountid"`
}

type UserResp struct {
	Username string    `json:"username" binding:"required"`
	ID       uint64    `json:"id" binding:"required"`
	CreatAt  time.Time `json:"createat" binding:"required"`
}
