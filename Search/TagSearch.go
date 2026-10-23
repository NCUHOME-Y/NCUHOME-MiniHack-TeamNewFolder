package Search

import (
	"Backend/Tags"
	"Backend/storage"
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ai辅助修的，这玩意我写的时候很糙，只有第一个tags全等
// TagScope 生成"帖子带有指定标签"的 GORM 查询条件（Scope 函数）。
// 四个分支覆盖目标标签在逗号分隔文本里所有合法位置：
//
//	tags = "钱包"          → 整串只有这一个标签
//	tags LIKE "钱包,%"     → 目标标签在开头
//	tags LIKE "%,钱包"     → 目标标签在结尾
//	tags LIKE "%,钱包,%"   → 目标标签在中间
func TagScope(tagName string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// 分隔符统一从 Tags 包取（当前是逗号），保证两个模块对格式的理解一致
		sep := Tags.TagSep
		return db.Where(
			"tags = ? OR tags LIKE ? OR tags LIKE ? OR tags LIKE ?",
			tagName,
			tagName+sep+"%",
			"%"+sep+tagName,
			"%"+sep+tagName+sep+"%",
		)
	}
}

// TagSearch 标签搜索
// Tags包来校验标签确实在标签库里，再返回带这个标签的全部帖子，按发布时间倒序，同时做到分页。
// 标签不存在时原样返回Tags.TagNFound
func TagSearch(ctx context.Context, db *gorm.DB, tagName string, page, pageSize int) ([]storage.Post, int64, error) {
	//该死的错误处理
	//ai修的，我真的恨死错误处理了tnnd
	// 只允许输入一个标签：出现分隔符（逗号）即拒绝
	if strings.Contains(tagName, Tags.TagSep) {
		return nil, 0, MultTag
	}
	// 校验标签在基准标签库中存在（不在库里的标签直接拒绝，不进 SQL）
	if _, err := Tags.GetTag(ctx, db, tagName); err != nil {
		return nil, 0, err
	}

	var (
		posts []storage.Post
		total int64
	)

	//统计该标签帖子总数量（需要前端展示）
	if err := db.WithContext(ctx).Model(&storage.Post{}).
		Scopes(TagScope(tagName)).
		Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计失败了唔T_T：%w", err)
	}

	// 取带TAG当前页帖子，倒序（最新发布的帖子排在最前）
	if err := db.WithContext(ctx).Model(&storage.Post{}).
		Scopes(TagScope(tagName)).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, fmt.Errorf("标签搜索失败了唔T_T：%w", err)
	}

	return posts, total, nil
}

//哪来的那么多错误处理啊T_T
