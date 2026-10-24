package Tags

import (
	"Backend/storage"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 标签相关常量
const (
	// DefTagName 早期用来占位的测试标签。
	// 【P0 修复后已不再写入标签库】保留这个常量只是为了不破坏外部可能的引用，
	// 真正初始化标签库用的是下面的 DefTags 列表。
	DefTagName = "TEST"
	// TagSep posts.tags 列中多个标签的分隔符（逗号分隔纯文本）
	// ⚠ 被 Tags 包内 3 处 + Search 包 2 处引用，删掉就编译不过
	TagSep = ","
	// MaxTagLength 单个标签名的最大字符数
	MaxTagLength = 32
)

// DefTags 默认标签库。
// 【P0 修复】这些名字必须和前端 Frontend/app.js 里的 TAG_EMOJI 的 key 完全一致，
// 否则用户在页面上选到的标签在标签库里查不到，
// 走标签搜索时会直接被 SearchPosts 判成 404（标签不存在）。
var DefTags = []string{
	"校园卡", "身份证证件", "钥匙", "耳机",
	"手机数码", "充电宝", "雨伞", "水杯",
	"书籍笔记", "眼镜", "钱包", "书包",
	"笔袋文具", "手表首饰", "衣物外套",
	"篮球器材", "球拍", "化妆品",
}

// ai帮写的错误处理
// 我恨你
var (
	TagNFound  = errors.New("标签不存在")
	TagInvalid = errors.New("标签名不合法：不能为空、不能超过32个字符、不能包含逗号")
)

// Tag标签库的基准
type Tag struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;size:32;uniqueIndex:idx_tags_name;not null" json:"name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// TableName 把 GORM 模型对应的表名固定为 "tags"：
// 以后就算重命名 Tag 结构体，GORM 也不会按新名字推断出别的表名
func (Tag) TableName() string {
	return "tags"
}

// InitTagLibr 初始化基准标签库：把 DefTags 里的标签全部写进去。
// CreateTag 内部是 FirstOrCreate，已存在的标签不会重复创建，可以反复调用。
//
// 【P0 修复】这个方法原来全项目零调用，导致 tags 表永远是空的：
//   - GET /api/tags 返回空数组 → 前端标签云 / 标签选择器一片空白
//   - 而发布页又强制「必须先选标签」才能提交 → 发帖功能完全不可用
//   - 标签只在发帖时由 NormTag 自动创建，于是形成死锁：发帖要先有标签，标签要靠发帖产生
//
// 现在由 main.go 在连上数据库后立刻调用一次（见 main.go 里的 Tags.InitTagLibr）。
func InitTagLibr(ctx context.Context, db *gorm.DB) error {
	for _, name := range DefTags {
		if _, err := CreateTag(ctx, db, name); err != nil {
			return fmt.Errorf("初始化标签库失败（%s）：%w", name, err)
		}
	}
	return nil
}

// CreateTag往基准标签库写入一个新标签
// 同名标签已存在，直接返回已有，不会报错
func CreateTag(ctx context.Context, db *gorm.DB, name string) (*Tag, error) {
	//标签名先去掉首尾空格（可能用户会有手癖，有就删），再做校验
	name = strings.TrimSpace(name)
	if name == "" || strings.Contains(name, TagSep) || utf8.RuneCountInString(name) > MaxTagLength {
		return nil, TagInvalid
	}

	tag := &Tag{Name: name}
	if err := db.WithContext(ctx).Where("name = ?", name).FirstOrCreate(tag).Error; err != nil {
		return nil, fmt.Errorf("写入标签失败了唔T_T：%w", err)
	}
	return tag, nil
}

// GetAllTags读整个基准标签库,按名称升序
func GetAllTags(ctx context.Context, db *gorm.DB) ([]Tag, error) {
	var tags []Tag
	if err := db.WithContext(ctx).Order("name ASC").Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("读取标签库失败了唔T_T：%w", err)
	}
	return tags, nil
}

// GetTag按名称读取标签
func GetTag(ctx context.Context, db *gorm.DB, name string) (*Tag, error) {
	var tag Tag
	err := db.WithContext(ctx).Where("name = ?", strings.TrimSpace(name)).First(&tag).Error
	if err != nil {
		//cnmAI写的错误处理
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, TagNFound
		}
		return nil, fmt.Errorf("查询标签失败了唔T_T：%w", err)
	}
	return &tag, nil
}

//ParseTags转译成标签的切片

func ParseTags(tags string) []string {
	var (
		result []string
		seen   = make(map[string]bool) //去重用的map,试着跑过一遍，不然看起来好恶心
	)
	for _, part := range strings.Split(tags, TagSep) {
		name := strings.TrimSpace(part)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}

// JoinTags把标签切片拼好
func JoinTags(tags []string) string {
	return strings.Join(tags, TagSep)
}

// NormTag规范化标签
func NormTag(ctx context.Context, db *gorm.DB, rawTags string) (string, error) {
	names := ParseTags(rawTags)
	for _, name := range names {
		// 库里已有就直接用，没有就自动创建TAG
		if _, err := CreateTag(ctx, db, name); err != nil {
			return "", fmt.Errorf("规范化帖子标签失败：%w", err)
		}
	}
	return JoinTags(names), nil
}

// SetPost给指定帖子写标签
// NormTag规范化
func SetPost(ctx context.Context, db *gorm.DB, postID uint64, rawTags string) error {
	normalized, err := NormTag(ctx, db, rawTags)
	if err != nil {
		return err
	}

	// 按帖更新tags列
	result := db.WithContext(ctx).Model(&storage.Post{}).Where("id = ?", postID).Update("tags", normalized)
	if result.Error != nil {
		return fmt.Errorf("更新帖子标签失败了唔T_T：%w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("更新帖子标签失败：帖子 %d 不存在T_T", postID)
	}
	return nil
}

// ApiToTag返回读取标签库的 gin 处理器（HTTP 层）：
func ApiToTag(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tags, err := GetAllTags(c.Request.Context(), db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取标签库失败：" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "获取成功", "tags": tags})
	}
}

func RegRoutes(api *gin.RouterGroup, db *gorm.DB) {
	api.GET("/tags", ApiToTag(db))
}
