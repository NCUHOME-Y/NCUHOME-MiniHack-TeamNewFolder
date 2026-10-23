package storage

import (
	"database/sql"
	"time"
)

type User struct {
	ID           uint64         `gorm:"primaryKey" json:"-"`
	AccountID    string         `gorm:"type:varchar(36);unique" json:"account_id"`
	Username     string         `gorm:"type:varchar(32);unique"`
	PasswordHash string         `gorm:"type:varchar(255)" json:"-"`
	Bio          sql.NullString `gorm:"type:text"`
	AvatarURL    string         `gorm:"column:avatar_url;size:255;default:null" json:"avatar_url"`
	CreatAt      time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
type Post struct {
	ID          uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID      uint64 `gorm:"column:user_id;index;not null" json:"user_id"`
	Type        int8   `gorm:"column:type;index;not null" json:"type"` // 1-寻物, 2-招领
	Title       string `gorm:"column:title;size:100;not null" json:"title"`
	Description string `gorm:"column:description;type:text;not null" json:"description"`
	ImageURL    string `gorm:"column:image_url;size:500" json:"image_url"`

	Tags        string `gorm:"column:tags;size:255" json:"tags"`
	Location    string `gorm:"column:location;size:255;not null" json:"location"`
	ContactInfo string `gorm:"column:contact_info;size:100;not null" json:"contact_info"`
	Status      int8   `gorm:"column:status;default:0;not null" json:"status"` // 0-进行中, 1-已结帖
	// 评论数。冗余存储：列表页要显示「12 条评论」，如果每次都 COUNT 会很慢，
	// 所以发帖/评论时维护这个数字，读的时候直接取。
	CommentCount uint32 `gorm:"column:comment_count;default:0;not null" json:"comment_count"`

	// GORM 会自动处理这两个时间字段的创建和更新
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// Comment 对应 comments 表。
type Comment struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PostID    uint64    `gorm:"column:post_id;index;not null" json:"post_id"`
	UserID    uint64    `gorm:"column:user_id;not null" json:"user_id"`
	Floor     uint32    `gorm:"column:floor;not null" json:"floor"`
	ReplyTo   *uint32   `gorm:"column:reply_to" json:"reply_to"` // 指针，因为 NULL 要能表示出来
	Content   string    `gorm:"column:content;type:text;not null" json:"content"`
	ImageURL  string    `gorm:"column:image_url;size:500" json:"image_url"`
	Status    int8      `gorm:"column:status;default:0;not null" json:"status"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 显式指定表名。不加的话 GORM 会把 Comment 变成 comments
func (Comment) TableName() string { return "comments" }

// PostImage 对应 post_images 表
// target_type + target_id 这种设计叫「多态关联」：一张表同时服务帖子和评论
// 省得建 post_images / comment_images 两张结构一样的表
type PostImage struct {
	ID         uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TargetType int8      `gorm:"column:target_type;not null" json:"target_type"`
	TargetID   uint64    `gorm:"column:target_id;index;not null" json:"target_id"`
	URL        string    `gorm:"column:url;size:500;not null" json:"url"`
	SortOrder  int       `gorm:"column:sort_order;default:0;not null" json:"sort_order"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (PostImage) TableName() string { return "post_images" }

// 这两个常量给 target_type 用，避免代码里到处写裸数字 1 / 2
const (
	TargetTypePost    int8 = 1 //图片属于帖子
	TargetTypeComment int8 = 2 //图片属于评论
)
