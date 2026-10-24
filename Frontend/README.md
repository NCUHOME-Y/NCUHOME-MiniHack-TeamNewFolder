# Program 前端（对接 Program 现有后端）

这是给 `Program/Backend` 配的**网页版前端**，纯 HTML + CSS + 原生 JS，**没有改动后端**。

## 运行

后端地址默认写死在 `app.js` 顶部两行，按需修改：

```js
const API_ORIGIN = 'http://127.0.0.1:8080';
const API_BASE = API_ORIGIN + '/api';
```

然后任选一种方式打开前端：

1. 直接双击 `index.html`（浏览器打开本地文件）；或
2. 用任意静态服务器托管本目录，例如：
   ```bash
   cd Program/Frontend
   python -m http.server 5173   # 打开 http://localhost:5173
   ```

启动后端（在 `Program/Backend` 下 `go run .`）后，前端即可调用接口。

> 注意：后端 CORS 会把任意 Origin 回显，所以前端跨端口调用也能通；如果以后收紧 CORS，需要把前端地址加进白名单。

## 已实现功能

| 页面 | 功能 | 后端接口 |
|---|---|---|
| 首页 | 帖子列表、全部/寻物/招领筛选、分页 | `GET /api/posts` |
| 搜索 | 关键词模糊搜索、标签筛选、标签云 | `GET /api/search/posts`、`GET /api/tags` |
| 发布 | 上传图片、选标签、发帖 | `POST /api/update/upload`、`POST /api/update/create/post` |
| 详情 | 帖子信息 + 评论列表/回复/删除 | `GET /api/posts/:id/comments`、`POST /api/update/comments`、`DELETE /api/update/comments/:id` |
| 废贴库 | 显示已结帖（status=1）的帖子 | `GET /api/posts`（前端按 `status` 过滤） |
| 个人中心 | 我的资料、编辑简介/头像、我的帖子 | `GET /api/users/me`、`POST /api/update/bio`、`POST /api/update/avatar` |
| 登录/注册 | 登录、注册、登出 | `POST /api/users/login`、`POST /api/users/register`、`POST /api/update/logout` |

## 针对现有后端的适配说明（重要）

Program 后端和 Program2 不一样，前端做了如下适配，**没动后端**：

1. **响应格式不统一**：注册返回 `registerinfo`，登录返回 `token`；列表用 `data.list`，搜索用 `posts`；标签返回对象数组 `{id,name,created_at}`。前端在 `app.js` 里各自解析。
2. **没有帖子详情接口**：详情页复用「列表项数据」+「评论接口」。因此：
   - 详情页只显示封面图（列表里只有单张 `image_url`，图集 `image_urls` 拿不到）；
   - 不显示评论数（列表项没有 `comment_count`）。
3. **没有结帖接口**：所以「标记已归还 / 认领」按钮没有实现；废贴库改为前端按 `status === 1` 过滤列表。
4. **没有“我的帖子”接口**：前端拉取列表后按 `username === 当前用户名` 过滤。
5. **联系方式接口浏览器拿不到**：后端 `GET /api/posts/:id/contactinfo` 是用 GET 读 JSON 请求体的，浏览器 fetch 无法给 GET 带 body，因此详情页没有做联系方式展示，改为引导用户「在留言区联系发帖人」。
6. **`/api/users/me` 的 `bio` 是对象**：后端直接序列化了 `sql.NullString`，返回 `{"String":"...","Valid":true}`；前端已做了兼容转换。

这些都是后端当前实现带来的限制；如果之后后端补齐了详情/结帖/联系方式接口，前端可以直接对应扩展。
