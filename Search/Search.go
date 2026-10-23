// 搜索框架，下分两个子函数：AmbSearch（模糊搜索）与 TagSearch（标签搜索）
// 一次只能进行一种搜索方式：keyword 与 tag 二选一
package Search

import (
	"Backend/Tags"
	"Backend/storage"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 搜索分页默认值与最大上限
const (
	DefPage     = 1  // 默认页码
	DefPageSize = 10 // 默认每页帖子条数
	MaxPageSize = 10 // 最大每页条数上限，防止用户翻到吐血OvooVO
)

// 搜索方式约束的哨兵错误（handler 依此返回 400）
var (
	SerachConfilct = errors.New("模糊搜索与标签搜索一次只能用一种，请二选一")
	MultTag        = errors.New("标签搜索一次只能输入一个标签")
)

// SearchPosts作为搜索框架的大一统入口
// 若keyword非空，用模糊搜索条件（ AmbSearch 的 FuzzyScope ）
// 若tag非空，用标签条件（ TagSearch 的 TagScope ）
// 两者都非空 → 返回 SerachConfilct（一次只能进行一种搜索方式）
// tag中带分隔符（想一次传多个标签）→ 返回 MultTag（只允许输入一个标签）
// 都为空返回全部帖子（按最新排序，相当于帖子列表）
// 我恨死错误处理了
// 返回当前页帖子, 符合条件的总条数, 错误；标签不存在时返回 Tags.TagNFound
func SearchPosts(ctx context.Context, db *gorm.DB, keyword, tag string, page, pageSize int) ([]storage.Post, int64, error) {
	// 互斥校验：一次只能进行一种搜索方式，模糊与标签二选一
	if keyword != "" && tag != "" {
		return nil, 0, SerachConfilct
	}
	// 标签搜索只允许输入一个标签：tag 里出现分隔符（逗号）即拒绝
	if strings.Contains(tag, Tags.TagSep) {
		return nil, 0, MultTag
	}

	var (
		posts []storage.Post
		total int64
	)

	// 标签搜索时，先经 Tags 包校验标签是否在标签库中
	if tag != "" {
		if _, err := Tags.GetTag(ctx, db, tag); err != nil {
			return nil, 0, err
		}
	}

	buildQuery := func() *gorm.DB {
		query := db.WithContext(ctx).Model(&storage.Post{})
		if keyword != "" {
			//开始准备用AmbSearch
			query = query.Scopes(FuzzyScope(keyword))
		}
		if tag != "" {
			//开始准备用TagSearch
			query = query.Scopes(TagScope(tag))
		}
		return query
	}

	//统计符合条件的总条数（给前端做分页展示）
	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计帖子数量了唔T_T：%w", err)
	}

	//取当前页数据，倒序（最新发布的帖子排在最前）
	if err := buildQuery().
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, fmt.Errorf("查询帖子列表失败：%w", err)
	}

	return posts, total, nil
}

// ApiPost 返回统一搜索的 gin 处理器（HTTP 层）
func ApiPost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		//读一下模糊关键词与标签名
		keyword := c.Query("keyword")
		tag := c.Query("tag")

		//狗操的错误处理我服了T_T
		//ai帮我修的
		page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(DefPage)))
		if err != nil || page < 1 {
			page = DefPage
		}
		pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(DefPageSize)))
		if err != nil || pageSize < 1 {
			pageSize = DefPageSize
		}
		if pageSize > MaxPageSize {
			pageSize = MaxPageSize
		}

		//调用框架统一搜索入口
		posts, total, err := SearchPosts(c.Request.Context(), db, keyword, tag, page, pageSize)
		if err != nil {
			//状态码
			// 搜索方式冲突 / 一次传多个标签 → 400（请求本身不合法）
			if errors.Is(err, SerachConfilct) || errors.Is(err, MultTag) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "搜索失败：" + err.Error()})
				return
			}
			if errors.Is(err, Tags.TagNFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "搜索失败：" + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "搜索失败：" + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":   "搜索成功",
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"posts":     posts,
		})
	}
}

//把搜索路由注册到给定路由组

func RegRoutes(api *gin.RouterGroup, db *gorm.DB) {
	api.GET("/search/posts", ApiPost(db))
}
