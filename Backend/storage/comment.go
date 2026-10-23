package storage

import (
	"Backend/trans"
	"context"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrPostNotFound    = errors.New("帖子不存在")
	ErrCommentNotFound = errors.New("评论不存在")
)

// CreateComment 发表评论，返回带楼层号的评论对象
func (s *Store) CreateComment(ctx context.Context, postID, userID uint64, content, imageURL string, replyTo *uint32) (*Comment, error) {
	c := &Comment{}

	// Transaction 里的代码要么全成功，要么全回滚
	err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1) 先确认帖子真的存在，不存在就没必要往下走
		var cnt int64
		if err := tx.Model(&Post{}).Where("id = ?", postID).Count(&cnt).Error; err != nil {
			return err
		}
		if cnt == 0 {
			return ErrPostNotFound
		}

		// 2) 原子自增评论数。这一句同时给 posts 这一行加了锁
		res := tx.Exec("UPDATE posts SET comment_count = comment_count + 1 WHERE id = ?", postID)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrPostNotFound
		}

		// 3) 把自增后的数字读回来 —— 它就是这条评论的楼层号
		var floor uint32
		if err := tx.Raw("SELECT comment_count FROM posts WHERE id = ?", postID).Scan(&floor).Error; err != nil {
			return err
		}

		// 4) 插入评论。GORM 会自动把自增主键回填到 c.ID
		*c = Comment{
			PostID:   postID,
			UserID:   userID,
			Floor:    floor,
			ReplyTo:  replyTo,
			Content:  content,
			ImageURL: imageURL,
			Status:   0, //0 = 正常
		}
		return tx.Create(c).Error
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetCommentList 取某一页评论，并把作者的用户名/头像一起拼好返回
func (s *Store) GetCommentList(ctx context.Context, postID uint64, offset, limit int) ([]trans.CommentListItem, int64, error) {
	// Count 拿总数，前端算总页数要用。status = 0 表示只统计未删除的
	var total int64
	if err := s.DB.WithContext(ctx).Model(&Comment{}).
		Where("post_id = ? AND status = 0", postID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 按楼层升序 = 从 1 楼往下排
	// 这里重新写了一遍 Model/Where 而不是复用上面那个查询对象
	var cs []Comment
	if err := s.DB.WithContext(ctx).Model(&Comment{}).
		Where("post_id = ? AND status = 0", postID).
		Order("floor ASC").Offset(offset).Limit(limit).Find(&cs).Error; err != nil {
		return nil, 0, err
	}

	// 收集所有发帖人 ID
	userIDs := make([]uint64, 0, len(cs))
	for _, c := range cs {
		userIDs = append(userIDs, c.UserID)
	}

	// 批量查用户。Select 只取需要的三列，不要把 password_hash 也读进内存
	userMap := make(map[uint64]User, len(userIDs))
	if len(userIDs) > 0 {
		var users []User
		if err := s.DB.WithContext(ctx).Model(&User{}).
			Select("id, username, avatar_url").
			Where("id IN ?", userIDs).Find(&users).Error; err != nil {
			return nil, 0, err
		}
		for _, u := range users {
			userMap[u.ID] = u
		}
	}
	// 组装返回值。用 make(...,0,len) 而不是 var list []T
	// 这样即使一条评论都没有，返回的也是 [] 而不是 null，前端不用写 ?? []
	list := make([]trans.CommentListItem, 0, len(cs))
	for _, c := range cs{
		list = append(list, trans.CommentListItem{
			ID: c.ID,
			Floor: c.Floor,
			ReplyTo: c.ReplyTo,
			Content: c.Content,
			ImageURL: c.ImageURL,
			CreatedAt: c.CreatedAt,
			UserID: c.UserID,
			// 用户可能已被删除，这时 userMap 取到零值，Username 是空串，不会 panic
			Username: userMap[c.UserID].Username,
			AvatarURL: userMap[c.UserID].AvatarURL,
		})
	}
	return list, total, nil
}
// DeleteComment 软删除评论：只把 status 改成 1，不真的删
// 硬删会让楼层号整体前移：原本 4 楼会变成 3 楼，正在看帖的人会发现内容串位
// 软删后前端显示成「该楼层已被删除」即可
func (s *Store) DeleteComment(ctx context.Context, commentID, userID uint64) error {
	res := s.DB.WithContext(ctx).Model(&Comment{}).
		Where("id = ? AND user_id = ? AND status = 0", commentID, userID).
		Update("status", 1)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrCommentNotFound
	}
	return nil 
}

// AddImages 给帖子或评论批量挂图片。urls 的顺序就是展示顺序
func (s *Store) AddImages(ctx context.Context, targetType int8, targetID uint64, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	imgs := make([]PostImage, 0, len(urls))
	for i, u := range urls {
		imgs = append(imgs, PostImage{
			TargetType: targetType,
			TargetID: targetID,
			URL: u,
			SortOrder: i,
		})
	}
	// 传切片给 Create 会拼成一条 INSERT 插入多行，比循环里一条条插快得多
	return s.DB.WithContext(ctx).Create(&imgs).Error
}

// GetImages 批量取图片，返回 map[目标ID][]图片地址
func (s *Store) GetImages(ctx context.Context, targetType int8, targetIDs []uint64) (map[uint64][]string, error) {
	out := make(map[uint64][]string)
	if len(targetIDs) == 0 {
		return out, nil
	}
	var imgs []PostImage
	if err := s.DB.WithContext(ctx).
		Where("target_type = ? AND target_id IN ?", targetType, targetIDs).
		Order("sort_order ASC").Find(&imgs).Error; err != nil {
			return nil, err
	}
	for _, im := range imgs {
		out[im.TargetID] = append(out[im.TargetID], im.URL)
	}
	return out, nil
}