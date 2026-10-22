package trans

import "time"

type RegisterReq struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	// max=72：bcrypt 只处理前 72 字节，超长会在 storage 层报错，这里提前拦成 400
	Password string `json:"password" binding:"required,min=8,max=72"`
}
type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type Token struct {
	Token string `json:"token"`
}
type PostListReq struct {
	Page     int    `json:"page" form:"page,default=1" binding:"min=1"`
	PageSize int    `json:"page_size" form:"page_size,default=10" binding:"min=1,max=50"`
	Type     *int8  `json:"type" form:"type"` // 指针，因为 0 是无效值，没传就是 nil
	Keyword  string `json:"keyword" form:"keyword"`
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
	// 长度与 avatar_url varchar(255) 对齐
	AvatarURL string `json:"avatar_url" binding:"omitempty,max=255"`
}
type CreatePost struct {
	Token string `json:"token"`
	// 各字段的长度与 posts 表列宽对齐，required 与 NOT NULL 对齐，
	// 缺字段时由 HTTP 层返回 400，而不是让 MySQL 报约束错误后再返回 500
	Type        int8   `json:"type" binding:"required,oneof=1 2"`
	Title       string `json:"title" binding:"required,max=100"`
	Description string `json:"description" binding:"required"`
	ImageURL    string `json:"image_url" binding:"omitempty,max=500"`
	Tags        string `json:"tags" binding:"omitempty,max=255"`
	Location    string `json:"location" binding:"required,max=255"`
	ContactInfo string `json:"contact_info" binding:"required,max=100"`
}
