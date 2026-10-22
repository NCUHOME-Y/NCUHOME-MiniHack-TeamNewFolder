package storage

import (
	"Backend/auth"
	"Backend/trans"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var ErrUserExists = errors.New("用户名已存在")
var DefaultImage = "https://i.imgs.ovh/2026/10/22/9aaf8be43feca969d89e45781b6abbc4.png"

type Store struct {
	DB *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{DB: db}
}

func (s *Store) CreateUser(ctx context.Context, username, password, bio string) (*User, error) {

	if len(password) > 72 {
		return nil, fmt.Errorf("密码过长")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, fmt.Errorf("加密密码失败：%w", err)
	}

	u := &User{
		AccountID:    uuid.NewString(),
		Username:     username,
		PasswordHash: string(hash),
		Bio:          sql.NullString{String: bio, Valid: bio != ""},
		AvatarURL:    DefaultImage,
	}

	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return nil, ErrUserExists
	}
	if err := s.DB.WithContext(ctx).Create(u).Error; err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) Login(ctx context.Context, username, password string) (*User, error) {
	u, uerr := s.GetUserByUsername(ctx, username)
	if uerr != nil {
		return nil, fmt.Errorf("用户不存在：%w", uerr)
	}

	hasherr := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	if hasherr != nil {
		return nil, fmt.Errorf("密码错误：%w", hasherr)
	}

	return u, nil
}

func (s *Store) GetUserByToken(ctx context.Context, rawToken string) (*User, error) {

	sqldb, err := s.DB.DB()
	if err != nil {
		return nil, err
	}

	userID, err := auth.VerifyToken(sqldb, rawToken)
	if err != nil {
		return nil, err
	}

	u, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id uint64) (*User, error) {
	var u User
	err := s.DB.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, err
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := s.DB.WithContext(ctx).Where("username = ?", username).First(&u).Error

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) UserUpdateBio(id uint64, bio string) error {
	bioNS := sql.NullString{String: bio, Valid: true}

	err := s.DB.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"bio": bioNS,
	})
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return errors.New("无变化")
	}

	return nil
}
func (s *Store) UserUpdateAvatar(id uint64, avatarURL string) error {
	err := s.DB.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"avatar_url": avatarURL,
	})
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return errors.New("无变化")
	}
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

func (s *Store) DeleteToken(rawToken string) error {
	tokenHash := auth.TokenHash(rawToken)

	result := s.DB.Where("token_hash = ?", tokenHash).Update("revoked", 1)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (s *Store) UserCreatePost(p *Post) error {
	err := s.DB.Create(&p).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetPostList(postType int8, status int8, offset int, limit int) ([]trans.PostListItem, int64, error) {
	query := s.DB.Table("posts").Model(&Post{})

	//根据过滤信息调整返回帖子
	if postType > 0 {
		query = query.Where("type = ?", postType)
	}
	if status >= 0 {
		query = query.Where("status = ?", status)
	}

	var tol int64 //符合要求帖子总数
	if err := query.Count(&tol).Error; err != nil {
		return nil, 0, err
	}

	var posts []Post
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&posts).Error
	if err != nil {
		return nil, 0, err
	}

	var userIDs []uint64
	for _, p := range posts {
		userIDs = append(userIDs, p.UserID)
	}

	var users []User
	if len(userIDs) > 0 {
		s.DB.Where("id IN ?", userIDs).Find(&users)
	}
	userMap := make(map[uint64]User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	var list []trans.PostListItem
	for _, p := range posts {
		item := trans.PostListItem{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			ImageURL:    p.ImageURL,
			Type:        p.Type,
			Tags:        p.Tags,
			Location:    p.Location,
			Status:      p.Status,
			CreatedAt:   p.CreatedAt,

			// 发帖人简易信息
			UserID:    p.UserID,
			Username:  userMap[p.UserID].Username,
			AvatarURL: userMap[p.UserID].AvatarURL,
		}
		list = append(list, item)
	}
	return list, tol, nil
}
