package storage

import (
	"Backend/auth"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
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

func (s *Store) GetUserByToken(ctx context.Context, token string) (*User, error) {

	tokenhash := auth.TokenHash(token)

	sqldb, err := s.db.DB()
	userID, err := auth.VerifyToken(sqldb, tokenhash)
	if err != nil {
		return nil, err
	}

	u, err := s.GetUserByID(ctx, userID)
	return u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id uint64) (*User, error) {
	var u User
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, err
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := s.db.WithContext(ctx).Where("username = ?", username).First(&u).Error

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) UserUpdateBio(id uint64, bio sql.NullString) error {

	err := s.db.Model(&User{}).Where("id = ?", id).Update("bio", bio)

	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return errors.New("用户不存在")
	}

	s.db.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"bio": bio,
	})
	return nil
}

func ConvertToGormDB(sqlDB *sql.DB) (*gorm.DB, error) {
	// 检查传入的 sqlDB 是否为空
	if sqlDB == nil {
		return nil, fmt.Errorf("传入的 sql.DB 为空")
	}

	// 使用 mysql.New 并传入已存在的 Conn (即 *sql.DB)
	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB, // 核心：直接复用现有的连接池
	}), &gorm.Config{
		// 这里可以配置 GORM 的全局选项，例如：
		// 禁用外键约束 (如果你在数据库层面自己管理的话)
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("封装 GORM 失败: %w", err)
	}

	return gormDB, nil
}
