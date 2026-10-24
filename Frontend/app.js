/* ============================================================
 * 校内失物招领 · 前端逻辑（对接 Program 现有后端，不改后端）
 *
 * 后端地址：默认 http://127.0.0.1:8080
 * 如果你把后端跑在别的端口/主机，改下面两行即可。
 * ============================================================ */

const API_ORIGIN = 'http://127.0.0.1:8080';
const API_BASE = API_ORIGIN + '/api';

/* ---------- 状态 ---------- */
const store = {
  token: localStorage.getItem('token') || '',
  user: null,
  filter: 'all',        // all | lost | found
  page: 1,
  pageSize: 12,
  searchKeyword: '',
  searchTag: '',
  detail: null,         // 点击进入详情时暂存的帖子对象
  replyTo: null,        // 评论时引用的楼层
  postType: 'lost',
  postTag: null,
  uploads: [],          // 发帖时已上传的图片 url
  postsById: {},        // 详情页复用：没有详情接口，点击时从这里取完整列表项
};

function cachePosts(list) {
  (list || []).forEach(p => { if (p && p.id != null) store.postsById[p.id] = p; });
}

try {
  store.user = JSON.parse(localStorage.getItem('user') || 'null');
} catch (e) {
  store.user = null;
}

/* ---------- 工具 ---------- */
function esc(s) {
  return String(s == null ? '' : s)
    .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
}
function toast(msg) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.classList.add('show');
  clearTimeout(t._h);
  t._h = setTimeout(() => t.classList.remove('show'), 1800);
}
function timeAgo(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  const diff = Date.now() - d.getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return '刚刚';
  if (m < 60) return m + ' 分钟前';
  const h = Math.floor(m / 60);
  if (h < 24) return h + ' 小时前';
  const days = Math.floor(h / 24);
  if (days < 30) return days + ' 天前';
  return d.toLocaleDateString('zh-CN');
}
function fullTime(iso) {
  if (!iso) return '';
  return new Date(iso).toLocaleString('zh-CN');
}

const TAG_EMOJI = {
  '校园卡': '💳', '身份证证件': '🪪', '钥匙': '🔑', '耳机': '🎧',
  '手机数码': '📱', '充电宝': '🔋', '雨伞': '☂️', '水杯': '🥤',
  '书籍笔记': '📚', '眼镜': '👓', '钱包': '👛', '书包': '🎒',
  '笔袋文具': '✏️', '手表首饰': '⌚', '衣物外套': '🧥',
  '篮球器材': '🏀', '球拍': '🏸', '化妆品': '💄',
};
function tagEmoji(tag) { return TAG_EMOJI[tag] || '📦'; }
function typeName(t) { return t === 1 ? '寻物' : (t === 2 ? '招领' : ''); }
function typeClass(t) { return t === 1 ? 'lost' : 'found'; }

function isImage(u) {
  return !!u && (/^https?:\/\//.test(u) || u.startsWith('/uploads/') || u.startsWith('/'));
}
function fullImg(u) {
  if (!u) return '';
  if (/^https?:\/\//.test(u)) return u;
  if (u.startsWith('/')) return API_ORIGIN + u;
  return u;
}
function avatarHTML(user) {
  const name = (user && user.username) || '?';
  if (user && user.avatar_url && isImage(user.avatar_url)) {
    return `<img src="${esc(fullImg(user.avatar_url))}" alt="${esc(name)}">`;
  }
  return esc(name.slice(0, 1));
}
function isAuthed() { return !!store.token; }
function clearAuth() {
  store.token = '';
  store.user = null;
  localStorage.removeItem('token');
  localStorage.removeItem('user');
  renderUserArea();
}
function setAuth(token, username) {
  store.token = token;
  store.user = { username: username, avatar_url: '', bio: '' };
  localStorage.setItem('token', token);
  localStorage.setItem('user', JSON.stringify(store.user));
  renderUserArea();
}

/* ---------- 请求封装（Program 后端响应格式不统一，这里返回原始 JSON） ---------- */
async function request(path, opts = {}) {
  const headers = {};
  if (store.token) headers['Token'] = store.token;
  let body = opts.body;
  if (opts.body !== undefined && !(opts.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
    body = JSON.stringify(opts.body);
  }
  const res = await fetch(API_BASE + path, { method: opts.method || 'GET', headers, body });
  const text = await res.text();
  let data = null;
  try { data = text ? JSON.parse(text) : null; } catch (e) { data = text; }

  if (!res.ok) {
    if (res.status === 401 && store.token) clearAuth();
    const msg = (data && data.error) || (typeof data === 'string' ? data : '') || ('请求失败 (' + res.status + ')');
    const err = new Error(msg);
    err.status = res.status;
    err.data = data;
    throw err;
  }
  return data;
}

/* ---------- 后端接口 ---------- */
async function apiLogin(u, p) { return (await request('/users/login', { method: 'POST', body: { username: u, password: p } })).token; }
async function apiRegister(u, p) { return (await request('/users/register', { method: 'POST', body: { username: u, password: p } })).registerinfo; }
async function apiMe() {
  const data = await request('/users/me');
  const b = data.bio;
  const bio = (b && typeof b === 'object') ? (b.String || '') : (b || '');
  return { username: data.username, avatar_url: data.avatar_url, bio: bio };
}
async function apiLogout() { return request('/update/logout', { method: 'POST' }); }
async function apiUpdateBio(bio) { return request('/update/bio', { method: 'POST', body: { bio } }); }
async function apiUpdateAvatar(avatar_url) { return request('/update/avatar', { method: 'POST', body: { avatar_url } }); }
async function apiCreatePost(p) { return request('/update/create/post', { method: 'POST', body: p }); }
async function apiPostList(params) { return request('/posts?' + params); }
async function apiCommentList(id) { return request('/posts/' + id + '/comments?page=1&page_size=50'); }
async function apiCreateComment(p) { return request('/update/comments', { method: 'POST', body: p }); }
async function apiDeleteComment(id) { return request('/update/comments/' + id, { method: 'DELETE' }); }
async function apiUpload(file) {
  const fd = new FormData();
  fd.append('file', file);
  return (await request('/update/upload', { method: 'POST', body: fd })).url;
}
async function apiTags() { return (await request('/tags')).tags || []; }
async function apiSearch(params) { return request('/search/posts?' + params); }

/* ---------- 认证弹窗 ---------- */
let authMode = 'login';
function openAuth(mode) {
  if (mode) authMode = mode;
  document.getElementById('auth-modal').hidden = false;
  document.querySelectorAll('#auth-modal .modal-tabs button').forEach(b => {
    b.classList.toggle('on', b.dataset.auth === authMode);
  });
  document.getElementById('auth-submit').textContent = authMode === 'login' ? '登录' : '注册';
  document.getElementById('auth-hint').textContent = '';
}
function closeAuth() { document.getElementById('auth-modal').hidden = true; }

async function submitAuth(e) {
  e.preventDefault();
  const form = document.getElementById('auth-form');
  const username = form.username.value.trim();
  const password = form.password.value;
  const hint = document.getElementById('auth-hint');
  hint.textContent = '';
  if (authMode === 'register' && password.length < 8) {
    hint.textContent = '密码至少 8 位';
    return;
  }
  try {
    const token = authMode === 'login'
      ? await apiLogin(username, password)
      : await apiRegister(username, password);
    setAuth(token, username);
    closeAuth();
    toast(authMode === 'login' ? '登录成功' : '注册成功');
    try { store.user = await apiMe(); localStorage.setItem('user', JSON.stringify(store.user)); renderUserArea(); } catch (e) {}
    render();
  } catch (err) {
    hint.textContent = err.message;
  }
}

/* ---------- 用户区 ---------- */
function renderUserArea() {
  const el = document.getElementById('user-area');
  if (isAuthed() && store.user) {
    el.innerHTML = `
      <a class="user-chip" href="#/profile">
        <span class="avatar">${avatarHTML(store.user)}</span>
        <span>${esc(store.user.username)}</span>
      </a>
      <button class="btn" id="logout-btn">退出</button>`;
    document.getElementById('logout-btn').onclick = async () => {
      try { await apiLogout(); } catch (e) {}
      clearAuth();
      toast('已退出登录');
      location.hash = '#/home';
      render();
    };
  } else {
    el.innerHTML = `<button class="btn" id="login-btn">登录</button>
      <button class="btn-primary" id="register-btn">注册</button>`;
    document.getElementById('login-btn').onclick = () => openAuth('login');
    document.getElementById('register-btn').onclick = () => openAuth('register');
  }
}

/* ---------- 路由 ---------- */
function nav(tab) {
  document.querySelectorAll('.nav a').forEach(a => a.classList.toggle('on', a.dataset.nav === tab));
}

async function render() {
  const hash = location.hash || '#/home';
  const app = document.getElementById('app');
  renderUserArea();
  if (hash.startsWith('#/detail/')) { nav(''); await renderDetail(hash.slice('#/detail/'.length)); return; }
  if (hash === '#/search') { nav('search'); renderSearch(); return; }
  if (hash === '#/post') {
    nav('post');
    if (!isAuthed()) { openAuth('login'); toast('请先登录'); location.hash = '#/home'; renderHome(); return; }
    renderPostForm();
    return;
  }
  if (hash === '#/archive') { nav('archive'); renderArchive(); return; }
  if (hash === '#/profile') {
    nav('');
    if (!isAuthed()) { openAuth('login'); return; }
    renderProfile();
    return;
  }
  nav('home');
  renderHome();
}

/* ---------- 卡片 ---------- */
function postCardHTML(p) {
  const thumb = isImage(p.image_url)
    ? `<img src="${esc(fullImg(p.image_url))}" alt="">`
    : `<span class="emoji">${tagEmoji(p.tags ? p.tags.split(',')[0] : '')}</span>`;
  const tags = (p.tags || '').split(',').map(t => t.trim()).filter(Boolean);
  const author = p.username
    ? `<span class="author"><span class="avatar">${avatarHTML({ username: p.username, avatar_url: p.avatar_url })}</span>${esc(p.username)}</span>`
    : '';
  return `
    <div class="card ${p.status === 1 ? 'archived' : ''}" onclick="openDetail('${p.id}')">
      <div class="card-thumb">${thumb}${p.status === 1 ? '<div class="stamp">已结帖</div>' : ''}</div>
      <div class="badges">
        <span class="badge ${typeClass(p.type)}">${typeName(p.type)}</span>
        ${tags.map(t => `<span class="tag">${esc(t)}</span>`).join('')}
      </div>
      <div class="card-title">${esc(p.title)}</div>
      <div class="card-desc">${esc(p.description)}</div>
      <div class="card-meta">
        <span>📍 ${esc(p.location)}</span>
        <span>${timeAgo(p.created_at)}</span>
        ${author}
      </div>
    </div>`;
}

function emptyHTML(msg) {
  return `<div class="empty"><span class="emoji">📭</span>${esc(msg)}</div>`;
}
function pagerHTML(page, total, pageSize, fn) {
  fn = fn || 'goPage';
  const pages = Math.max(1, Math.ceil(total / pageSize));
  if (pages <= 1) return '';
  return `<div class="pager">
    <button ${page <= 1 ? 'disabled' : ''} onclick="${fn}(${page - 1})">上一页</button>
    <span>${page} / ${pages}</span>
    <button ${page >= pages ? 'disabled' : ''} onclick="${fn}(${page + 1})">下一页</button>
  </div>`;
}

/* ---------- 首页 ---------- */
function renderHome() {
  const app = document.getElementById('app');
  app.innerHTML = `
    <div class="page-head"><h1>失物招领</h1><p>捡到东西挂出来，丢了东西找回来 · 校内互助</p></div>
    <div class="search-bar" onclick="location.hash='#/search'">
      <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/></svg>
      <input placeholder="搜索物品 / 地点关键词" readonly>
    </div>
    <div class="seg">
      <button class="${store.filter === 'all' ? 'on' : ''}" onclick="setFilter('all')">全部</button>
      <button class="${store.filter === 'lost' ? 'on' : ''}" onclick="setFilter('lost')">寻物</button>
      <button class="${store.filter === 'found' ? 'on' : ''}" onclick="setFilter('found')">招领</button>
    </div>
    <div class="grid" id="home-list"><div class="empty">加载中…</div></div>
    <div id="home-pager"></div>`;
  loadHome();
}
function setFilter(f) { store.filter = f; store.page = 1; renderHome(); }
async function goPage(p) { store.page = p; await loadHome(); }
async function loadHome() {
  const params = new URLSearchParams({ page: store.page, page_size: store.pageSize });
  if (store.filter !== 'all') params.set('type', store.filter === 'lost' ? '1' : '2');
  try {
    const data = await apiPostList(params.toString());
    const list = data.data.list || [];
    cachePosts(list);
    document.getElementById('home-list').innerHTML = list.length ? list.map(postCardHTML).join('') : emptyHTML('暂无帖子，点顶部「发布」发第一条');
    document.getElementById('home-pager').innerHTML = pagerHTML(data.data.page, data.data.total, data.data.page_size);
  } catch (e) {
    document.getElementById('home-list').innerHTML = emptyHTML(e.message);
  }
}

/* ---------- 搜索 ---------- */
function renderSearch() {
  const app = document.getElementById('app');
  app.innerHTML = `
    <div class="page-head"><h1>搜索</h1><p>按关键词模糊搜索，或按标签精确筛选</p></div>
    <div class="search-bar">
      <svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/></svg>
      <input id="kw" placeholder="输入物品 / 地点关键词，回车搜索" value="${esc(store.searchKeyword)}">
      <button class="btn-primary" onclick="doKeywordSearch()">搜索</button>
    </div>
    <div class="section-title" style="font-size:13px;font-weight:700;color:var(--sub);margin:4px 0 10px">常用标签 · 点一下筛选</div>
    <div class="tag-cloud" id="tag-cloud">加载中…</div>
    <div class="grid" id="search-list" style="margin-top:20px"></div>
    <div id="search-pager"></div>`;
  document.getElementById('kw').addEventListener('keydown', e => { if (e.key === 'Enter') doKeywordSearch(); });
  loadTags();
  if (store.searchKeyword || store.searchTag) doSearch();
}
async function loadTags() {
  try {
    const tags = await apiTags();
    document.getElementById('tag-cloud').innerHTML = tags.map(t =>
      `<span class="tag-chip ${store.searchTag === t.name ? 'on' : ''}" onclick="quickTag('${esc(t.name)}')">${tagEmoji(t.name)} ${esc(t.name)}</span>`
    ).join('');
  } catch (e) {
    document.getElementById('tag-cloud').innerHTML = emptyHTML(e.message);
  }
}
function doKeywordSearch() {
  store.searchKeyword = document.getElementById('kw').value.trim();
  store.searchTag = '';
  store.page = 1;
  doSearch();
}
function quickTag(t) {
  store.searchTag = t;
  store.searchKeyword = '';
  store.page = 1;
  const input = document.getElementById('kw');
  if (input) input.value = '';
  loadTags();
  doSearch();
}
async function goSearchPage(p) { store.page = p; await doSearch(); }
async function doSearch() {
  const params = new URLSearchParams({ page: store.page, page_size: 10 });
  if (store.searchKeyword) params.set('keyword', store.searchKeyword);
  if (store.searchTag) params.set('tag', store.searchTag);
  try {
    const data = await apiSearch(params.toString());
    const list = data.posts || [];
    cachePosts(list);
    document.getElementById('search-list').innerHTML = list.length ? list.map(postCardHTML).join('') : emptyHTML('没有找到相关帖子');
    document.getElementById('search-pager').innerHTML = pagerHTML(data.page, data.total, data.page_size, 'goSearchPage');
  } catch (e) {
    document.getElementById('search-list').innerHTML = emptyHTML(e.message);
  }
}

/* ---------- 发帖 ---------- */
function renderPostForm() {
  const app = document.getElementById('app');
  app.innerHTML = `
    <div class="page-head"><h1>发布帖子</h1><p>如实描述，方便失主/拾到者核对</p></div>
    <div class="form">
      <div class="seg">
        <button class="${store.postType === 'lost' ? 'on' : ''}" onclick="setPostType('lost')">我要寻物</button>
        <button class="${store.postType === 'found' ? 'on' : ''}" onclick="setPostType('found')">我捡到了</button>
      </div>
      <div class="field">
        <label>物品照片（选填，最多 9 张）</label>
        <div class="upload-grid" id="upload-grid">
          <div class="upload-box" onclick="pickImage()"><span style="font-size:24px">＋</span><span>上传图片</span></div>
        </div>
        <input type="file" id="file-input" accept="image/*" multiple hidden>
      </div>
      <div class="field">
        <label>选择标签</label>
        <div class="pick-grid" id="tag-picker">加载中…</div>
      </div>
      <div class="field">
        <label>标题</label>
        <input type="text" id="p-title" maxlength="100" placeholder="一句话说清丢了/捡到什么">
        <div class="counter"><span id="title-count">0</span>/100</div>
      </div>
      <div class="field">
        <label>详细描述</label>
        <textarea id="p-desc" maxlength="5000" placeholder="补充时间、地点、物品特征、识别暗号（用于防冒领）"></textarea>
        <div class="counter"><span id="desc-count">0</span>/5000</div>
      </div>
      <div class="field">
        <label>所在地点</label>
        <input type="text" id="p-loc" maxlength="255" placeholder="如：图书馆三楼">
      </div>
      <div class="field">
        <label>联系方式</label>
        <input type="text" id="p-contact" maxlength="100" placeholder="QQ / 微信号 / 手机号">
      </div>
      <button class="btn-primary" id="submit-post">发布</button>
    </div>`;
  document.getElementById('p-title').addEventListener('input', e => document.getElementById('title-count').textContent = e.target.value.length);
  document.getElementById('p-desc').addEventListener('input', e => document.getElementById('desc-count').textContent = e.target.value.length);
  document.getElementById('submit-post').onclick = submitPost;
  document.getElementById('file-input').addEventListener('change', onFilePick);
  loadTagPicker();
  renderUploads();
}
async function loadTagPicker() {
  try {
    const tags = await apiTags();
    document.getElementById('tag-picker').innerHTML = tags.map(t =>
      `<span class="tag-chip ${store.postTag === t.name ? 'on' : ''}" onclick="setPostTag('${esc(t.name)}')">${tagEmoji(t.name)} ${esc(t.name)}</span>`
    ).join('');
  } catch (e) {}
}
function setPostType(t) { store.postType = t; renderPostForm(); }
function setPostTag(t) { store.postTag = t; renderPostForm(); }
function pickImage() { document.getElementById('file-input').click(); }
function renderUploads() {
  const grid = document.getElementById('upload-grid');
  if (!grid) return;
  grid.innerHTML = store.uploads.map((u, i) => `
    <div class="upload-box"><img src="${esc(fullImg(u))}"><button class="remove" onclick="removeUpload(${i})">×</button></div>`
  ).join('') + `
    <div class="upload-box" onclick="pickImage()"><span style="font-size:24px">＋</span><span>上传图片</span></div>`;
}
function removeUpload(i) { store.uploads.splice(i, 1); renderUploads(); }
async function onFilePick(e) {
  const files = Array.from(e.target.files || []);
  if (store.uploads.length + files.length > 9) { toast('最多 9 张'); return; }
  for (const f of files) {
    if (!f.type.startsWith('image/')) { toast('请选择图片'); continue; }
    try {
      const url = await apiUpload(f);
      store.uploads.push(url);
      renderUploads();
    } catch (err) {
      toast(err.message);
    }
  }
  e.target.value = '';
}
async function submitPost() {
  const btn = document.getElementById('submit-post');
  const title = document.getElementById('p-title').value.trim();
  const desc = document.getElementById('p-desc').value.trim();
  const loc = document.getElementById('p-loc').value.trim();
  const contact = document.getElementById('p-contact').value.trim();
  if (!store.postTag) { toast('请先选一个标签'); return; }
  if (!title || !desc || !loc || !contact) { toast('请填写标题、描述、地点和联系方式'); return; }
  btn.disabled = true; btn.textContent = '发布中…';
  try {
    await apiCreatePost({
      type: store.postType === 'lost' ? 1 : 2,
      title, description: desc, image_urls: store.uploads,
      tags: store.postTag, location: loc, contact_info: contact,
    });
    store.postTag = null;
    store.uploads = [];
    store.filter = 'all';
    toast('发布成功');
    location.hash = '#/home';
    render();
  } catch (e) {
    toast(e.message);
  } finally {
    btn.disabled = false; btn.textContent = '发布';
  }
}

/* ---------- 详情 ---------- */
async function ensurePost(id) {
  const nid = Number(id);
  if (store.postsById[nid] && store.postsById[nid].title != null) return store.postsById[nid];
  try {
    const data = await apiPostList('page=1&page_size=50');
    cachePosts(data.data.list || []);
  } catch (e) {}
  return store.postsById[nid] || { id: nid };
}
function openDetail(id) {
  // Program 后端没有帖子详情接口，详情页复用列表项 + 评论接口
  store.detail = store.postsById[id] || { id: Number(id) };
  location.hash = '#/detail/' + id;
}
async function renderDetail(id) {
  const app = document.getElementById('app');
  app.innerHTML = `<div class="page"><div class="empty">加载中…</div></div>`;
  store.replyTo = null;
  const p = await ensurePost(id);
  store.detail = p;
  const hero = isImage(p.image_url)
    ? `<img src="${esc(fullImg(p.image_url))}">`
    : `<span class="emoji">${tagEmoji(p.tags ? p.tags.split(',')[0] : '')}</span>`;
  const tags = (p.tags || '').split(',').map(t => t.trim()).filter(Boolean);
  const author = { username: p.username || '', avatar_url: p.avatar_url || '' };

  app.innerHTML = `
    <div class="detail">
      <div class="detail-hero">${hero}${p.status === 1 ? '<div class="stamp">已结帖</div>' : ''}</div>
      <div class="badges" style="margin-bottom:8px">
        <span class="badge ${typeClass(p.type)}">${typeName(p.type)}</span>
        ${tags.map(t => `<span class="tag">${esc(t)}</span>`).join('')}
        <span class="badge archived">${p.status === 1 ? '已结帖' : '进行中'}</span>
      </div>
      <div class="detail-title">${esc(p.title || '')}</div>
      <div class="detail-desc">${esc(p.description || '')}</div>
      <div class="info-row"><svg viewBox="0 0 24 24"><path d="M12 21s-7-5.5-7-11a7 7 0 0 1 14 0c0 5.5-7 11-7 11z"/><circle cx="12" cy="10" r="2.5"/></svg>${esc(p.location || '')}</div>
      <div class="info-row"><svg viewBox="0 0 24 24"><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></svg>${fullTime(p.created_at)}</div>
      <div class="contact-card">想联系发帖人？在下方留言说明来意，发帖人会看到。</div>
      <div class="author-row">
        <span class="avatar" style="width:38px;height:38px">${avatarHTML(author)}</span>
        <div><div class="name">${esc(author.username || '匿名')}</div><div class="time">发布于 ${timeAgo(p.created_at)}</div></div>
      </div>
    </div>
    <div class="comments">
      <h3>留言（<span id="cmt-total">0</span>）</h3>
      <div id="cmt-list">加载中…</div>
      <div class="comment-input">
        <input id="cmt-input" placeholder="友善留言，约定取物时间地点">
        <button class="btn-primary" onclick="addComment(${Number(id)})">发送</button>
      </div>
    </div>`;
  loadComments(Number(id));
}

async function loadComments(postId) {
  try {
    const data = await apiCommentList(postId);
    const list = data.data.list || [];
    document.getElementById('cmt-total').textContent = data.data.total;
    document.getElementById('cmt-list').innerHTML = list.length ? list.map(c => {
      const replied = c.reply_to ? list.find(x => x.floor === c.reply_to) : null;
      const canDel = isAuthed() && store.user && store.user.username === c.username;
      return `
        <div class="cmt">
          <span class="avatar">${avatarHTML({ username: c.username, avatar_url: c.avatar_url })}</span>
          <div class="cmt-body">
            <div class="cmt-head">
              <span class="uname">${esc(c.username)}</span>
              <span>#${c.floor}楼</span>
              <span>${timeAgo(c.created_at)}</span>
              ${canDel ? `<button class="del" onclick="delComment(${c.id})">删除</button>` : ''}
            </div>
            <div class="cmt-text">${esc(c.content)}</div>
            ${c.image_url && isImage(c.image_url) ? `<img src="${esc(fullImg(c.image_url))}" style="max-width:120px;border-radius:8px;margin-top:6px">` : ''}
            ${replied ? `<div class="cmt-reply"><b>@${esc(replied.username)}：</b>${esc(replied.content)}</div>` : (c.reply_to ? `<div class="cmt-reply"><b>引用楼层已删除</b></div>` : '')}
            <div class="cmt-actions"><button onclick="setReply(${c.floor}, '${esc(c.username)}')">回复</button></div>
          </div>
        </div>`;
    }).join('') : '<div class="empty">还没有留言，来帮个忙吧</div>';
  } catch (e) {
    document.getElementById('cmt-list').innerHTML = emptyHTML(e.message);
  }
}
function setReply(floor, name) {
  store.replyTo = floor;
  const input = document.getElementById('cmt-input');
  if (input) input.placeholder = `回复 #${floor} 楼 ${name}`;
  input.focus();
}
async function addComment(postId) {
  if (!isAuthed()) { openAuth('login'); toast('请先登录'); return; }
  const input = document.getElementById('cmt-input');
  const content = input.value.trim();
  if (!content) { toast('先写点内容'); return; }
  try {
    await apiCreateComment({ post_id: postId, content, reply_to: store.replyTo });
    input.value = '';
    store.replyTo = null;
    input.placeholder = '友善留言，约定取物时间地点';
    await loadComments(postId);
    toast('留言成功');
  } catch (e) {
    toast(e.message);
  }
}
async function delComment(id) {
  if (!confirm('确定删除这条留言？')) return;
  try {
    await apiDeleteComment(id);
    toast('删除成功');
    await loadComments(store.detail.id);
  } catch (e) {
    toast(e.message);
  }
}

/* ---------- 废贴库（后端没有状态筛选，前端按 status 过滤） ---------- */
async function renderArchive() {
  const app = document.getElementById('app');
  app.innerHTML = `
    <div class="page-head"><h1>废贴库</h1><p>这些物品已经物归原主</p></div>
    <div class="grid" id="archive-list"><div class="empty">加载中…</div></div>`;
  try {
    const data = await apiPostList('page=1&page_size=50');
    const list = (data.data.list || []).filter(p => p.status === 1);
    cachePosts(data.data.list);
    document.getElementById('archive-list').innerHTML = list.length ? list.map(postCardHTML).join('') : emptyHTML('暂无归档');
  } catch (e) {
    document.getElementById('archive-list').innerHTML = emptyHTML(e.message);
  }
}

/* ---------- 个人中心 ---------- */
async function renderProfile() {
  const app = document.getElementById('app');
  app.innerHTML = `<div class="page"><div class="empty">加载中…</div></div>`;
  try {
    const me = await apiMe();
    store.user = me;
    localStorage.setItem('user', JSON.stringify(me));
    renderUserArea();
    const data = await apiPostList('page=1&page_size=50');
    const all = data.data.list || [];
    cachePosts(all);
    const mine = all.filter(p => p.username === me.username);
    app.innerHTML = `
      <div class="page-head"><h1>个人中心</h1></div>
      <div class="detail" style="margin-bottom:18px">
        <div style="display:flex;align-items:center;gap:16px">
          <span class="avatar" style="width:64px;height:64px;font-size:24px">${avatarHTML(me)}</span>
          <div style="flex:1">
            <div class="detail-title" style="margin:0">${esc(me.username)}</div>
            <div style="color:var(--sub);font-size:13px">${esc(me.bio || '这个人很懒，什么都没写')}</div>
          </div>
          <button class="btn" onclick="openProfileEdit()">编辑资料</button>
        </div>
      </div>
      <div class="page-head" style="margin-top:8px"><h2 style="font-size:18px">我发布的帖子（${mine.length}）</h2></div>
      <div class="grid">${mine.length ? mine.map(postCardHTML).join('') : emptyHTML('还没有发过帖子')}</div>`;
  } catch (e) {
    app.innerHTML = emptyHTML(e.message);
  }
}
function openProfileEdit() {
  const app = document.getElementById('app');
  const u = store.user || {};
  app.innerHTML = `
    <div class="page-head"><h1>编辑资料</h1></div>
    <div class="form">
      <div class="field">
        <label>头像地址（http(s) 或 /uploads/...）</label>
        <input type="text" id="e-avatar" value="${esc(u.avatar_url || '')}" placeholder="留空则保持默认">
      </div>
      <div class="field">
        <label>个人简介</label>
        <textarea id="e-bio" maxlength="500">${esc(u.bio || '')}</textarea>
      </div>
      <button class="btn-primary" onclick="submitProfile()">保存</button>
    </div>`;
}
async function submitProfile() {
  const avatar = document.getElementById('e-avatar').value.trim();
  const bio = document.getElementById('e-bio').value;
  try {
    if (bio !== undefined) await apiUpdateBio(bio);
    if (avatar !== '') await apiUpdateAvatar(avatar);
    toast('保存成功');
    store.user = await apiMe();
    localStorage.setItem('user', JSON.stringify(store.user));
    location.hash = '#/profile';
    render();
  } catch (e) {
    toast(e.message);
  }
}

/* ---------- 全局事件 ---------- */
document.getElementById('auth-form').addEventListener('submit', submitAuth);
document.getElementById('auth-close').onclick = closeAuth;
document.querySelectorAll('#auth-modal .modal-tabs button').forEach(b => {
  b.onclick = () => { authMode = b.dataset.auth; openAuth(); };
});
window.addEventListener('hashchange', render);

/* 启动 */
(async function boot() {
  if (isAuthed() && store.user) {
    try { store.user = await apiMe(); localStorage.setItem('user', JSON.stringify(store.user)); } catch (e) {}
  }
  render();
})();
