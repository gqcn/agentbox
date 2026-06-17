## Context

当前仓库已经具备源码插件目录规范、统一插件 API 命名空间`/x/{plugin-id}/api/v1/...`、插件 public assets 托管路径`/x-assets/{plugin-id}/{version}/...`和默认管理工作台入口`/admin`。`john-ai-agentbox`要提供独立产品入口，但该需求来自插件自身展示入口，不应扩展`apps/lina-core`核心宿主协议。入口实现必须闭环在插件内部，继续复用现有源码插件路由注册和 public assets 能力。

外部项目`/Users/john/Workspace/github/gqcn/agent-box`是一个 GoFrame v2 后端和 React/Vite 前端组成的独立应用。现有后端使用`/api`注册 JSON API，使用`/ws/agents/...`提供 Chat/Shell WebSocket，使用服务代理或 preview handler 支持 Agent runtime 服务访问；前端硬编码了`/api`、`/ws`等根级路径。现有 SQL 包含`ai_providers`、`provider_models`、`coding_agents`、`agent_chat_sessions`、`users`、`user_sessions`、`ai_capability_tiers`、`ai_invocation_logs`等表，需要迁移到插件自有表空间。

本项目是全新项目，不需要为外部`agent-box`保留旧路径兼容。迁移目标是一次性进入 LinaPro 插件治理，而不是先做反向代理或保留双应用部署。

## Goals / Non-Goals

**Goals:**

- 将 AgentBox 迁移为源码插件`john-ai-agentbox`，通过`apps/lina-plugins/john-ai-agentbox/`维护完整后端、前端、SQL、配置和测试资源。
- 让`/`进入 AgentBox 功能界面，`/login`进入 AgentBox 独立登录页，`/admin`继续打开 LinaPro 管理工作台。
- 将 AgentBox API、WebSocket、preview 和服务代理全部收敛到插件拥有的命名空间，避免占用宿主`/api`、根级`/ws`或根级代理路径。
- 保留 AgentBox 独立用户、独立会话和`agent_box_session`Cookie 语义，确保它不影响 LinaPro 管理工作台登录状态。
- 将 AgentBox 自有`AI`能力表全部重命名为`john_ai_agentbox_`前缀，避免与宿主`AI`能力和`linapro-ai-core`插件数据重合。
- 补充插件门户、认证、API、SQL、前端和 E2E 验证边界，使后续实现可按任务逐步落地。

**Non-Goals:**

- 不把 AgentBox 改造成`apps/lina-core`内建模块。
- 不把 AgentBox 页面强制嵌入 LinaPro 管理工作台 SPA 或左侧菜单。
- 不复用 LinaPro 管理工作台 JWT、用户、角色、权限、在线用户或登录日志作为 AgentBox 产品登录态。
- 不复用`linapro-ai-core`的渠道、模型、档位、调用日志或宿主文本`AI`能力作为 AgentBox 自身数据源。
- 不为外部`agent-box`旧路径`/api`、`/ws`、`/s`、`/assets`保留生产兼容入口。
- 不在提案阶段实现代码；本变更仅建立后续实现的 OpenSpec 契约。

## Decisions

### 以源码插件迁移，而不是宿主内建模块

`john-ai-agentbox`作为源码插件随宿主编译和发布。插件目录遵守统一结构：`plugin.yaml`、`plugin_embed.go`、`backend/`、`frontend/`、`manifest/`、`README.md`和`README.zh-CN.md`。后端使用插件同构结构`backend/api/`、`backend/internal/controller/`、`backend/internal/service/`、`backend/internal/dao/`和`backend/plugin.go`。

选择源码插件的原因是 AgentBox 需要后端 Go 运行时、Docker/容器编排、WebSocket、静态资产和 SQL 生命周期资源。源码插件可以复用宿主启动装配、插件启停检查、GoFrame 路由注册和 public assets 机制，同时保持业务代码归属插件目录。替代方案是把 AgentBox 作为宿主模块或独立服务反向代理；前者会污染`lina-core`边界，后者无法纳入插件生命周期和 SQL 治理。

### 插件内部精确页面入口，不扩展宿主 fallback

AgentBox 页面入口通过`john-ai-agentbox`源码插件自身的`http.route.register`回调注册两个精确路由：`GET /`和`GET /login`。这两个入口不注册通配符，不提供宿主级 SPA fallback，也不修改`plugin.yaml`清单协议。`GET /`直接返回插件嵌入的`frontend/dist/index.html`作为 AgentBox 工作台入口，浏览器地址保持为`/`，不把用户重定向到`/x-assets/john-ai-agentbox/{version}/index.html`；静态 JS/CSS 继续由宿主已有`/x-assets`能力按`public_assets`边界提供。

`/login`入口同样直接返回插件嵌入的`frontend/dist/index.html`，前端根据当前浏览器路径渲染独立登录页，避免浏览器刷新依赖宿主 catch-all。登录成功后前端把地址栏恢复为`/`并回到 AgentBox 应用壳。`/admin`、`/api/**`、`/x/**`、`/x-assets/**`和`/api.json`继续由宿主既有路由与资产处理，不引入新的主框架路由优先级规则。

曾评估过新增宿主通用插件门户 fallback，但这会修改`apps/lina-core`清单字段、运行时 fallback、缓存和测试，超出本次插件迁移边界。也不采用插件`/*`通配 fallback，因为该方案会吞掉宿主和其他插件路径。精确`GET /`与`GET /login`是当前范围内最小实现。

### 路由映射一次性收敛到插件命名空间

AgentBox 路由迁移采用以下目标映射：

| 原路径类别 | 目标路径 | 说明 |
|------------|----------|------|
| 前端入口 | `/`、`/login` | 插件精确入口直接返回 AgentBox SPA `index.html`，地址栏保持入口路径 |
| 管理工作台 | `/admin`、`/admin/*` | 继续由 LinaPro 默认工作台处理 |
| JSON API | `/x/john-ai-agentbox/api/v1/...` | 通过源码插件 registrar 的`APIPrefix()`注册 |
| WebSocket | `/x/john-ai-agentbox/api/v1/ws/...` | 由插件注册，必须复用插件启用检查和 AgentBox 会话鉴权 |
| 服务代理和 preview | `/x/john-ai-agentbox/api/v1/proxy/...`或`/x/john-ai-agentbox/api/v1/previews/...` | 不再占用根级`/s`或`/preview` |
| public assets | `/x-assets/john-ai-agentbox/{version}/...` | 仅发布`plugin.yaml public_assets`声明目录 |

前端必须通过集中配置的 API base 和 WebSocket base 构造请求，不得在组件中继续硬编码`/api`或`/ws`根路径。后端 DTO 的`g.Meta path`保持资源路径语义，例如`/agents`、`/providers`，由注册前缀统一映射到完整公开路径。

### 独立认证和会话边界

AgentBox 继续维护自己的用户表、会话表、密码哈希、会话 token 哈希和 HttpOnly Cookie。Cookie 名称保留`agent_box_session`，与 LinaPro 管理工作台 token 不同。由于插件门户位于`/`，Cookie path 可以继续使用`/`；但宿主管理工作台不得读取或解释该 Cookie，AgentBox 插件也不得读取 LinaPro 管理 Cookie 作为登录凭证。

`POST /x/john-ai-agentbox/api/v1/auth/sessions`用于登录并写入 Cookie，`GET /x/john-ai-agentbox/api/v1/auth/session`读取当前 AgentBox 会话，`DELETE /x/john-ai-agentbox/api/v1/auth/session`注销并清理 Cookie。`/login`路由只是页面入口，不承担后端认证动作。

### 插件自有数据模型和表前缀

所有从外部`agent-box`迁入的数据表、索引、外键、唯一约束、Seed DML 和 Mock 数据都必须改为`john_ai_agentbox_`前缀。建议映射包括：

| 原表 | 目标表 |
|------|--------|
| `users` | `john_ai_agentbox_users` |
| `user_sessions` | `john_ai_agentbox_user_sessions` |
| `ai_providers` | `john_ai_agentbox_ai_providers` |
| `provider_models` | `john_ai_agentbox_provider_models` |
| `coding_images` | `john_ai_agentbox_coding_images` |
| `coding_agents` | `john_ai_agentbox_coding_agents` |
| `agent_runtimes` | `john_ai_agentbox_agent_runtimes` |
| `agent_chat_sessions` | `john_ai_agentbox_agent_chat_sessions` |
| `agent_chat_messages` | `john_ai_agentbox_agent_chat_messages` |
| `agent_chat_interactions` | `john_ai_agentbox_agent_chat_interactions` |
| `agent_terminal_sessions` | `john_ai_agentbox_agent_terminal_sessions` |
| `agent_box_settings` | `john_ai_agentbox_settings` |
| `agent_box_user_settings` | `john_ai_agentbox_user_settings` |
| `ai_capability_tiers` | `john_ai_agentbox_ai_capability_tiers` |
| `ai_capability_bindings` | `john_ai_agentbox_ai_capability_bindings` |
| `ai_invocation_logs` | `john_ai_agentbox_ai_invocation_logs` |
| `system_prompt_overrides` | `john_ai_agentbox_system_prompt_overrides` |

插件 SQL 放在`apps/lina-plugins/john-ai-agentbox/manifest/sql/`，卸载 SQL 放在`manifest/sql/uninstall/`，演示数据放在`manifest/sql/mock-data/`。SQL 源使用 PostgreSQL 14+ 语法，必须幂等，不显式写入自增 ID，并为列表、筛选、Agent ownership、会话历史、日志时间范围、provider/model 关联和运行时状态查询建立必要索引。DAO 生成结果必须位于插件自己的`backend/internal/dao/`和`backend/internal/model/`，禁止依赖宿主 DAO。

### 数据权限和用户可见性

AgentBox 的产品用户体系独立于 LinaPro 角色管理，因此插件自身业务数据采用插件内用户 ID 作为权威可见性边界。列表、详情、下载、工作区资源、Chat 会话、Terminal 会话、服务代理和写操作都必须在查询或动作执行前校验当前 AgentBox 用户对目标 Agent、workspace、session 或 runtime 的归属关系。

这属于数据权限规则的业务例外：不接入 LinaPro 角色管理数据权限，但不能放开全量数据。例外依据是 AgentBox 不是管理工作台功能，它拥有独立登录态和自有数据域。拒绝策略为返回认证失败、权限不足或不可见资源，不通过名称、数量、分页总数或代理路径泄露其他用户数据存在性。

### 性能和查询装配

迁移不得保留会诱导前端逐项补查的接口契约。Agent 列表需要一次返回当前页所需 provider、model、image、runtime 和 activity 投影；provider 列表需要集合化装配模型摘要；Chat 会话和消息按 Agent/session 范围分页或有界查询；调用日志按数据库侧过滤、排序和分页；workspace tree 和服务发现必须有范围、深度或刷新边界。

插件服务层应优先使用批量查询、集合化查询、投影查询和必要缓存。高频接口不得在动态结果集循环中逐行查询 provider、model、image、session 或 runtime。

### 缓存和运行时状态

AgentBox 会持有会话、Agent runtime、服务发现、WebSocket 连接、Terminal 后端、AI 档位解析或配置读取等状态。实现期必须区分权威数据源和缓存状态：用户会话、Agent 配置、Chat 历史和 AI 配置以插件数据库为权威；WebSocket 连接、运行中进程和服务发现是运行时状态。若引入缓存，必须声明单机和集群行为、失效触发点、最大陈旧时间和故障降级。

首期可以按单实例本地运行时实现 AgentBox runtime，但必须在设计和任务记录中明确集群模式限制：当`cluster.enabled=true`时，不能把插件可用性、会话认证或权限判断退化为仅当前节点可见的默认实例。

### i18n 边界

`john-ai-agentbox`初期按单语言插件治理，`plugin.yaml`可以不启用`i18n.enabled: true`。这种情况下，插件自己的运行时 UI 文案和 API 文档源文本可以沿用当前中文或英文源内容，不要求补齐插件`manifest/i18n`和`apidoc`翻译资源，也不得把插件翻译键写入`lina-core`。如果后续启用插件 i18n，必须使用插件自己的`manifest/i18n/<locale>/`和`manifest/i18n/<locale>/apidoc/`资源。

### 开发工具和构建

前端迁移应优先复用现有 React/Vite 构建方式，但输出必须进入插件 public asset 或 embedded asset 视图。若需要新增长期维护的构建、资源打包或代码生成入口，应使用跨平台 Go 工具或现有`linactl`能力；不得新增只适用于 macOS/Linux 的 shell 脚本作为默认入口。

## Risks / Trade-offs

- [Risk] 插件根路径入口可能与宿主工作台根路径部署模式冲突。→ Mitigation：本轮不修改宿主根工作台策略；AgentBox 只注册精确`/`和`/login`入口，`/admin`继续归宿主管理。如需根工作台与插件门户动态切换，另行确认宿主能力变更。
- [Risk] 外部项目表名和 SQL 较多，手工重命名容易漏掉外键、索引或代码查询。→ Mitigation：先建立表名映射清单，再通过静态检索旧表名、SQL smoke 和插件 DAO 生成验证无残留。
- [Risk] 前端硬编码`/api`、`/ws`较多，迁移后可能出现部分请求仍打到宿主根路径。→ Mitigation：集中抽象 API base、WebSocket base 和 asset base，并用静态检索和 E2E 覆盖。
- [Risk] AgentBox 独立登录与 LinaPro 管理登录并存，用户可能混淆状态。→ Mitigation：`/login`只服务 AgentBox，`/admin`只服务 LinaPro；Cookie 名称、API 和退出动作完全独立，E2E 验证两者互不影响。
- [Risk] Docker runtime、Terminal、WebSocket 和服务代理迁入插件后启动依赖复杂。→ Mitigation：所有运行期依赖通过插件构造和启动装配显式传入，缺失时返回初始化错误，不在请求路径临时`New()`关键服务。
- [Risk] 迁移后插件仍可能复用宿主`AI`能力或`linapro-ai-core`表。→ Mitigation：规范和任务明确禁止，数据库表前缀、DAO 包路径和静态检索共同验证。
- [Risk] 单实例 runtime 状态在集群部署下不可迁移。→ Mitigation：首期记录集群限制；若要支持集群，需要另立变更设计跨节点 runtime 调度、会话粘滞或共享协调。

## Migration Plan

1. 创建`apps/lina-plugins/john-ai-agentbox/`插件骨架、`plugin.yaml`、`plugin_embed.go`、`backend/plugin.go`、`frontend/`、`manifest/`和目录说明文档。
2. 迁移 AgentBox 后端代码到插件同构目录，按 LinaPro GoFrame 插件规范重写包路径、Controller、Service、Middleware、路由注册和错误处理。
3. 将 API、WebSocket、preview 和服务代理路径迁移到`/x/john-ai-agentbox/api/v1/...`，同步更新前端 API base 和 WebSocket base。
4. 合并和重写外部 SQL 为插件安装 SQL，完成表名前缀、索引、唯一约束、Seed DML、卸载 SQL 和可选 mock 数据迁移，再执行 DAO 生成。
5. 迁移 React/Vite 前端为插件 public asset，配置`public_assets`和源码插件精确`/`、`/login`入口，确保页面入口不依赖宿主新增 fallback。
6. 补充单元测试、插件后端编译门禁、SQL 静态检查、前端构建验证和插件专属 E2E。
7. 静态检索旧包路径、旧表名、旧根级路径和宿主`AI`数据访问，确认无残留后进入 lina-review。

Rollback 策略：本项目无兼容负担，若实现阶段失败，应禁用或移除`john-ai-agentbox`插件和门户配置，回退到`/admin`管理工作台可访问状态；插件安装 SQL 必须提供卸载路径以清理插件自有表。

## Open Questions

- AgentBox runtime 是否需要在首期支持`cluster.enabled=true`的多节点调度，还是明确首期仅支持单实例 Agent runtime。
- AgentBox public assets 采用插件`public_assets`统一托管，还是由源码插件自管 HTTP 静态资源 handler；两者都允许，但实现期应选择一种作为主路径。
- 是否将外部`agent-box`的历史 OpenSpec 规格迁入当前变更归档，还是只以本次插件迁移规格作为当前仓库事实来源。
