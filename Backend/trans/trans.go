package trans

import "Backend/storage"

func UserToUserResp(user *storage.User) *UserResp {
	return &UserResp{
		Username: user.Username,
		ID:       user.ID,
		CreatAt:  user.CreatAt,
	}
}

func UserToLoginResp(user *storage.User) *LoginResp {
	return &LoginResp{
		AccountID: user.AccountID,
		User:      UserToUserResp(user),
	}
}
