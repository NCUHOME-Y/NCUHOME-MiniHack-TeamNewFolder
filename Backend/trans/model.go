package trans

import "time"

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
type PostListReq struct {
	Page     int   `json:"page" form:"page,default=1" binding:"min=1"`
	PageSize int   `json:"page_size" form:"page_size,default=10" binding:"min=1,max=50"`
	Type     *int8 `json:"type" form:"type"` // 指针，因为 0 是无效值，没传就是 nil
}
type PostListItem struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	Type        int8      `json:"type"`
	Tags        string    `json:"tags"`
	Location    string    `json:"location"`
	Status      int8      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`

	// 发帖人简易信息
	UserID    uint64 `json:"user_id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

type UpdateBioReq struct {
	Bio string `json:"bio"`
}
type UpdateAvatarReq struct {
	AvatarURL string `json:"avatar_url"`
}
type CreatePost struct {
	Token       string `json:"token"`
	Type        int8   `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Tags        string `json:"tags"`
	Location    string `json:"location"`
	ContactInfo string `json:"contact_info"`
}
