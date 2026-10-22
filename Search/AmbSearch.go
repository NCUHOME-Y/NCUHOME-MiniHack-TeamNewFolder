package Search

import (
	"Backend/storage"
	"context"
	"fmt"

	"gorm.io/gorm"
)

// FuzzyScope 生成“标题或正文包含关键词”的模糊查询条件（Scope 函数）
func FuzzyScope(keyword string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {

		like := "%" + keyword + "%"
		return db.Where("title LIKE ? OR description LIKE ?", like, like)
	}
}

// AmbSearch 模糊搜索
// 设计为返回标题或正文包含搜索关键词的全部帖子，按发布时间倒序、分页。
// page 从 1 起，pageSize是每页条数；返回（当前页帖子, 符合条件的条数, err）
func AmbSearch(ctx context.Context, db *gorm.DB, keyword string, page, pageSize int) ([]storage.Post, int64, error) {
	var (
		posts []storage.Post
		total int64
	)

	//统计符合条件的总条数（用户可见，希望能够加到前端的页面显示里）
	if err := db.WithContext(ctx).Model(&storage.Post{}).
		Scopes(FuzzyScope(keyword)).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计失败了唔T_T：%w", err)
	}

	// 第二步：取当前页数据，按帖子的发布顺序倒序（最新发布的帖子排在最前）
	if err := db.WithContext(ctx).Model(&storage.Post{}).
		Scopes(FuzzyScope(keyword)).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, fmt.Errorf("搜索失败了唔T_T：%w", err)
	}

	return posts, total, nil
}
