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
