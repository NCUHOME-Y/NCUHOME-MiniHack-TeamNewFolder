package storage

import (
	"database/sql"
	"time"
)

type User struct {
	ID           uint64         `gorm:"primaryKey"`
	AccountID    string         `gorm:"type:varchar(36);unique"`
	Username     string         `gorm:"type:varchar(32);unique"`
	PasswordHash string         `gorm:"type:varchar(255)"`
	Bio          sql.NullString `gorm:"type:text"`
	CreatAt      time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
