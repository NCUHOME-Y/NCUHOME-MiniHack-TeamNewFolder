package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, username, password, bio string) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("加密密码失败：%w", err)
	}

	u := &User{
		AccountID:    uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
		Bio:          sql.NullString{String: bio, Valid: bio != ""},
	}

	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) Login(ctx context.Context, username, password, bio string) (*User, error) {
	u, uerr := s.GetUserByUsername(ctx, username)
	if uerr == nil {
		return nil, fmt.Errorf("用户不存在：%w", uerr)
	}

	hasherr := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if hasherr != nil {
		return nil, fmt.Errorf("密码错误：%w", hasherr)
	}

	return u, nil
}

func (s *Store) GetUserByAccountID(ctx context.Context, accountID string) (*User, error) {
	var u User
	err := s.db.WithContext(ctx).Where("account_id = ?", accountID).First(&u).Error

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&u).Error

	if err != nil {
		return nil, err
	}
	return &u, nil
}
