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

	// GORM 会自动处理这两个时间字段的创建和更新
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}
