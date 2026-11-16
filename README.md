<div align="center">

![new-api](/web/public/logo.png)

# New API

**An AI gateway for models, applications, and agents**

<p align="center">
  <a href="./README.zh_CN.md">简体中文</a> |
  <a href="./README.zh_TW.md">繁體中文</a> |
  <strong>English</strong> |
  <a href="./README.fr.md">Français</a> |
  <a href="./README.ja.md">日本語</a>
</p>

<p align="center">
  <a href="https://raw.githubusercontent.com/Calcium-Ion/new-api/main/LICENSE">
    <img src="https://img.shields.io/github/license/Calcium-Ion/new-api?color=brightgreen" alt="license">
  </a><!--
  --><a href="https://github.com/Calcium-Ion/new-api/releases/latest">
    <img src="https://img.shields.io/github/v/release/Calcium-Ion/new-api?color=brightgreen&include_prereleases" alt="release">
  </a><!--
  --><a href="https://hub.docker.com/r/CalciumIon/new-api">
    <img src="https://img.shields.io/badge/docker-dockerHub-blue" alt="docker">
  </a>
  <a href="https://atomgit.com/QuantumNous/new-api" target="_blank">
    <img alt="AtomGit G-Star" src="https://atomgit.com/QuantumNous/new-api/star/badge.svg"/>
  </a>
</p>

<p align="center">
  <a href="https://trendshift.io/repositories/20180" target="_blank">
    <img src="https://trendshift.io/api/badge/repositories/20180" alt="QuantumNous%2Fnew-api | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/>
  </a>
  <br>
  <a href="https://hellogithub.com/repository/QuantumNous/new-api" target="_blank">
    <img src="https://api.hellogithub.com/v1/widgets/recommend.svg?rid=539ac4217e69431684ad4a0bab768811&claim_uid=tbFPfKIDHpc4TzR" alt="Featured｜HelloGitHub" style="width: 250px; height: 54px;" width="250" height="54" />
  </a><!--
  -->
  <a href="https://atomgit.com/QuantumNous/new-api" target="_blank">
    <img alt="AtomGit G-Star" src="https://atomgit.com/QuantumNous/new-api/star/new_badge.svg" width="250" height="55" />
  </a>
</p>

<p align="center">
  <a href="#screenshots">Screenshots</a> •
  <a href="#capabilities">Capabilities</a> •
  <a href="#quick-start">Quick start</a> •
  <a href="#deployment">Deployment</a> •
  <a href="#development">Development</a> •
  <a href="#documentation">Documentation</a>
</p>

</div>

> **Fork 说明**：本仓库基于 [QuantumNous/new-api](https://github.com/QuantumNous/new-api) 修改，为 Tavo 社区公益站定制了以下功能。上游已移除 classic 主题并把 default 前端收进 `web/src`，下列前端路径按合并后的布局记录。

### ✨ Fork 改动

#### 1. Discord OAuth 登录验证（Discord Gate）
在 Discord OAuth 登录流程中增加了服务器成员身份校验，阻止非社区用户注册：
- **黑名单机制**：检查用户是否在指定黑名单服务器中，在则拒绝登录
- **白名单机制**：检查用户是否在指定白名单服务器中，不在则拒绝登录；可选 `DISCORD_ALLOWED_ROLE_IDS` 再校验身份组
- **检查顺序**：先黑名单 → 再白名单
- **前端改动**：OAuth scope 增加 `guilds.members.read`
- **相关文件**：`oauth/discord.go`、`web/src/lib/oauth.ts`
- **环境变量**：
  - `DISCORD_ALLOWED_GUILD_ID` — 白名单服务器 ID
  - `DISCORD_BANNED_GUILD_IDS` — 黑名单服务器 ID（逗号分隔多个）
  - `DISCORD_ALLOWED_ROLE_IDS` — 可选，白名单服务器内允许登录的身份组 ID

#### 2. 客户端封锁（Client Blocking）
在 API 请求层拦截特定客户端，阻止非授权客户端访问服务。
- 通过环境变量配置拦截规则，不设置则不生效（默认放行所有）
- **环境变量**：
  - `BLOCK_NODEFETCH` — 设为 `true` 拦截特定默认 UA 的请求
  - `BLOCKED_USER_AGENTS` — 自定义 UA 关键词黑名单（逗号分隔，如 `badbot,crawler`）
  - `REQUIRE_CLIENT_HEADER` — 要求请求必须携带指定 Header（如 `X-App-Token`）
  - `REQUIRE_CLIENT_HEADER_VALUE` — 该 Header 的值必须匹配此值（不设置则只检查 Header 存在性）
- **相关文件**：`middleware/client-block.go`、`router/relay-router.go`
- **已注册路由**：`/v1`（OpenAI 兼容）、`/v1beta`（Gemini 兼容），在 TokenAuth 之后、RateLimit 之前生效

#### 3. 经典主题 Service Worker 残留修复
旧 classic 主题留下的 Service Worker 会让 WebView 控制台出现 404。
- 当前唯一前端的 `web/index.html` 会在加载时注销残留 Service Worker

#### 4. WebView / 低版本浏览器兼容性
- **动态导入失败恢复**：路由 chunk 加载失败（含 WebView 的 dynamically imported module 错误）时自动刷新一次
- **构建目标**：Rsbuild browserslist 降到 Chrome 64 / Safari 12 一代（约 es2018）
- **相关文件**：`web/index.html`、`web/rsbuild.config.ts`、`web/src/features/errors/general-error.tsx`

#### 5. 兑换码一码多用
支持一个兑换码被多个用户使用，适用于群发兑换码场景。
- **最大使用次数**：创建兑换码时可设置 `max_uses`（0=一码一用，>0=一码多用）
- **使用记录**：关联表 `redemption_uses` 记录每个用户的使用情况，防止同一用户重复使用
- **状态显示**：多用码部分使用显示为「Partially Used」，用完显示为「Used」
- **管理功能**：管理员可查看兑换码的使用记录列表
- **API**：`GET /api/redemption/:id/uses` 返回使用记录
- **相关文件**：`model/redemption.go`、`model/redemption_use.go`、`controller/redemption.go`、`controller/redemption_use.go`、`web/src/features/redemption-codes/`

#### 6. 用户排行榜
展示用户 API 使用量排行榜，已合并到公共排行榜页面（`/rankings`），不再需要单独登录访问。
- **排行榜数据**：按消耗额度排序，显示用户名、请求数、Token 用量
- **时间段筛选**：支持日、周、月、总四个时间段（复用排行榜页面顶部的 period 切换）
- **进度条可视化**：每个用户下方显示用量占比进度条
- **排名高亮**：前三名金色闪光动画，4-10 名骚粉色高亮
- **API**：
  - `GET /api/user/ranking?period=all&limit=50` — 获取排行榜
  - `GET /api/user/ranking/self?period=all` — 获取当前用户排名
- **前端路由**：`/rankings`（公共页面，无需登录）；旧路由 `/user-ranking` 自动跳转
- **相关文件**：`model/user_ranking.go`、`controller/ranking.go`、`web/src/features/user-ranking/`、`web/src/features/rankings/components/user-ranking-section.tsx`

#### 7. 排行榜增强（Neo-Brutalism 主题配套）
- **排名高亮样式**：`ranking-highlight.css`，前三名金色渐变 + shimmer 动画，4-10 名 `#ff69b4` 高亮
- **进度条可视化**：模型排行榜每行增加 token 占比进度条，按排名自动变色（金/粉/紫）
- **相关文件**：`web/src/styles/ranking-highlight.css`、`web/src/styles/neo-brutalism.css`、`web/src/features/rankings/components/model-leaderboard.tsx`

#### 8. 使用日志表格优化
- **User 列全员可见**：User 列对所有用户可见，并同时显示用户名和 ID
- **管理员新增 IP 列**：管理员视图新增 IP 地址列；详情弹窗对管理员显示 IP
- **详情弹窗 User 行**：日志详情弹窗新增 User 行，同时显示用户名和用户 ID
- **移动端**：用户列进入手机卡片；主题调色盘按钮在手机端可见
- **相关文件**：`web/src/features/usage-logs/components/columns/common-logs-columns.tsx`、`web/src/features/usage-logs/components/dialogs/details-dialog.tsx`、`web/src/features/usage-logs/components/common-log-mobile-card.tsx`

#### 9. 拟物风格主题（Skeuomorphic）
新增「Skeuomorphic」拟物化主题，用数字界面模拟现实世界的材质和结构：
- **设计理念**：金属面板 + 皮革质感 + 黄铜按钮 + 木纹底座
- **色彩语言**：深色金属面板底色（`oklch(0.215 0.012 55)`）+ 黄铜主色（`oklch(0.72 0.145 55)`）+ 皮革棕副色（`oklch(0.42 0.065 45)`）
- **相关文件**：`web/src/styles/theme-presets.css`、`web/src/lib/theme-customization.ts`

#### 10. Noir 主题
超深哑光暗色预设（来自 fork `dev`）：墨色底、哑光玫瑰主色、细边框。
- **相关文件**：`web/src/styles/theme-presets.css`、`web/src/lib/theme-customization.ts`

#### 11. 管理 Bot API Token 与分组接口
- AdminAuth 接受启用中的中继 API Token（`sk-xxx`），以令牌所属用户的角色调用管理接口
- `PATCH /api/user/:id/group` 只更新分组，不影响其他字段
- **相关文件**：`middleware/auth.go`、`controller/user.go`、`model/user.go`、`router/api-router.go`


---

## 📝 Project Description

New API is a self-hosted AI gateway for applications, agents, and teams. Connect upstream model services, expose a consistent API to your clients, and manage routing, access, usage, and costs in one place.

Use it to share authorized model access across a team, switch providers without configuring every client again, or operate a private multi-model service with a web console. Upstreams include OpenAI, Anthropic, Google Gemini, Azure OpenAI, AWS Bedrock, Vertex AI, DeepSeek, Qwen, and other compatible services.

> [!IMPORTANT]
> - This project is intended solely for lawful and authorized AI API gateway, organization-level authentication, multi-model management, usage analytics, cost accounting, and private deployment scenarios.
> - Users must lawfully obtain upstream API keys, accounts, model services, and interface permissions, and must comply with upstream terms of service and applicable laws and regulations.
> - Users should ensure their use complies with upstream terms of service and applicable laws and regulations.
> - When providing generative AI services to the public, users should comply with applicable regulatory requirements and fulfill all filing, licensing, content safety, real-name verification, log retention, tax, and upstream authorization obligations required by their jurisdiction.

<!-- -->

> [!WARNING]
> When operating this project as a public generative AI service or API resale service, users should first complete all required filing, licensing, content safety, real-name verification, log retention, tax, payment, and upstream authorization obligations.

---

<a id="screenshots"></a>

## Screenshots

Accounts, channels, plugin installation states, usage, prices, and costs shown below are simulated data. The marketplace displays the official plugin catalog. Click an image to view it at full size.

| Usage dashboard | Plugin marketplace |
| --- | --- |
| [![Usage dashboard with simulated requests, costs, and performance metrics](assets/screenshots/dashboard.en.jpg)](assets/screenshots/dashboard.en.jpg) | [![Official plugin marketplace with simulated installation status](assets/screenshots/plugin-marketplace.en.jpg)](assets/screenshots/plugin-marketplace.en.jpg) |
| **Model square** | **Usage logs** |
| [![Model square with simulated model pricing and availability](assets/screenshots/models.en.jpg)](assets/screenshots/models.en.jpg) | [![Usage logs with simulated tokens, costs, and response times](assets/screenshots/usage-logs.en.jpg)](assets/screenshots/usage-logs.en.jpg) |

---

## 🤝 Trusted Partners

<p align="center">
  <em>No particular order</em>
</p>

<p align="center">
  <a href="https://www.cherry-ai.com/" target="_blank">
    <img src="./docs/images/cherry-studio.png" alt="Cherry Studio" height="80" />
  </a><!--
  --><a href="https://github.com/iOfficeAI/AionUi/" target="_blank">
    <img src="./docs/images/aionui.png" alt="Aion UI" height="80" />
  </a><!--
  --><a href="https://bda.pku.edu.cn/" target="_blank">
    <img src="./docs/images/pku.png" alt="Peking University" height="80" />
  </a><!--
  --><a href="https://www.aliyun.com/" target="_blank">
    <img src="./docs/images/aliyun.png" alt="Alibaba Cloud" height="80" />
  </a><!--
  --><a href="https://io.net/" target="_blank">
    <img src="./docs/images/io-net.png" alt="IO.NET" height="80" />
  </a>
</p>

---

## 🙏 Special Thanks

<p align="center">
  <a href="https://www.jetbrains.com/?from=new-api" target="_blank">
    <img src="https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.png" alt="JetBrains Logo" width="120" />
  </a>
</p>

<p align="center">
  <strong>Thanks to <a href="https://www.jetbrains.com/?from=new-api">JetBrains</a> for providing free open-source development license for this project</strong>
</p>

---

<a id="capabilities"></a>

## Capabilities

| Area | What you can do |
| --- | --- |
| Model access | Use OpenAI Chat Completions, Responses, Anthropic Messages, and Gemini APIs; stream responses and use tools, reasoning, and multimodal inputs where supported |
| Routing | Configure model mappings, channel priorities and weights, retries, channel affinity, and multiple upstream keys |
| Usage and costs | Manage quotas, subscriptions, usage logs, cache accounting, and expression-based pricing for different usage tiers |
| Access control | Manage users, groups, fine-grained permissions, and API key restrictions; use OAuth/OIDC, passkeys, two-factor authentication, and login session management |
| Asynchronous tasks | Extend image, video, and other task APIs with JavaScript plugins, including task status and output retrieval |
| Web console | Configure channels and models, inspect usage and audit logs, and try models in the playground; available in English, Simplified Chinese, Traditional Chinese, French, Japanese, Russian, and Vietnamese |

### Protocols and endpoints

| Interface | Common endpoints |
| --- | --- |
| OpenAI Chat / Responses | `POST /v1/chat/completions`, `POST /v1/responses` |
| Anthropic Messages | `POST /v1/messages` |
| Gemini | `POST /v1beta/models/{model}:generateContent`, `POST /v1beta/models/{model}:streamGenerateContent` |
| Realtime / Responses WebSocket | `GET /v1/realtime`, `GET /v1/responses` (WebSocket upgrade) |
| Images / audio | `/v1/images/generations`, `/v1/images/edits`, `/v1/audio/speech`, `/v1/audio/transcriptions`, `/v1/audio/translations` |
| Embeddings / rerank | `POST /v1/embeddings`, `POST /v1/rerank` |
| Task plugins | `POST /v1/tasks/{pluginKey}`, `GET /v1/tasks/{taskId}`, plus routes declared by each plugin |

[RelayKit](./relaykit/README.md) provides request, response, and streaming conversion between the four text protocols. Available features depend on the channel, upstream model, and conversion path; protocol-specific tools and fields may not map exactly. WebSocket support also requires a compatible upstream and channel configuration.

This README describes the current source tree. Check the release notes for the version you deploy.

<a id="quick-start"></a>

## Quick start

### Try locally with Docker

This starts a single instance with SQLite and binds it to localhost:

```bash
mkdir -p data
docker run --name new-api -d --restart unless-stopped \
  -p 127.0.0.1:3000:3000 \
  -e TZ=Asia/Shanghai \
  -v "$(pwd)/data:/data" \
  calciumion/new-api:latest
```

Open [http://localhost:3000](http://localhost:3000) and complete the setup wizard to create the administrator account. The `data` directory persists the SQLite database across container replacements.

### Make your first request

1. Add a channel with your upstream API key, available models, and group assignment; run a channel test.
2. Configure model pricing and ensure the user has quota or a valid subscription.
3. Create an API key in the console with access to the same group and models.
4. Set your client's base URL to `http://localhost:3000/v1` for OpenAI-compatible clients and use the **New API-issued key**.

Set `NEW_API_KEY` in your shell to that key. List the models accessible to it:

```bash
curl --fail-with-body http://localhost:3000/v1/models \
  -H "Authorization: Bearer ${NEW_API_KEY}"
```

Then call Responses, replacing `your-enabled-model` with an enabled model that supports this interface:

```bash
curl --fail-with-body http://localhost:3000/v1/responses \
  -H "Authorization: Bearer ${NEW_API_KEY}" \
  -H "Content-Type: application/json" \
  -d '{"model":"your-enabled-model","input":"Hello!"}'
```

<a id="deployment"></a>

## Deployment

### Docker Compose

The repository's [Compose configuration](./docker-compose.yml) starts **New API + PostgreSQL + Redis** by default. It also contains examples for MySQL and a separate ClickHouse log database.

```bash
git clone https://github.com/QuantumNous/new-api.git
cd new-api
```

Before starting, edit `docker-compose.yml`: replace the database and Redis example passwords in both the services and connection strings, and set a persistent random `SESSION_SECRET` (generate one with `openssl rand -hex 32`). For an HTTPS console, configure `SESSION_COOKIE_SECURE=true` and `SESSION_COOKIE_TRUSTED_URL` with its exact public HTTPS origin.

```bash
docker compose up -d
docker compose logs -f new-api
```

### Storage and configuration

| Component | Options |
| --- | --- |
| Main database | SQLite, MySQL ≥ 5.7.8, or PostgreSQL ≥ 9.6 |
| Separate log database | Configure with `LOG_SQL_DSN`; also supports ClickHouse |
| Cache | Optional Redis plus in-memory caching; use shared Redis when application nodes need shared rate limits |
| Container platforms | Linux amd64 / arm64 |

| Variable | Purpose |
| --- | --- |
| `SQL_DSN` | Main database connection; unset uses SQLite |
| `LOG_SQL_DSN` | Optional separate log database connection |
| `REDIS_CONN_STRING` | Redis connection string |
| `SESSION_SECRET` | Persistent authentication secret; all nodes must use the same value |
| `CRYPTO_SECRET` | Defaults to `SESSION_SECRET`; nodes sharing Redis must use the same effective value |
| `SESSION_COOKIE_SECURE` | Set to `true` for an HTTPS console; enables Secure refresh cookies and strict refresh/logout origin checks |
| `SESSION_COOKIE_TRUSTED_URL` | Required in Secure mode: comma-separated exact HTTPS origins, without paths or wildcards; leave unset for local HTTP |
| `TRUSTED_PROXIES` | Trusted reverse-proxy IPs/CIDRs, or `none`; explicitly configure for your network |

See the [environment example](./.env.example), [environment reference](https://docs.newapi.ai/en/docs/installation/config-maintenance/environment-variables), and [authentication and session guide](./docs/authentication.md) for full configuration. Configure container variables in Compose's `environment` or `env_file`; copying `.env.example` alone does not inject variables into the container.

For production, put the console behind HTTPS and configure your reverse proxy for streaming and WebSocket upgrades. Persist and back up the database and mounted data. Multi-node deployments must share the main database and authentication secrets; separate Redis instances or in-memory rate limiters count limits independently per node. The session guide describes propagation behavior for each topology.

Pin an image version from [Releases](https://github.com/QuantumNous/new-api/releases), review its upgrade notes, and back up before upgrading. The `latest` tag follows published builds and can change; migrations and compatibility must be assessed for your existing installation.

<a id="development"></a>

## Development and extensions

The backend uses Go and Gin. The web console uses React 19, TypeScript, Rsbuild, TanStack, and Tailwind CSS 4. Use Bun for frontend dependencies and scripts; see [go.mod](./go.mod) for the Go language baseline and [Dockerfile](./Dockerfile) for the container build toolchain.

Build the frontend before starting the backend, which embeds `web/dist`:

```bash
# Repository root
cd web
bun install --frozen-lockfile
bun run build
cd ..
go run .
```

In a second terminal, start the frontend development server:

```bash
cd web
bun run dev -- --port 5173
```

Open [http://localhost:5173](http://localhost:5173); the development server proxies API requests to the backend on port 3000. For a containerized development backend, see [docker-compose.dev.yml](./docker-compose.dev.yml) and the `make dev` target in [makefile](./makefile).

| Location | Responsibility |
| --- | --- |
| `router/`, `middleware/`, `controller/` | HTTP routes, access checks, and API handlers |
| `relay/` | Upstream adapters and request routing |
| `service/`, `model/` | Business logic and persistence |
| [relaykit/](./relaykit/README.md) | Independently buildable Go module for protocol DTOs and conversions |
| [plugins/tasks/](./plugins/tasks/) | JavaScript task plugins; see [Task Plugin API v1](./docs/plugin-api/v1.md) for authoring and host boundaries |
| `web/` | Web console; see [frontend conventions](./web/AGENTS.md) |
| [electron/](./electron/README.md) | Desktop wrapper and packaging |

Read [AGENTS.md](./AGENTS.md) before contributing. Run checks appropriate to your change, including `make test` for the Go modules and `bun check -b`, `bun run lint`, `bun run test`, and `bun run build` in `web/` for frontend changes. Changes to RelayKit must also pass `GOWORK=off go build ./...` from `relaykit/`.

<a id="documentation"></a>

## Documentation and community

| Resource | Link |
| --- | --- |
| Official documentation | [Guides](https://docs.newapi.ai/en/docs) · [Installation](https://docs.newapi.ai/en/docs/installation) · [API reference](https://docs.newapi.ai/en/docs/api) |
| Project exploration | [DeepWiki](https://deepwiki.com/QuantumNous/new-api) |
| Questions and discussion | [FAQ](https://docs.newapi.ai/en/docs/support/faq) · [Community](https://docs.newapi.ai/en/docs/support/community-interaction) |
| Bugs and feature requests | [GitHub Issues](https://github.com/QuantumNous/new-api/issues) |
| Security reports | Follow the [security policy](./.github/SECURITY.md) for private reporting |

For bug reports, include the version, deployment method, reproduction steps, and redacted logs. Documentation, translations, provider integrations, and focused regression tests are all welcome contributions.

---

## 🔗 Related Projects

### Upstream Projects

| Project | Description |
|------|------|
| [One API](https://github.com/songquanpeng/one-api) | Original project base |
| [Midjourney-Proxy](https://github.com/novicezk/midjourney-proxy) | Midjourney interface support |

### Supporting Tools

| Project | Description |
|------|------|
| [new-api-key-tool](https://github.com/Calcium-Ion/new-api-key-tool) | Key quota query tool |
| [new-api-horizon](https://github.com/Calcium-Ion/new-api-horizon) | New API high-performance optimized version |

---

## 📜 License

This project is licensed under the [GNU Affero General Public License v3.0 (AGPLv3)](./LICENSE).

Additional terms under AGPLv3 Section 7 apply. Modified versions must preserve
the author attribution notice `Frontend design and development by New API
contributors.` in the appropriate legal notices and in any prominent about,
legal, footer, or attribution location presented by the user interface.

Modified versions that present a user interface must also preserve a visible
link to the original project: <https://github.com/QuantumNous/new-api>.

This is an open-source project developed based on [One API](https://github.com/songquanpeng/one-api) (MIT License).

If your organization's policies do not permit the use of AGPLv3-licensed software, or if you wish to avoid the open-source obligations of AGPLv3, please contact us at: [support@quantumnous.com](mailto:support@quantumnous.com)

See [NOTICE](./NOTICE) and [third-party licenses](./THIRD-PARTY-LICENSES.md) for attribution and dependency notices.

---

## 🌟 Star History

<div align="center">

[![Star History Chart](https://api.star-history.com/svg?repos=Calcium-Ion/new-api&type=Date)](https://star-history.com/#Calcium-Ion/new-api&Date)

</div>

---

<div align="center">

### 💖 Thank you for using New API

If this project is helpful to you, welcome to give us a ⭐️ Star！

**[Official Documentation](https://docs.newapi.ai/en/docs)** • **[Issue Feedback](https://github.com/Calcium-Ion/new-api/issues)** • **[Latest Release](https://github.com/Calcium-Ion/new-api/releases)**

<sub>Built with ❤️ by QuantumNous</sub>

</div>
