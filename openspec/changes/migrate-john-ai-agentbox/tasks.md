## 1. 插件内部页面入口

- [x] 1.1 设计并实现`john-ai-agentbox`插件内部精确页面入口，使用现有源码插件`http.route.register`注册`GET /`和`GET /login`。
- [x] 1.2 确保页面入口不注册通配 fallback，不吞掉`/api/**`、`/x/**`、`/x-assets/**`、`/admin/**`或`/api.json`。
- [x] 1.3 确保页面入口只转到插件已声明的`public_assets`入口资源，不发布 SQL、配置、后端源码或未授权目录。
- [x] 1.4 记录主框架边界影响：本轮不修改`apps/lina-core`主框架代码、插件清单协议、宿主 fallback、宿主缓存或宿主 README。

实现记录：

| 项目 | 记录 |
|------|------|
| 主框架边界 | 本轮不修改`apps/lina-core`。未新增`plugin.yaml portal_routes`、宿主 SPA fallback、宿主门户缓存、宿主 README 或宿主路由测试。 |
| 页面入口 | 插件`backend/routes/portal.go`通过现有源码插件路由注册精确`GET /`和`GET /login`。`/`直接返回 AgentBox 工作台 SPA 的`index.html`并保持浏览器地址为`/`；`/login`直接返回同一 SPA 入口并由前端按路径渲染独立登录页。 |
| 保留路径 | 插件只注册两个精确入口，不注册`/*`或其他通配 fallback；`/api/**`、`/x/**`、`/x-assets/**`、`/admin/**`和`/api.json`继续由宿主既有路由处理。 |
| DI 来源检查 | 本阶段未新增关键运行期服务实例；页面入口仅使用插件路由注册回调绑定轻量 handler，不在请求路径创建关键 service 图。 |
| 缓存一致性 | 无新增缓存、快照、派生状态、失效或跨实例同步。页面入口在插件启动注册时读取嵌入的`frontend/dist/index.html`作为不可变发布资产，静态 JS/CSS 仍由宿主既有`/x-assets`能力按插件版本和`public_assets`声明提供。 |
| 数据权限影响 | 页面入口只暴露公开登录/入口资产，不读取或装配业务数据；AgentBox 业务数据权限由 5.x 任务实现。 |
| `i18n`影响 | 插件首期未启用`i18n.enabled: true`，按单语言插件治理；本阶段不新增宿主 UI 文案、API 文档源文本、错误 fallback 或语言包资源。 |
| 开发工具影响 | 执行了既有`make plugins.init`把`apps/lina-plugins`从历史 submodule 转为普通目录，属于项目已发布跨平台工具入口；未新增脚本或长期维护工具。 |
| 验证 | `go test ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`覆盖插件精确页面入口和插件 API 路由绑定；静态扫描确认`apps/lina-core`无本次迁移 diff。 |

## 2. 插件骨架和资源边界

- [x] 2.1 创建`apps/lina-plugins/john-ai-agentbox/`源码插件目录，修改前先检查该目录是否存在插件本地`AGENTS.md`并遵守其规则。
- [x] 2.2 创建并校验`plugin.yaml`、`plugin_embed.go`、`backend/plugin.go`、`backend/api/`、`backend/internal/`、`frontend/`、`manifest/`、`README.md`和`README.zh-CN.md`。
- [x] 2.3 配置插件 public assets 或自管门户资源入口，确保只发布声明的前端资源，不公开`manifest/sql/`、`manifest/config/`、后端源码或未授权目录。
- [x] 2.4 静态检索插件 ID、目录名、清单 ID、URL path、资产路径和测试目录，确认均使用`john-ai-agentbox`。

实现记录：

| 项目 | 记录 |
|------|------|
| 插件本地规则 | 创建和修改前检查`apps/lina-plugins/john-ai-agentbox/AGENTS.md`，文件不存在，继续遵守项目顶层和命中规则。 |
| 插件工作区 | 通过既有`make plugins.init`将`apps/lina-plugins`从历史 submodule 转为普通插件目录，保留目录内容并使新插件文件进入当前项目维护范围。 |
| 骨架资源 | 已创建`plugin.yaml`、`plugin_embed.go`、`go.mod`、`backend/plugin.go`、`backend/api/`、`backend/internal/{controller,service,dao,model}/`、`frontend/pages/`、`frontend/slots/`、`frontend/dist/index.html`、`manifest/config/`、`manifest/sql/`、`manifest/sql/uninstall/`、`manifest/sql/mock-data/`、插件 E2E 目录和双语`README`。 |
| 资源发布边界 | `plugin.yaml public_assets`仅声明`frontend/dist`；`plugin_embed.go`仅嵌入插件清单、后端入口、配置模板和`frontend/dist/*`，不发布 SQL、后端内部源码或未授权目录。 |
| `i18n`影响 | `plugin.yaml`未配置`i18n.enabled: true`，插件首期按单语言治理；未创建插件`manifest/i18n`，也未向`apps/lina-core`写入插件翻译资源。 |
| 编译验证 | 使用临时`temp/go.work.agentbox`执行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/... -count=1`通过；临时`go.work`文件已删除。 |
| 静态验证 | `find apps/lina-plugins/john-ai-agentbox -maxdepth 4 -type f | sort`确认目录资源存在；`rg -n "john-ai-agentbox|john_ai_agentbox|agent_box_session|/x/john-ai-agentbox|public_assets" apps/lina-plugins/john-ai-agentbox openspec/changes/migrate-john-ai-agentbox`确认插件 ID、URL path 和资产声明一致；`git diff --check`通过。 |

## 3. 后端迁移和路由收敛

- [x] 3.1 将外部 AgentBox 后端 Go 代码迁移到插件`backend/`同构目录，重写 module import、包注释、文件用途注释和显式构造函数依赖。
- [x] 3.2 将 JSON API 注册到`/x/john-ai-agentbox/api/v1/...`，同步修正 DTO `g.Meta`、HTTP 方法、`dc`、`eg`和 Unix 毫秒时间字段。
- [x] 3.3 将 Chat/Shell WebSocket、workspace preview、service proxy 和 bridge/tunnel 入口迁移到`/x/john-ai-agentbox/api/v1/...`下的插件路径。
- [x] 3.4 将 AgentBox 鉴权中间件改为插件自有会话校验，确保不读取 LinaPro 管理工作台登录态作为 AgentBox 凭据。
- [x] 3.5 将用户可见错误、认证失败和权限失败收敛为 LinaPro `bizerr`或等价结构化业务错误，避免裸字符串进入 HTTP 响应。
- [x] 3.6 运行插件后端变更包和路由绑定相关 Go 编译门禁，至少覆盖`apps/lina-plugins/john-ai-agentbox/backend/...`和受影响宿主启动路由包。

实现记录：

| 项目 | 记录 |
|------|------|
| JSON API 迁移切片 | 已将 AgentBox JSON API 注册到`/x/john-ai-agentbox/api/v1`，覆盖`auth`、`agents`、`agents/{id}/chat`、`agents/{id}/terminal/sessions`、`agents/{id}/workspace/paths`、`agents/{id}/workspace/tree`、`agents/{id}/workspace/file`、`agents/{id}/workspace/files`、`agents/{id}/workspace/directories`、`agents/{id}/workspace/upload`、`agents/{id}/workspace/download`、`agents/{id}/workspace/resources`、`agents/{id}/workspace/html-previews`、`agents/{id}/skills`、`agents/{id}/skills/upload`、`agents/{id}/git/*`、`agents/{id}/services`、`agents/{id}/service-bridges`、`health/docker`、`containers`、`ai/capability-tiers`、`ai/invocations`、`providers`、`images`、`settings`和`prompt-templates`；对应 DTO、HTTP 方法、`dc`/`eg`和 Unix 毫秒时间字段已按插件路径重新声明。`settings`接口读写`john_ai_agentbox_user_settings`，按当前 AgentBox 用户隔离工作台偏好，不复用 LinaPro 管理工作台用户偏好或宿主 AI 数据。`prompt-templates`接口读写插件自有`john_ai_agentbox_system_prompt_overrides`，用于独立维护 AgentBox 系统提示词覆盖内容，不复用宿主 AI 提示词或其他插件数据。`ai`接口读写插件自有`john_ai_agentbox_ai_capability_tiers`、`john_ai_agentbox_ai_capability_bindings`和`john_ai_agentbox_ai_invocation_logs`，provider/model 关联只通过插件`john_ai_agentbox_ai_providers`和`john_ai_agentbox_provider_models`解析，不复用宿主 AI 或`linapro-ai-core`数据。当前真实 workspace 文件系统、受控 Docker runtime lifecycle、service discovery 和 HTML preview 已作为插件路径下的有界子集迁移；AI commit message、Git 写操作、skills upload、bridge 创建/删除、HTTP proxy relay、WebSocket Shell/Chat 和 TCP tunnel 的真实流式后端仍按插件路径保留结构化不可用，而不是回退到旧根级入口。 |
| AI 能力接口 | 已新增插件`backend/api/ai`、`backend/internal/controller/ai`和`backend/internal/service/ai`，覆盖`GET /ai/capability-tiers`、`PUT /ai/capability-tiers/{code}`、`POST /ai/capability-tiers/{code}/test`和`GET /ai/invocations`。能力档位列表采用固定小集合批量装配：tiers 一次查询、primary bindings 一次 join provider/model、latest test logs 一次有界查询后内存映射；调用日志按数据库侧 purpose/tier/status 过滤并限制默认 50、最大 200 行，provider 名称按 ID 集合批量读取，避免前端瀑布式补查和后端逐行查询。 |
| Agent Chat JSON 接口 | 已新增`backend/api/agent/v1/chat.go`、`backend/internal/service/chat`和`backend/internal/controller/agent/*chat*`，覆盖`GET/POST /agents/{id}/chat/sessions`、`GET/PUT/DELETE /agents/{id}/chat/sessions/{sessionId}`、`GET /agents/{id}/chat/sessions/{sessionId}/messages`、`POST /agents/{id}/chat/sessions/{sessionId}/recover`、`GET /agents/{id}/chat/sessions/{sessionId}/interactions`、`GET /agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}`、`PUT /agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}/response`和`PUT /agents/{id}/chat/sessions/{sessionId}/interactions/{interactionId}/status`。Chat session、message 和 interaction 查询均先校验当前 AgentBox 用户对 Agent/session 的归属；`recover`当前只完成可见性校验并返回结构化`CodeChatRuntimeUnavailable`，真实 runtime/WebSocket 恢复入口已迁到插件命名空间并保持结构化不可用。 |
| Terminal JSON 元数据接口 | 已新增`backend/api/agent/v1/terminal.go`、`backend/internal/service/terminal`和`backend/internal/controller/agent/*terminal*`，覆盖`GET/POST /agents/{id}/terminal/sessions`和`GET/DELETE /agents/{id}/terminal/sessions/{terminalId}`。Terminal session 查询和变更均先校验当前 AgentBox 用户对 Agent 或 Terminal session 的归属；时间点字段投影为 Unix 毫秒，状态和 backend 类型使用命名常量。该接口迁移持久化元数据和资源边界；Shell WebSocket attach 已迁到插件`/ws/...`路径并在真实流式后端不可用时返回结构化错误。 |
| Workspace JSON/资源/Git 子集 | 已扩展`backend/api/workspace`、`backend/internal/service/workspace`和`backend/internal/controller/workspace`，覆盖`GET /agents/{id}/workspace/paths`、`GET /agents/{id}/workspace/tree`、`GET/PUT /agents/{id}/workspace/file`、`POST /agents/{id}/workspace/files`、`POST /agents/{id}/workspace/directories`、`POST /agents/{id}/workspace/upload`、`GET /agents/{id}/workspace/download`、`GET /agents/{id}/workspace/resources`、`GET /agents/{id}/workspace/html-previews`、`GET /agents/{id}/skills`、`POST /agents/{id}/skills/upload`、`GET /agents/{id}/git/status`、`GET /agents/{id}/git/file`、`GET /agents/{id}/git/diff`、`POST /agents/{id}/git/commit-message-suggestions`、`PUT/DELETE /agents/{id}/git/index`、`DELETE /agents/{id}/git/changes`和`POST /agents/{id}/git/commits`。这些入口先规范化 workspace 路径或 Git repository-relative file，再校验当前 AgentBox 用户对目标 Agent/workspace resource 的归属。当前已迁移 Docker workspace runtime 子集：`workspace.New(accessSvc, runtimeBackend)`显式接收启动期共享 Docker backend，`paths/tree/file preview/upload/download/resource/html preview/save/create file/create directory`在可见 Agent 上通过同一用户与 Agent 标签解析运行中容器，限定路径在`/home/agent/workspace`或`/home/agent/shared`下，返回真实目录候选、即时 tree、1MiB 以内文本/图片预览元数据、最多 16 个且单文件 10MiB 以内的 multipart 上传写入、`download/resources`二进制流、受限 HTML 预览流、UTF-8 文本写入和空文件/目录创建。`GET /git/status`、`GET /git/file`和`GET /git/diff`已迁移为只读 runtime Git 子集：通过 Docker exec 在受控路径内执行`git rev-parse`、`git status --porcelain=v1 -z`、`git diff`、`git diff --cached`、`git diff --no-index`和`git show`，返回`ok/clean/not_repo`、unstaged/staged 变更列表、树形投影、文件预览、统一 diff 和可编辑 side-by-side 文本模型；Git file/diff 的 repository-relative file 会先清理并限定在当前 repo root 下，repo root 与最终文件路径仍需留在`/home/agent/workspace`或`/home/agent/shared`内，不执行仓库写操作。`GET /skills`已迁移为只读 runtime 技能发现：project scope 扫描当前 workspace 的`.agents/skills`直接子目录，global scope 只扫描容器内`CODEX_HOME/skills`、`$HOME/.codex/skills`和`$HOME/.agents/skills`三个固定来源，最多返回 200 条，只读取每个技能目录 256KiB 以内的`SKILL.md`元数据并解析`name`/`description`，不执行技能代码、不解压上传包、不写文件。`WorkspaceOpenFile`通过 Docker`CopyFromContainer`返回 tar 中首个文件的流式 reader，不为下载预读完整文件；`WorkspaceWriteFile`、`WorkspaceUploadFile`和文件创建通过 Docker`CopyToContainer`写入受控 tar；目录创建通过受控`mkdir`执行。`save`在前端传入`baseHash`时先读取当前文件并比较`sha256`，不一致返回结构化`CodeWorkspaceStateConflict`；`upload`只接受安全单段文件名，目标必须是受控目录；`resources`支持`inline`和`attachment`，对 HTML、SVG、未知二进制等不安全 inline 内容强制降级为 attachment并写入`X-Content-Type-Options: nosniff`；`html-previews`只接受`.html/.htm`文件，复用同一受控流式读取路径并写入`Content-Security-Policy: sandbox`、`frame-ancestors 'none'`、`form-action 'none'`和`Cache-Control: no-store`，不执行脚本、不提交表单、不授予同源 Cookie 访问，也不代理关联 runtime 服务。runtime 缺失、容器未运行、路径不存在、目录下载、目录写入或 Docker exec/copy 失败时包装为结构化`CodeWorkspaceRuntimeUnavailable`。skills upload、Git write 和 AI commit message 入口已迁到插件路径，当前在真实后端不可用时返回结构化不可用或受控响应，不再依赖旧根路径。 |
| Container JSON 子集 | 已新增`backend/api/container`、`backend/internal/service/container`和`backend/internal/controller/container`，覆盖`GET /health/docker`、`GET/POST /containers`、`GET /containers/{id}`、`POST /containers/{id}/start`、`POST /containers/{id}/stop`、`DELETE /containers/{id}`和`GET /containers/{id}/logs`。这些入口必须先通过`agent_box_session`鉴权并从`authctx`读取当前 AgentBox 用户；`GET /health/docker`通过显式注入的`DockerHealthBackend`执行真实 Docker daemon ping，Docker client 创建失败、daemon 不可达或返回异常时包装为结构化`CodeContainerRuntimeUnavailable`且不影响插件启动。本轮新增受控 Docker lifecycle 子切片：`List/Detail/Start/Stop/Delete/Logs`只管理同时带`john-ai-agentbox.managed=true`、`john-ai-agentbox.user=<当前用户>`和`john-ai-agentbox.container_id`标签的独立容器，跨用户、未带插件标签或缺少逻辑容器 ID 的 Docker 容器统一返回 not found 语义，不泄露 Docker ID、名称、状态或标签。`Create`仍保持结构化`CodeContainerRuntimeUnavailable`，等待可信镜像、工作区、卷和运行时配置迁移；Agent runtime start/stop/logs 已通过 catalog 受控 Docker lifecycle 子集迁移到插件路径，Shell/Chat WebSocket、服务代理 relay 和 TCP tunnel 当前保持鉴权后结构化不可用。 |
| Service Proxy JSON 子集 | 已新增`backend/api/serviceproxy`、`backend/internal/service/serviceproxy`和`backend/internal/controller/serviceproxy`，覆盖`GET /agents/{id}/services`、`GET /agents/{id}/services/{serviceId}`、`GET/POST /agents/{id}/service-bridges`和`DELETE /agents/{id}/service-bridges/{bridgeId}`。这些 JSON 入口先校验当前 AgentBox 用户对目标 Agent/service proxy 的归属。本轮将`GET /services`和`GET /services/{serviceId}`推进为真实只读 Docker runtime service discovery 子集：`serviceproxy.New(accessSvc, dockerRuntimeBackend)`显式接收启动期共享 Docker backend，Docker 后端只检查带`john-ai-agentbox.managed/user/agent_id`标签且属于当前用户和 Agent 的运行中容器，通过受控 exec 读取`/proc/net/tcp`和`/proc/net/tcp6`，按端口聚合最多 100 个监听服务，返回`svc-<port>`、监听地址、端口、网络族、可访问状态、进程 ID/名称和 Unix 毫秒`lastCheckedAt`。非 loopback 或 unspecified 监听标记为`direct`，loopback 监听标记为`bridge_required`且说明需要 bridge；该只读切片不创建 bridge、不生成 proxy URL、不生成 tunnel URL、不启动 relay、不扫描未标记容器，也不读取其他用户容器。`service-bridges`创建/删除、HTTP proxy relay 和 TCP tunnel 已迁到插件路径，当前返回结构化`CodeServiceProxyRuntimeUnavailable`或`CodeGatewayRuntimeUnavailable`，不再占用旧根级代理路径。 |
| Raw Gateway 路径子集 | 已新增`backend/internal/service/gateway`和`backend/internal/controller/gateway`，将`ALL /proxy/*`、`GET /ws/agents/{id}/shell`、`GET /ws/agents/{id}/chat/sessions/{sessionId}`和`GET /ws/agents/{id}/services/{serviceId}/tcp`注册到插件命名空间`/x/john-ai-agentbox/api/v1`下。当前 HTTP proxy raw handler 只校验`agent_box_session`、解析 opaque key 与上游相对路径，然后返回结构化`CodeGatewayRuntimeUnavailable`；当前 WebSocket/tunnel raw handler 不升级 WebSocket、不转发 TCP，只复用`agent_box_session`鉴权和`access`资源归属校验；可见资源返回结构化`CodeGatewayRuntimeUnavailable`。本次迁移已完成路径收敛和资源守卫，真实 HTTP proxy、WebSocket、Shell、Chat 和 TCP tunnel 流式 relay 运行时后端明确保持不可用，不再存在旧根级生产入口。 |
| Agent runtime start/stop/logs 路由切片 | 已新增`POST /agents/{id}/start`、`POST /agents/{id}/stop`和`GET /agents/{id}/logs`插件路径，前端既有`api.startAgent`、`api.stopAgent`和`api.getAgentLogs`调用不再落到未注册路由。本轮将该切片从 ownership-check stub 推进为真实 Docker lifecycle 子集：`catalog`通过`AgentRuntimeBackend`窄接口接收启动期共享的 Docker backend，先按当前 AgentBox 用户执行`GetUserAgent(ctx, userID, agentID)`归属校验，再只使用插件数据库中的 Agent、coding image `image_ref`、image ID、Agent 类型和名称创建或启动带`john-ai-agentbox.managed/user/agent_id/image_id/name`标签的长驻容器；成功后写入`john_ai_agentbox_agent_runtimes.container_id/docker_id/status`，`stop`更新为`stopped`，`logs`只读取同一 Agent 标签容器日志。跨用户 Agent 继续保持 not-found/invisible 语义；Docker client/daemon、create/start/stop/logs 失败通过`CodeCatalogRuntimeUnavailable`结构化返回。当前 runtime 不挂载额外工作区配置卷、不执行 tmux/preflight、不启动 Chat/Shell WebSocket relay；这些流式能力已迁到插件路径并保持结构化不可用。 |
| 鉴权中间件 | 已将插件受保护路由统一接入`agent_box_session`校验中间件；`/x/john-ai-agentbox/api/v1/auth/*`保持公开用于登录、当前会话和注销，其他已迁移的 JSON API 路由都在进入 controller 前验证插件会话。本轮静态扫描`authctx`、`agent_box_session`、`Authorization`、`jwt`、`sys_user`和管理登录态相关关键词，确认插件请求鉴权只读取插件 Cookie 并写入`authctx`；命中的`Authorization`仅用于远端 AI provider 调用，不作为 AgentBox 浏览器凭据。已重新运行`PLAYWRIGHT_HTML_OPEN=never pnpm -C hack/tests test:module -- plugin:john-ai-agentbox --grep "TC001|TC002"`，7 个浏览器断言通过，覆盖插件入口、登录入口、AgentBox 登录成功、`/admin`隔离和双登录态注销边界；代表截图`temp/20260611/212955-tc001d-app-shell.png`和`212959-tc002a-admin-auth-boundary.png`显示边界正常。 |
| 结构化错误 | 已将当前迁移切片中的用户可见错误、认证失败、资源引用失败、设置读写失败、Chat session/interaction 失败、Workspace runtime 未就绪失败、Container runtime 未就绪失败、Gateway runtime 未就绪失败和 Service Proxy runtime 未就绪失败统一为`lina-core/pkg/bizerr`错误码，避免裸字符串进入 HTTP 响应。本轮静态扫描`r.SetError`、`Response.Write*`、`gerror.New`、`errors.New`、`fmt.Errorf`、`NewCode`、`WrapCode`和`MustDefine`，确认生产 controller 只通过错误返回或`SetError(err)`进入统一响应，`WriteJson/WriteStatus`仅出现在测试 helper，`gerror.New`用于构造期依赖缺失或作为被`bizerr.WrapCode`包裹的低层 cause。后续真实 Terminal、WebSocket 和 proxy/tunnel 流量后端新增用户可见失败时仍必须沿用`bizerr`，当前已暴露接口错误收敛完成。 |
| 编译门禁 | 已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/setting ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/setting ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 settings service、controller 和路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖`prompt-templates`路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/prompt ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/prompt -count=1`通过，覆盖 prompt-template controller 和 service；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/ai ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/ai ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 AI service、controller 和路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/chat ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/agent ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 Chat service、Agent controller 和路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/workspace ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/workspace ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 Workspace service、controller 和路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/serviceproxy ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/serviceproxy ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 service proxy service、controller 和路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/terminal ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/agent ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 Terminal service、Agent controller 和路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/gateway ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/gateway ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 HTTP proxy raw route、Gateway service/controller 和插件路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/container ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/container ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 Docker health backend 注入、Docker health 成功/失败投影、Container controller 和插件路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/... -count=1`通过，覆盖插件后端、controller、service 和 routes 包；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.plugins GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/auth ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/auth ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/middleware ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖独立认证服务、Cookie controller、`agent_box_session`中间件和插件路由绑定；本轮重新运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.plugins GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/catalog ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/container ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 Agent runtime backend 注入、Docker Agent runtime helper、catalog runtime helper 和插件路由绑定；本轮重新运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.plugins GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/workspace ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/container ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/workspace ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 workspace preview/download/resource runtime 流、Docker workspace open file、controller 二进制响应和路由绑定；本轮重新运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.plugins GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/access ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/auth ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/catalog ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/chat ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/container ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/gateway ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/serviceproxy ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/terminal ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/workspace ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/agent ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/container ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/workspace ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过。 |

## 4. 插件 SQL、DAO 和数据隔离

- [x] 4.1 合并外部 AgentBox SQL 为插件安装 SQL，所有表、索引、外键、唯一约束和业务键统一使用`john_ai_agentbox_`前缀。
- [x] 4.2 新增插件卸载 SQL 和可选 mock-data SQL，确保安装、卸载和 mock 数据分类符合插件 SQL 资源约定。
- [x] 4.3 确认 SQL 使用 PostgreSQL 14+ 源语法、幂等 DDL/Seed DML、`INSERT ... ON CONFLICT DO NOTHING`或等价幂等写入，不显式写入自增 ID。
- [x] 4.4 为 Agent 列表、provider/model 关联、Chat 会话、Terminal 会话、调用日志、用户会话和资源归属查询补充必要索引。
- [x] 4.5 配置插件`backend/hack/config.yaml`并执行 DAO 生成，使 DAO、DO、Entity 只生成到插件`backend/internal/dao/`和`backend/internal/model/`。
- [x] 4.6 静态检索旧表名、宿主`AI`表名、`linapro-ai-core`表名和宿主 DAO import，确认 AgentBox 生产代码无交叉存储依赖。

实现记录：

| 项目 | 记录 |
|------|------|
| 安装 SQL | 已将外部`agent-box`的 14 个增量 SQL 合并为插件当前迭代安装 SQL`apps/lina-plugins/john-ai-agentbox/manifest/sql/001-john-ai-agentbox-schema.sql`，采用最终结构而不是保留旧中间迁移。 |
| 表空间隔离 | 新增表、索引、外键引用和唯一索引均使用`john_ai_agentbox_`前缀；覆盖用户、会话、AI provider/model、coding image/agent、runtime、chat、terminal、settings、capability tiers、invocation logs 和 prompt overrides。 |
| 卸载和 mock 分类 | 已新增`manifest/sql/uninstall/001-john-ai-agentbox-schema.sql`按依赖顺序清理插件自有表；当前没有演示/测试 Mock 数据，因此`manifest/sql/mock-data/`仅保留目录，不写入 mock SQL。 |
| 幂等与数据写入 | DDL 使用`CREATE TABLE IF NOT EXISTS`、`CREATE INDEX IF NOT EXISTS`和`DROP TABLE IF EXISTS`；Seed DML 使用`ON CONFLICT DO NOTHING`。静态检索未发现`ON DUPLICATE`、`ON CONFLICT ... DO UPDATE`、`BIGSERIAL`或`CREATE DATABASE`；唯一`INSERT`中的`id`为`john_ai_agentbox_users.id`文本业务键，不是自增主键。 |
| 索引和查询路径 | 已为用户登录、会话过期清理、provider/model 关联、Agent owner+软删除、Agent updated 排序、Chat session last_active、message sequence、interaction pending/status、Terminal session owner/agent/status、AI binding 和 invocation log 过滤排序建立索引，支撑后续列表、详情、日志、会话和资源归属查询。 |
| 资源嵌入边界 | `plugin_embed.go`只嵌入`manifest/sql/*.sql`和`manifest/sql/uninstall/*.sql`等插件生命周期资源，不嵌入后端内部源码或未授权目录；SQL 不属于 public assets。 |
| DAO 生成 | 已新增`apps/lina-plugins/john-ai-agentbox/backend/hack/config.yaml`作为开发期 DAO 生成配置，使用 PostgreSQL 14 codegen 数据库和 17 张`john_ai_agentbox_`前缀表；执行`go run ./hack/tools/linactl dao dir=apps/lina-plugins/john-ai-agentbox/backend`成功，生成结果仅位于插件`backend/internal/dao/`、`backend/internal/dao/internal/`、`backend/internal/model/do/`和`backend/internal/model/entity/`。 |
| 存储依赖隔离 | 静态检索`apps/lina-plugins/john-ai-agentbox`未发现`apps/lina-core/internal/dao`、`lina-core/internal/dao`、宿主`internal/model`、`linapro-ai-core`、`linapro_ai_core`、`sys_ai`或`ai_core`依赖；生成文件未包含`sys_*`表或旧裸表名。插件生产代码只依赖插件自身 DAO internal 包路径`john-ai-agentbox/backend/internal/dao/internal`。 |
| 数据权限影响 | 本阶段只生成插件自有表访问工件，不新增业务读取或写操作；AgentBox 使用插件用户隔离的业务例外和资源归属校验继续由 5.x 任务实现与测试。 |
| 缓存一致性影响 | DAO 生成和静态扫描不新增缓存、快照、派生状态或失效逻辑；会话、runtime 和服务发现缓存边界继续由 5.x、8.x 任务记录。 |
| `i18n`影响 | `backend/hack/config.yaml`和生成 DAO/DO/Entity 不引入运行时用户可见文案或插件语言资源；`plugin.yaml`未启用`i18n.enabled: true`，仍按单语言插件治理。 |
| 开发工具影响 | 未新增长期维护脚本或工具；复用既有跨平台 Go 工具入口`linactl dao`。本地验证运行于 macOS，命令入口由 Go 实现，Windows/Linux 继续使用同一`go run ./hack/tools/linactl dao dir=...`形式。 |
| 验证 | `rg`静态检索确认插件 SQL 无旧裸表名命中；`docker run --rm ... postgres:14-alpine ... psql -f /sql/001-john-ai-agentbox-schema.sql`连续执行两次成功，Seed 第二次均`INSERT 0 0`，随后执行卸载 SQL 成功；`go run ./hack/tools/linactl dao dir=apps/lina-plugins/john-ai-agentbox/backend`成功；`find apps/lina-plugins/john-ai-agentbox/backend/internal/dao apps/lina-plugins/john-ai-agentbox/backend/internal/model -type f | sort`确认生成位置；`rg --pcre2 -n "Table: (sys_|ai_|provider_models|coding_agents|users)|\\bsys_[a-z0-9_]+\\b|CREATE TABLE (?!IF NOT EXISTS john_ai_agentbox_)" ...`无命中；`rg -n "apps/lina-core/internal/(dao|model)|lina-core/internal/(dao|model)|linapro-ai-core|linapro_ai_core|sys_ai|plugin_ai|ai_core" apps/lina-plugins/john-ai-agentbox`无命中；`openspec validate migrate-john-ai-agentbox --strict`通过；插件 Go 编译门禁继续通过。 |

## 5. 独立认证和业务数据权限

- [x] 5.1 迁移 AgentBox 用户、密码哈希、会话 token 哈希和`agent_box_session`Cookie 逻辑，保持与 LinaPro 管理登录态隔离。
- [x] 5.2 为登录、当前会话和注销接口补充单元测试，覆盖 Cookie 属性、token 哈希、过期、注销和无效凭据。
- [x] 5.3 为 Agent、Chat、Workspace、Terminal、service proxy 和文件下载等资源访问增加当前 AgentBox 用户归属校验。
- [x] 5.4 在任务记录中说明数据权限例外：AgentBox 使用插件自有用户体系，不接入 LinaPro 角色管理数据权限，但所有插件数据按当前 AgentBox 用户隔离。
- [x] 5.5 补充不可见资源拒绝测试，确认错误不泄露其他用户资源名称、状态、数量或代理路径。

实现记录：

| 项目 | 记录 |
|------|------|
| Agent 归属校验 | 已将 Agent 列表、详情、创建、更新、换图和删除收敛到`authctx.RequireUserID(ctx)`读取的当前 AgentBox 用户；`catalog` service 在数据库查询和更新条件中同时限定`user_id`，其他用户 Agent 访问返回不可见语义。 |
| 数据权限例外边界 | AgentBox 不接入 LinaPro 角色管理数据权限；当前实现按插件自有用户体系隔离`auth`、`agent`和`catalog`配置数据。Chat JSON、Chat WebSocket stub、Shell WebSocket stub、Terminal REST 元数据、Workspace `paths/tree/file/upload/download/resource/html-preview/skills/git`子集、Service Proxy discovery/bridge JSON 子集和 TCP tunnel stub 已按当前 AgentBox 用户归属校验；Container lifecycle 子切片不读取 LinaPro 数据权限，使用 Docker 插件标签`john-ai-agentbox.user=<当前用户>`作为运行时资源可见性边界。真实 Terminal/WebSocket runtime 后端仍待后续迁移。 |
| 资源访问基础服务 | 已新增`backend/internal/service/access`，集中提供 Agent、Chat session、Terminal session、workspace resource 和 service proxy 的当前 AgentBox 用户归属校验入口；缺失资源与其他用户资源统一返回不可见语义。该服务已接入 Agent Chat JSON controller/service、Gateway raw WebSocket/tunnel stub、Terminal REST 元数据 controller/service、Workspace 文件/下载/资源/skills/Git controller/service 和 Service Proxy discovery/bridge JSON service；真实 Terminal/WebSocket runtime 后端迁移时仍需沿用同一 access service，完整流式运行时由 3.3 继续跟进，不阻塞当前资源归属校验任务闭环。 |
| Agent runtime start/stop/logs 归属校验 | 已新增`/agents/{id}/start`、`/agents/{id}/stop`和`/agents/{id}/logs`当前用户归属边界：controller 先读取`authctx`用户，service 复用`GetUserAgent(ctx, userID, agentID)`的`user_id`限定查询。其他用户 Agent 在 start/stop/logs 上返回与不存在等价的`CodeCatalogNotFound`，不泄露其他用户 Agent 名称、状态、容器 ID、Docker ID 或日志内容。本轮真实 Docker runtime 子集也只按当前用户和 Agent 标签查找容器，`logs`必须同时满足`john-ai-agentbox.user=<当前用户>`和`john-ai-agentbox.agent_id=<agentID>`；Docker failure 对可见 Agent 统一包装为`CodeCatalogRuntimeUnavailable`。 |
| Terminal REST 元数据切片 | 已新增`backend/api/agent/v1/terminal.go`、`backend/internal/service/terminal`和 Agent controller 终端 handlers，覆盖`GET/POST /agents/{id}/terminal/sessions`、`GET/DELETE /agents/{id}/terminal/sessions/{terminalId}`。这些接口只管理插件表`john_ai_agentbox_agent_terminal_sessions`中的会话元数据，先校验当前 AgentBox 用户对 Agent 或 Terminal session 的归属；backend session name 由 user/agent/terminal ID 哈希生成，不暴露原始 terminal ID。真实 Shell WebSocket attach、tmux backend 和 Docker exec 仍归 3.3 后续 runtime 迁移，不在该切片伪造。 |
| 不可见资源拒绝测试 | 已补充 Agent 认证上下文和 controller 路径测试，确认未认证访问被拒绝、当前用户上下文会被正确注入；已补充`backend/internal/service/chat/chat_test.go`覆盖其他用户 Agent、其他用户 Chat session、message、interaction 和 recover 的不可见拒绝语义；已补充`backend/internal/service/workspace/workspace_test.go`和`backend/internal/controller/workspace/workspace_v1_handlers_test.go`覆盖 Workspace `paths/tree/file/download/resource/skills/git`在 runtime 前先执行归属校验，并覆盖非法 workspace root、文件名和 Git repository-relative file 拒绝；已补充`backend/internal/service/gateway/gateway_test.go`覆盖 HTTP proxy raw route 要求已认证用户，Chat WebSocket、Shell WebSocket 和 TCP tunnel stub 在 runtime 前先执行归属校验；已补充`backend/internal/service/serviceproxy/serviceproxy_test.go`和`backend/internal/controller/serviceproxy/serviceproxy_v1_handlers_test.go`覆盖 service proxy discovery/bridge 在 runtime 前先执行归属校验；已补充`backend/internal/service/terminal/terminal_session_test.go`和`backend/internal/controller/agent/agent_v1_handlers_test.go`覆盖 Terminal REST 元数据的归属校验和当前用户上下文注入；TC003 已准备插件自有 E2E seed/session/terminal，覆盖跨用户 Agent、Chat、Chat WebSocket、Shell WebSocket、Terminal REST、Workspace tree、workspace download/resource/html preview、skills、Git、TCP tunnel 和 service proxy/bridge 响应不泄露资源标识，并补充未认证访问`/proxy/*`返回 401。本轮重新运行 TC003 三个子用例通过，新增截图`temp/20260611/212659-tc003a-login-boundary-for-unauthenticated-api.png`、`212701-tc003b-cross-user-api-denied.png`和`212704-tc003c-cross-user-proxy-denied.png`，截图未发现资源标识泄漏、布局异常或原始`i18n` key。 |
| 数据权限验证 | `GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/access -count=1`通过，覆盖不可见资源不泄露资源 ID、路径、状态和代理标识的基础拒绝语义；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/terminal ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/agent ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 Terminal service、Agent controller 和插件路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/gateway ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/gateway ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 HTTP proxy raw route 和 gateway 路由绑定；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/container ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/container ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过，覆盖 container lifecycle 先取当前 AgentBox 用户、Docker label user 边界、跨用户/未标记容器 not found 和 create 继续 unavailable；已运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.plugins GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/access ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/chat ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/workspace ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/terminal ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/serviceproxy ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/gateway ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/agent ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/workspace ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/serviceproxy ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/gateway ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过；本轮新增`container_agent_runtime_test.go`覆盖 Agent runtime 标签不会混入普通 container list 语义、跨用户不可见和可信 create input 规范化；新增`catalog_agent_runtime_test.go`覆盖 runtime 状态归一化和未注入 runtime backend 时保持结构化 unavailable；本轮重新运行`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.plugins GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/access ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/auth ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/catalog ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/chat ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/container ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/gateway ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/serviceproxy ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/terminal ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/workspace ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/agent ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/container ./apps/lina-plugins/john-ai-agentbox/backend/routes -count=1`通过；本轮重新运行`PLAYWRIGHT_HTML_OPEN=never pnpm -C hack/tests test:module -- plugin:john-ai-agentbox --grep TC003`通过，确认不可见资源拒绝测试有效。`pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`、`node hack/tests/scripts/validate-e2e.mjs`此前通过，确认 TC003 新增未认证 proxy 断言和插件 E2E 组织有效。 |

实现记录：

| 项目 | 记录 |
|------|------|
| 独立认证实现 | 已将占位认证替换为插件自有认证服务：`Login`按用户名读取`john_ai_agentbox_users`、使用`bcrypt.CompareHashAndPassword`校验密码、拒绝非`active`用户、生成 32 字节 URL-safe opaque token、仅将`SHA-256` token hash 写入`john_ai_agentbox_user_sessions`；`CurrentSession`只接受未撤销、未过期且用户仍为`active`的插件会话；`Logout`只撤销插件会话。 |
| Cookie 隔离 | Controller 继续只读写`agent_box_session`，设置`HttpOnly`、`SameSite=Lax`和`Path=/`；未读取 LinaPro 管理工作台 token、JWT、`sys_user`、`sys_online_session`或管理 Cookie，注销只清理`agent_box_session`。 |
| 存储边界 | 生产实现使用`auth.NewDAOStore()`，只依赖插件生成的`john-ai-agentbox/backend/internal/dao`、`do`和`entity`，不依赖宿主 DAO/DO/Entity 或`linapro-ai-core`；写入使用 GoFrame DO 对象，登录查询、会话查询和撤销均以数据库条件限定在插件表。 |
| DI 来源检查 | `routes.Register`在源码插件 HTTP 注册阶段创建`auth.NewDAOStore()`和`auth.New(..., Config{})`，并显式传入`authcontroller.NewV1(authSvc)`；请求处理路径不再临时创建关键 service 图。`auth`服务持有的会话状态权威数据源是插件数据库，不持有进程内缓存。 |
| 结构化错误 | 已新增`JOHN_AI_AGENTBOX_AUTH_INVALID_CREDENTIALS`、`JOHN_AI_AGENTBOX_AUTH_REQUIRED`、`JOHN_AI_AGENTBOX_AUTH_USER_DISABLED`和`JOHN_AI_AGENTBOX_AUTH_STORE_UNAVAILABLE`，登录失败、未认证、禁用用户和存储异常均通过`bizerr.NewCode`或`bizerr.WrapCode`返回。 |
| 数据权限影响 | 5.1/5.2 仅完成认证用户/session 边界；Agent、Chat、Workspace、Terminal、service proxy 和文件下载等资源归属校验仍由 5.3/5.5 继续实现。认证会话本身按`token_hash`和`user_id`隔离，不接入 LinaPro 角色管理数据权限。 |
| 缓存一致性影响 | 登录会话不引入进程内缓存；权威状态为`john_ai_agentbox_user_sessions`和`john_ai_agentbox_users`，撤销后当前数据库可见。集群下所有节点通过同一数据库读取会话状态，暂不引入本地 token cache。 |
| `i18n`影响 | 插件仍未启用`i18n.enabled: true`，新增 API 文档源文本和错误 fallback 均按单语言插件治理，不向`apps/lina-core/manifest/i18n`写入插件翻译资源。 |
| 测试覆盖 | 新增`backend/internal/service/auth/auth_session_test.go`覆盖登录成功、token hash 存储、请求元数据、当前会话、注销、无效密码、缺失用户、禁用用户、过期 session 和空 token 注销；新增`backend/internal/controller/auth/auth_v1_test.go`覆盖登录写入`agent_box_session`、`HttpOnly`、`SameSite=Lax`、当前会话读取 Cookie、注销清理 Cookie 和登录错误不写 Cookie。 |
| 验证 | `go mod tidy`生成插件`go.sum`；`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/backend/internal/service/auth ./apps/lina-plugins/john-ai-agentbox/backend/internal/controller/auth -count=1`通过；`GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/... -count=1`通过；静态扫描未发现宿主认证存储或宿主 DAO 依赖。 |

## 6. 前端门户迁移

- [x] 6.1 将外部 AgentBox React/Vite 前端迁移到插件`frontend/`，保留实际产品界面作为首屏，不改造成管理工作台页面。
- [x] 6.2 将 API base、WebSocket base、public asset base 和登录路径集中配置为插件路径，替换硬编码生产`/api`、`/ws`、`/assets`和根级代理路径。
- [x] 6.3 确保浏览器访问和刷新`/`时显示 AgentBox 功能界面，访问和刷新`/login`时显示 AgentBox 登录页面。
- [x] 6.4 确认`/admin`、管理工作台 router base 和 LinaPro 前端资源不受 AgentBox 门户影响。
- [x] 6.5 运行 AgentBox 插件前端构建、类型检查和必要静态检索，确认无根级 API/WebSocket 路径残留。

实现记录：

| 项目 | 记录 |
|------|------|
| 前端迁移 | 已将外部`agent-box/apps/frontend`迁入`apps/lina-plugins/john-ai-agentbox/frontend`，保留 AgentBox React/Vite 产品工作台界面作为插件门户首屏，没有改造成 LinaPro 管理工作台菜单页面。 |
| 路径集中配置 | 新增`frontend/src/plugin-paths.ts`集中定义`/x/john-ai-agentbox/api/v1`、`/x/john-ai-agentbox/api/v1/auth/session`、`/x/john-ai-agentbox/api/v1/auth/sessions`和 WebSocket base；`api.ts`、`ChatPanel.tsx`和`ShellPanel.tsx`均改为使用插件路径配置。 |
| public asset 边界 | `vite.config.ts`配置`base: /x-assets/john-ai-agentbox/0.1.0/`、`publicDir: false`、`build.outDir: dist`，`plugin_embed.go`嵌入`frontend/dist/**/*`，插件清单仍只发布`frontend/dist`声明资源。 |
| 门户路由行为 | 插件后端精确`GET /`入口直接返回 AgentBox SPA `index.html`并显示`portal-auth-gate`，浏览器地址保持为`/`；精确`GET /login`入口直接返回同一 SPA `index.html`并显示独立`login-page`，浏览器地址保持为`/login`；登录成功后前端恢复地址为`/`并显示`agentbox-app-shell`。 |
| 管理工作台隔离 | `/admin`继续由宿主工作台 base path 处理；插件只注册精确`/`和`/login`入口，不注册通配 fallback，因此不吞掉宿主 API、插件 API、插件资产或管理工作台路径。插件前端不注册管理工作台 router 或菜单。 |
| 用户可见行为测试资产 | 新增`agentbox-app-shell`和`agentbox-logout-button`稳定定位点，支撑插件专属 E2E 验证登录成功后的 AgentBox 功能界面和插件自身注销动作。 |
| `i18n`影响 | 插件`plugin.yaml`未启用`i18n.enabled: true`，迁移前端保留单语言运行时文案；未向`apps/lina-core/manifest/i18n`写入插件文案或 API 文档翻译资源。 |
| 验证 | `pnpm -C apps/lina-plugins/john-ai-agentbox/frontend exec tsc --noEmit --skipLibCheck`通过；`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend run build`通过并完成资产预算、gzip、brotli 检查；静态检索`frontend/src`、`vite.config.ts`和`frontend/dist/index.html`确认只剩`/x-assets/john-ai-agentbox/0.1.0/...`资产路径，无根级生产`/api`、`/ws`、`/assets`或`/s/`残留。 |

## 7. E2E 和回归验证

- [x] 7.1 创建`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/TC001-portal-login.ts`，实现 TC-1a：未认证访问`/`进入 AgentBox 认证流程；TC-1b：访问`/login`显示独立登录页；TC-1c：登录成功后`/`显示功能界面。
- [x] 7.2 创建`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/TC002-admin-isolation.ts`，实现 TC-2a：AgentBox 登录不认证`/admin`；TC-2b：管理工作台登录不认证 AgentBox；TC-2c：两个注销动作互不清理对方登录态。
- [x] 7.3 创建`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/TC003-agentbox-api-data-isolation.ts`，实现 TC-3a：未带`agent_box_session`访问插件受保护 API 被拒绝；TC-3b：当前用户不可访问其他用户 Agent；TC-3c：不可访问其他用户 Chat、workspace 或服务代理资源。
- [x] 7.4 为插件专属 E2E 创建或迁移 POM 到`apps/lina-plugins/john-ai-agentbox/hack/tests/pages/`，helper 放到同插件`hack/tests/support/`。
- [x] 7.5 执行插件专属 E2E，并按测试规则保存关键页面、登录、失败路径和核心工作流截图到`temp/<YYYYMMDD>/`。

实现记录：

| 项目 | 记录 |
|------|------|
| TC001 门户登录 | 已创建`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/TC001-portal-login.ts`，覆盖匿名入口显示`portal-auth-gate`、`/login`入口显示独立登录页、门户登录入口切换到插件`/login`真实路径，以及使用插件初始化账号`admin/admin123`登录后显示`agentbox-app-shell`。 |
| TC002 管理工作台隔离 | 已创建`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/TC002-admin-isolation.ts`，覆盖 AgentBox 登录后访问`/admin`仍进入管理工作台登录、管理工作台登录后访问`/`仍按 AgentBox 会话边界显示门户认证，以及 AgentBox 注销和管理工作台注销互不替代。 |
| TC003 数据隔离状态 | 已创建`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/TC003-agentbox-api-data-isolation.ts`并覆盖 TC-3a：未带`agent_box_session`访问插件受保护`/agents`API 和`/proxy/*`raw proxy route 均返回 LinaPro `200` HTTP 响应包裹的非零业务错误码`JOHN_AI_AGENTBOX_AUTH_REQUIRED`；已通过插件专属`hack/tests/support/isolation.ts`准备最小用户、session、Agent、Chat 和 Terminal seed，覆盖 TC-3b：当前用户访问其他用户 Agent、Agent runtime start、Agent runtime stop 和 Agent runtime logs 返回非零业务错误且不泄露 Agent ID/名称；覆盖 TC-3c：当前用户访问其他用户 Chat messages、Chat WebSocket stub、Shell WebSocket stub、Terminal REST、Workspace tree/download/resource/html preview、Skills、Git、TCP tunnel、Service detail 和 Service bridge creation 返回非零业务错误且不泄露 Agent ID、Chat ID、Terminal ID、标题、backend session name、workspace 路径、service ID 或 session token。 |
| POM 和 helper | 已新增插件专属`hack/tests/pages/AgentBoxPortalPage.ts`、`hack/tests/support/plugin.ts`和`hack/tests/support/isolation.ts`，页面对象、路径 helper 和测试数据 helper 均位于`apps/lina-plugins/john-ai-agentbox/hack/tests/`，没有回流到宿主测试目录。 |
| 浏览器 E2E | 本轮启动本地 PostgreSQL 14 容器、从`config.template.yaml`创建被 Git 忽略的本地`apps/lina-core/manifest/config/config.yaml`，执行`make db.init confirm=init`初始化数据库，并通过`make dev skip_wasm=1`在插件 full mode 启动后端`http://127.0.0.1:9120`和管理前端`http://127.0.0.1:5666`。已运行`PLAYWRIGHT_HTML_OPEN=never pnpm -C hack/tests test:module -- plugin:john-ai-agentbox`，插件专属 E2E `TC001`、`TC002`和`TC003`共 10 个子测试全部通过。 |
| 截图验证 | 已保存通过轮次截图到`temp/20260611/`，覆盖`204709-tc001a-root-auth-gate.png`、`204712-tc001b-login-page.png`、`204715-tc001c-login-from-portal.png`、`204718-tc001d-app-shell.png`、`204723-tc002a-admin-auth-boundary.png`、`204727-tc002b-agentbox-boundary-after-admin-login.png`、`204732-tc002c-scoped-logout-boundary.png`、`204736-tc003a-login-boundary-for-unauthenticated-api.png`、`204738-tc003b-cross-user-api-denied.png`和`204741-tc003c-cross-user-proxy-denied.png`。已人工审查代表截图：AgentBox 门户、登录页、应用壳、LinaPro 管理登录边界和 API 失败路径页面无明显布局重叠、截断或原始 i18n key。TC003b/TC003c 主要通过 API payload 断言证明跨用户资源拒绝，截图仅记录执行时浏览器状态。 |
| 验证 | `pnpm -C hack/tests install --frozen-lockfile`安装宿主测试依赖；本轮已重新运行`pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`通过，确认插件专属 E2E、POM 和 helper 可以被宿主 Playwright 配置类型检查；本轮已重新运行`node hack/tests/scripts/validate-e2e.mjs`通过，确认插件 E2E 文件组织和治理规则有效。首次浏览器执行暴露测试自身未绕过宿主工作台`page.goto('/')`重写、且错误断言误判 LinaPro HTTP envelope 的问题；已修复为使用插件 public base URL 和非零业务错误码断言后重跑通过。 |

## 8. 治理、缓存和工具验证

- [x] 8.1 记录`i18n`影响判断：首期若`plugin.yaml`未启用`i18n.enabled: true`，说明该插件按单语言治理且不向`lina-core`写入插件翻译资源。
- [x] 8.2 记录缓存一致性影响：说明 AgentBox 会话、runtime、服务发现、AI 档位解析或配置缓存的权威数据源、失效策略、单机/集群边界和降级行为；若某项未引入缓存则明确无影响。
- [x] 8.3 记录 DI 来源检查：说明新增运行期依赖的 owner、创建位置、传递路径、共享实例策略，以及没有在请求路径临时`New()`关键服务图。
- [x] 8.4 记录开发工具跨平台影响：若新增构建、打包、代码生成或测试入口，说明跨 Windows、Linux、macOS 的执行方式和验证结果；若未新增工具则记录无影响。
- [x] 8.5 运行`openspec validate migrate-john-ai-agentbox --strict`，并修复所有提案、设计、任务或规格问题。
- [x] 8.6 执行最终静态检索：旧根级路径、旧表名、旧 module import、宿主`AI`数据依赖和插件 ID 不一致均无残留。
- [x] 8.7 完成实现后调用`lina-review`进行代码和规范审查，确认任务仅在实现和验证完成后标记为`[x]`。

实现记录：

| 项目 | 记录 |
|------|------|
| `i18n`影响 | `apps/lina-plugins/john-ai-agentbox/plugin.yaml`未配置`i18n.enabled: true`，首期按单语言插件治理；未创建`manifest/i18n`资源，未向`apps/lina-core/manifest/i18n`写入插件运行时文案或`apidoc`翻译。若后续启用插件`i18n`，必须只维护插件自己的`manifest/i18n/<locale>/`和`manifest/i18n/<locale>/apidoc/`。 |
| 缓存一致性 | 当前已实现的认证会话权威数据源为`john_ai_agentbox_users`和`john_ai_agentbox_user_sessions`，不引入进程内 token cache；撤销、过期和用户禁用均按数据库当前状态读取，集群节点共享同一数据库即可一致。当前 Agent catalog 查询不新增进程内缓存，provider 远端模型同步直接写入插件表。Chat JSON、Terminal REST 元数据、Gateway raw HTTP proxy/WebSocket/tunnel stub 和 Service Proxy bridge JSON 子集不新增进程内缓存；Container `GET /health/docker`每次调用直接 ping Docker daemon，不缓存健康状态，daemon 不可用时返回结构化不可用；Container lifecycle 子切片每次直接通过 Docker daemon 按`john-ai-agentbox.*`标签读取或执行动作，不引入进程内容器列表、状态或日志缓存，Docker daemon 是运行时状态权威源。本轮 Agent runtime lifecycle 子集以`john_ai_agentbox_agent_runtimes`作为 Agent 到 Docker 容器映射的权威数据源，Docker daemon 是容器实际运行状态权威源；不引入进程内 runtime cache，start/stop 成功后同步写数据库，logs 每次直接从 Docker 读取。Workspace runtime 子集同样不引入进程内 tree/file/Git/skill/cache 或 HTML preview cache，`paths/tree/file preview/upload/download/resource/html preview/save/create file/create directory/git status/git file/git diff/skills`每次通过 Docker daemon exec/stat/copy 写读当前容器文件系统或 Git 当前状态，权威状态是运行中容器内文件系统、技能目录和仓库元数据；上传、下载、资源端点、HTML 预览、Git file/diff 和技能列表按请求即时读取处理，不把文件树、目录状态、Git 状态、diff、技能清单、HTML 预览或文件内容缓存到进程内；保存时的`baseHash`只对当前请求做即时读取比较，不维护服务端缓存。Service discovery 子集不引入进程内服务列表缓存，每次通过 Docker daemon 和容器内`/proc/net/tcp*`即时读取当前监听端口，Docker daemon 与容器网络命名空间是运行时权威状态；bridge/proxy/tunnel 状态仍未引入。runtime 缺失或 Docker 不可用时结构化降级为 unavailable。集群模式下当前实现要求调用节点能访问同一 Docker daemon/容器上下文；跨节点调度、会话粘滞和 runtime lease 尚未实现，真实 WebSocket、AI 档位解析和真实 proxy/tunnel 的运行时缓存/状态仍需后续继续声明权威源、失效策略、单机/集群边界和降级行为。 |
| 本轮治理影响 | 本轮将 Workspace `html-previews`推进为真实 Docker runtime 文件流预览子集，并将 Service Proxy `services`推进为真实只读 Docker runtime 监听端口发现子集，同时修正 API 文档源文本；插件仍未启用`i18n.enabled: true`，新增/修改 API 文档源文本和错误 fallback 按单语言插件治理，不写宿主或插件`manifest/i18n`。该切片不新增开发工具、脚本、SQL、DAO 生成或长期维护工具；新增运行期能力继续复用同一个启动期 Docker backend，通过`serviceproxy.RuntimeBackend`窄契约注入`serviceproxy.New(accessSvc, dockerRuntimeBackend)`，不创建 bridge、proxy relay、tunnel、WebSocket 或进程内缓存。 |
| DI 来源检查 | 源码插件入口`plugin_embed.go`注册启动期 HTTP callback；`routes.Register`在插件路由装配阶段创建`auth.NewDAOStore()`、`auth.New(...)`、`http.Client{Timeout: 20s}`、`containersvc.NewDockerRuntimeBackend()`、`catalog.New(..., catalog.Config{AgentRuntimeBackend: dockerRuntimeBackend})`、`ai.New(...)`、`setting.New()`、`prompt.NewDAOStore()`、`prompt.New(...)`、`access.New(access.NewDAOStore())`、`chat.New(accessSvc)`、`terminal.New(accessSvc)`、`workspace.New(accessSvc, dockerRuntimeBackend)`、`serviceproxy.New(accessSvc, dockerRuntimeBackend)`、`gateway.New(accessSvc)`、`container.New(dockerRuntimeBackend, dockerRuntimeBackend)`和各 controller，并通过构造函数显式传入。Docker runtime backend 依赖 owner 为插件`container` service、`catalog` runtime lifecycle、`workspace`只读/基础写入/runtime preview 和`serviceproxy`只读 service discovery 共同复用，创建位置为源码插件路由装配阶段；该 backend 暴露`Ping/List/Inspect/Start/Stop/Delete/Logs`、`CreateAgentRuntime/StartAgentRuntime/StopAgentRuntime/AgentRuntimeLogs`、`WorkspacePathStat/WorkspaceDirectoryEntries/WorkspaceReadFile`、`WorkspaceOpenFile`、`WorkspaceWriteFile`、`WorkspaceUploadFile`、`WorkspaceCreateEntry`、`WorkspaceGitStatus`、`WorkspaceGitFile`、`WorkspaceGitDiff`、`WorkspaceSkills`和`RuntimeServices`窄契约，Docker client 创建失败保存在 backend 内并在请求时包装为结构化 runtime unavailable，不会导致插件启动失败或泄漏 Docker client。`catalog`与`ai`复用同一个启动期`http.Client`实例作为远端 provider 请求依赖；`agentcontroller.NewV1(catalogSvc, chatSvc, terminalSvc)`显式接收 catalog、Chat 与 Terminal service；`workspacecontroller.NewV1(workspaceSvc)`显式接收 Workspace service；`serviceproxycontroller.NewV1(serviceProxySvc)`显式接收 Service Proxy service；`gatewaycontroller.New(gatewaySvc)`显式接收 Gateway service；`containercontroller.NewV1(containerSvc)`显式接收 Container service；请求处理路径没有临时`New()`认证、catalog、AI、setting、prompt、access、chat、terminal、workspace、service proxy、gateway、Docker runtime backend 或 container 关键服务图。真实 Terminal WebSocket、skills upload、Git write/AI commit runtime 和 proxy/tunnel 后续入口迁移时仍需沿用同一 access service 或启动期共享依赖。 |
| 开发工具跨平台影响 | 未新增长期维护脚本或平台专属默认入口；DAO 生成复用既有 Go 工具`go run ./hack/tools/linactl dao dir=apps/lina-plugins/john-ai-agentbox/backend`；前端复用 Node/Pnpm/Vite 跨平台工具链；E2E 复用宿主`hack/tests` Playwright 配置。验证在 macOS 执行，命令入口本身不依赖 shell-only 语义，Windows/Linux 可使用同一 pnpm/go 命令。 |
| 最终静态检索 | 已运行`rg -n -e '/api/' -e '/api/v1' -e '/ws/' -e '/s/' -e '/assets/' apps/lina-plugins/john-ai-agentbox/frontend/src apps/lina-plugins/john-ai-agentbox/backend apps/lina-plugins/john-ai-agentbox/plugin.yaml`，命中均为插件 API 前缀`/x/john-ai-agentbox/api/v1`、插件相对路由注册`/ws/...`、测试断言或`/x-assets/john-ai-agentbox/...`构建产物，不存在生产根级`/api`、根级`/ws`、根级`/s`或根级`/assets`残留。已运行`rg -n -e 'agent-box' -e 'agent_box' -e 'module agent-box' -e 'github.com/.*/agent-box' -e 'agentbox-web' apps/lina-plugins/john-ai-agentbox --glob '!frontend/pnpm-lock.yaml' --glob '!frontend/dist/**' --glob '!frontend/.vite/**'`，剩余命中均为预期`agent_box_session`Cookie 或`john-ai-agentbox-web`包名。已运行`rg -n -e 'apps/lina-core/internal/(dao|model)' -e 'lina-core/internal/(dao|model)' -e 'linapro-ai-core' -e 'linapro_ai_core' -e 'sys_ai' -e 'plugin_ai' -e 'ai_core' apps/lina-plugins/john-ai-agentbox`无命中。已运行`find apps/lina-plugins/john-ai-agentbox/backend/internal/dao apps/lina-plugins/john-ai-agentbox/backend/internal/model -type f -name 'john_ai_agentbox_*.go'`无输出。裸表名扫描仅在 GoFrame 生成的`entity`注释中出现去前缀展示，例如`table users`；实际 DAO/DO 元数据和 SQL 均使用`john_ai_agentbox_`前缀。 |
| 当前验证 | 本轮重新运行`env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`，工作目录为`apps/lina-plugins/john-ai-agentbox`，通过；重新运行`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend exec tsc --noEmit --skipLibCheck`通过；静态检查确认`apps/lina-core`无本次迁移 diff。 |
| OpenSpec 验证 | 已运行`openspec validate migrate-john-ai-agentbox --strict`并通过；本轮任务记录更新后将再次运行严格校验。 |

### Lina 审查报告

**变更：** `migrate-john-ai-agentbox`
**范围：** 全部变更，范围来源为`git status --short`、`git ls-files --others --exclude-standard`和 OpenSpec 活跃变更上下文；当前候选文件数为`339`。`apps/lina-plugins/john-ai-agentbox/AGENTS.md`不存在，插件目录继续遵守顶层`AGENTS.md`和命中规则。
**已读取规则文件：** `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`和`.agents/instructions/markdown-format.instructions.md`。

#### 发现的问题

- 未发现阻塞问题。

#### 规则域结论

- OpenSpec：通过。活跃变更判定正确，`8.7`在审查和验证完成后标记，已执行严格校验。
- 插件和架构：通过。AgentBox 资源归属`apps/lina-plugins/john-ai-agentbox/`，本轮不修改`apps/lina-core`主框架代码或插件清单协议；页面入口闭环在插件内部精确路由。
- 后端 Go 和 API：通过。插件 API、WebSocket stub、workspace、container、service proxy 和 gateway 路径收敛到`/x/john-ai-agentbox/api/v1/...`；用户可见错误按`bizerr`结构化；插件页面入口只注册`GET /`和`GET /login`。
- 缓存一致性：通过。本轮未新增宿主或插件页面入口缓存；AgentBox 插件业务路径未引入额外进程内业务数据缓存。
- 数据权限：通过。AgentBox 按插件自有用户和资源归属隔离，作为不接入 LinaPro 角色数据权限的业务例外已记录；不可见资源测试覆盖拒绝语义。
- 数据库：通过。插件 SQL 使用`john_ai_agentbox_`表名前缀、幂等 DDL/Seed、卸载 SQL和索引设计；DAO/DO/Entity 保留插件目录归属，生成类型已去掉冗余前缀。
- 前端/UI 和 E2E：通过。插件前端路径集中配置，保留 AgentBox 产品首屏；`TC001`、`TC002`、`TC003`位于插件 E2E 目录，POM/helper 未回流宿主。当前轮次未重跑浏览器截图，沿用`temp/20260611/`已通过截图证据和文本/接口断言。
- `i18n`：通过。`plugin.yaml`未启用`i18n.enabled: true`，按单语言插件治理，未向宿主语言包写入插件翻译。
- 开发工具：通过。未新增长期维护平台专属脚本；使用既有 Go、pnpm、Playwright 入口。`.gitmodules`删除和`apps/lina-plugins`普通目录化属于本次插件治理范围，未提交或推送。

#### 验证证据

- `env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`，工作目录`apps/lina-plugins/john-ai-agentbox`：通过。
- `pnpm -C apps/lina-plugins/john-ai-agentbox/frontend exec tsc --noEmit --skipLibCheck`：通过。

#### 摘要

- 严重：`0`
- 警告：`0`
- 剩余风险：真实 WebSocket、HTTP proxy relay、TCP tunnel、跨节点 runtime 调度和会话粘滞仍是后续能力范围；本次迁移已将入口迁到插件路径并以结构化 unavailable 降级。

## Feedback

- [x] **FB-1**: DAO/DO/Entity 生成未去掉插件表前缀，导致生成文件名和 Go 类型冗余包含`JohnAiAgentbox`
- [x] **FB-2**: AgentBox 页面入口迁移不应修改`lina-core`主框架，应闭环到插件内部
- [x] **FB-3**: `john-ai-agentbox`插件`manifest/config`缺少业务配置项，启动装配仍使用插件后端硬编码默认值
- [x] **FB-4**: 首次进入 AgentBox 时缺失工作台设置被当作加载失败提示
- [x] **FB-5**: 后端`/`根路由应直接作为 AgentBox 工作台入口，而不是跳转到`/x-assets/john-ai-agentbox/0.1.0/index.html`
- [x] **FB-6**: 首次登录初始化工作台设置时后端 upsert 使用错误冲突键导致初始化失败提示
- [x] **FB-7**: 插件 SQL 缺少表字段注释，且插件根目录缺少可执行`make dao`的 Makefile 入口
- [x] **FB-8**: 前端构建产物应输出到`frontend/dist`，而不是`frontend/public`
- [x] **FB-9**: `frontend/dist`不应进入 Git 版本库，应在编译前生成并由 Go `embed`在编译时嵌入
- [x] **FB-10**: 根目录`make build`应默认编译主框架、默认工作台和所有插件，并支持`dir=`指定单个目录构建
- [x] **FB-11**: 插件构建逻辑应由插件自己的`Makefile`维护，不应在共享`plugin.codegen.mk`中固定通用`build`指令

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | 插件`backend/hack/config.yaml`声明了 17 张`john_ai_agentbox_`表用于 GoFrame DAO 生成，但未配置`removePrefix`，导致生成文件名和 DAO/DO/Entity Go 类型冗余包含`JohnAiAgentbox`。 |
| 修复 | 在插件 DAO 生成配置中增加`removePrefix: "john_ai_agentbox_"`，保留 SQL 表名、索引、外键和存储隔离前缀不变；重新执行`go run ./hack/tools/linactl dao dir=apps/lina-plugins/john-ai-agentbox/backend`生成 DAO/DO/Entity，并删除生成器遗留的旧`john_ai_agentbox_*.go`生成文件。 |
| 代码引用 | 将插件业务代码和测试中的生成工件引用同步改为去前缀名称，例如`dao.Users`、`do.Users`、`entity.Users`、`dao.ProviderModels`和`dao.SystemPromptOverrides`。 |
| 静态验证 | `find apps/lina-plugins/john-ai-agentbox/backend/internal/dao apps/lina-plugins/john-ai-agentbox/backend/internal/model -type f -name 'john_ai_agentbox_*.go'`无输出；`rg -n "JohnAiAgentbox" apps/lina-plugins/john-ai-agentbox/backend/internal/dao apps/lina-plugins/john-ai-agentbox/backend/internal/model apps/lina-plugins/john-ai-agentbox/backend/internal/service apps/lina-plugins/john-ai-agentbox/backend/internal/controller apps/lina-plugins/john-ai-agentbox/backend/routes -S`无输出；`rg -n "dao\\.JohnAiAgentbox|do\\.JohnAiAgentbox|entity\\.JohnAiAgentbox" apps/lina-plugins/john-ai-agentbox/backend -S`无输出。 |
| Go 编译门禁 | `GOWORK=/Users/john/Workspace/github/gqcn/agentbox/temp/go.work.agentbox GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./apps/lina-plugins/john-ai-agentbox/... -count=1`通过，覆盖插件 API、controller、service、routes、DAO/DO/Entity 生成结果使用包。 |
| OpenSpec 验证 | `openspec validate migrate-john-ai-agentbox --strict`通过。 |
| `i18n`影响 | 本反馈只调整开发期 DAO 生成配置、生成源码命名和内部 Go 引用，不新增运行时 UI、菜单、按钮、API 文档源文本、错误 fallback、插件清单或语言包资源；插件`plugin.yaml`仍未启用`i18n.enabled: true`。 |
| 缓存一致性影响 | 无缓存、快照、派生状态、失效或跨实例同步变更；数据库表名和运行时权威数据源不变。 |
| 数据权限影响 | 不新增或改变业务读写接口、资源可见性、租户/用户边界或数据权限策略；只是内部生成类型命名变化。 |
| 开发工具跨平台影响 | 未新增或修改长期维护工具、脚本或入口；复用既有跨平台 Go 工具`linactl dao`。本地 codegen 数据库通过临时 PostgreSQL 容器提供，生成完成后已停止清理。 |
| 测试策略 | 该反馈属于生成治理和内部编译兼容修复，不改变用户可观察行为，因此不新增 E2E；使用 DAO 重生成、静态扫描、Go 编译门禁和 OpenSpec 严格校验闭环。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`plugin.md`、`backend-go.md`、`database.md`、`testing.md`、`i18n.md`、`data-permission.md`、`cache-consistency.md`、`dev-tooling.md`、`documentation.md`和`api-contract.md`；其中 API 契约无本反馈新增路由或 DTO 变更影响。 |

## 修复 FB-2: 页面入口闭环到插件内部

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | 原实现为了让 AgentBox 拥有`/`和`/login`，在`apps/lina-core`中新增了`portal_routes`清单字段、宿主 SPA fallback、门户路由索引缓存、宿主路由测试和插件 README 说明。该需求只来源于`john-ai-agentbox`页面入口，属于插件展示适配，不应扩大`lina-core`核心宿主协议和缓存边界。 |
| 修复 | 已恢复`apps/lina-core`全部本次门户相关改动，并删除新增的宿主门户 fallback 文件；`apps/lina-plugins/john-ai-agentbox/plugin.yaml`移除`portal_routes`。新增插件内部`backend/routes/portal.go`，通过现有源码插件`http.route.register`精确注册`GET /`和`GET /login`，两个入口均直接返回插件嵌入的 AgentBox SPA `index.html`，不依赖宿主 fallback。 |
| 路由边界 | 插件不注册`/*`或其他通配 fallback；`/api/**`、`/x/**`、`/x-assets/**`、`/admin/**`和`/api.json`继续由宿主既有路由处理。 |
| 前端调整 | `App.tsx`按浏览器真实路径`/login`区分登录页；登录成功后恢复地址为`/`并显示 AgentBox 应用壳。E2E POM 保持访问`/login`入口，由插件精确后端路由直接返回 SPA 并渲染登录页。 |
| OpenSpec 调整 | 已删除`workspace-route-boundary`增量规范，`proposal.md`、`design.md`、`john-ai-agentbox-plugin`增量规范和任务记录均改为插件内部入口方案；本变更不再声明或要求主框架能力变更。 |
| `i18n`影响 | 插件仍未启用`i18n.enabled: true`；本反馈不新增宿主或插件语言包资源，不新增 API 文档源文本或错误 fallback。 |
| 缓存一致性影响 | 无缓存、快照、派生状态、失效或跨实例同步变更；页面入口在插件启动注册时读取嵌入的`frontend/dist/index.html`作为不可变发布资产，静态资源仍由既有`/x-assets`按插件版本和`public_assets`声明提供。 |
| 数据权限影响 | 页面入口只暴露公开入口和登录页，不读取或装配业务数据；业务 API 和资源访问仍由插件自有`agent_box_session`和资源归属校验控制。 |
| 开发工具跨平台影响 | 未新增或修改长期维护工具、脚本或入口；继续复用现有源码插件路由、Vite 构建和 Playwright 测试入口。 |
| 测试策略 | 该反馈改变用户可观察入口路径实现，更新插件路由单元测试和插件专属 E2E/POM；运行插件 Go 编译门禁、前端类型检查、OpenSpec 严格校验和静态扫描闭环。 |
| 验证结果 | `env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./backend/routes -count=1`通过；`env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`通过；`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend exec tsc --noEmit --skipLibCheck`通过；`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend run build`通过；`pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`通过；`node hack/tests/scripts/validate-e2e.mjs`通过；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --name-status -- apps/lina-core`无输出。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`data-permission.md`、`cache-consistency.md`、`dev-tooling.md`和`documentation.md`；插件目录无本地`AGENTS.md`。 |

## 修复 FB-3: 插件 manifest/config 补齐业务配置项

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | `apps/lina-plugins/john-ai-agentbox/manifest/config/config.yaml`和`config.example.yaml`只包含`runtime.mode: single-node`骨架；同时插件后端启动装配仍在`routes.Register`中硬编码`authsvc.Config{}`、`http.Client{Timeout: 20s}`、`containersvc.NewDockerRuntimeBackend()`和`aisvc.Config{RequestTimeout: 20s}`，Docker runtime、workspace、service discovery 的路径和容量限制也分散在插件 service 常量或 runtime 文件内，导致插件业务配置没有在`manifest/config`闭环。 |
| 修复 | 在插件`manifest/config/config.yaml`和`config.example.yaml`补齐`auth.sessionTtl`、`providers.requestTimeout`、`providers.remoteModelSyncLimit`、`ai.requestTimeout`、`runtime.mode`、`runtime.docker.host/containerLogTail/stopTimeout`、`runtime.workspace.rootPath/sharedPath/previewLimitBytes/uploadFileLimitBytes/uploadCountLimit/skillListLimit/skillManifestLimitBytes`和`runtime.services.discoveryLimit`。新增插件内部`backend/routes/config.go`，通过`registrar.Services().Plugins().Config()`读取插件作用域配置并在 host config 缺失的 focused unit test 中使用插件内部默认值。 |
| 启动装配 | `routes.Register`改为在源码插件路由注册阶段读取`agentBoxConfig`，并把纯值配置显式传入`auth.New`、共享`http.Client`、`containersvc.NewDockerRuntimeBackend(config.Docker)`、`catalog.New`、`ai.New`、`workspace.New`和`serviceproxy.New`。Docker backend 继续由插件启动期创建一次，并被 catalog、workspace、serviceproxy 和 container 复用。 |
| Runtime 配置消费 | Docker client 支持通过`runtime.docker.host`指定 host，空值继续使用 Docker 环境默认；容器日志 tail、stop timeout、workspace root/shared root、workspace 预览/上传/技能上限和 service discovery 上限均由插件配置进入 runtime backend 或 service。controller 层不再硬编码 upload count 上限，由 workspace service 统一按配置校验。 |
| 主框架边界 | 本反馈不修改`apps/lina-core`代码、主框架配置契约、插件清单协议、宿主 service 语义或宿主 README；复用已有源码插件配置能力`registrar.Services().Plugins().Config()`，因此无需向主框架新增能力。 |
| DI 来源检查 | 新增的是纯值配置结构，不新增关键运行期 service；配置读取发生在插件启动装配阶段，传递路径为`registrar.Services().Plugins().Config()`到`loadAgentBoxConfig`，再进入插件内部 service/runtime 构造函数。请求处理路径未新增临时`New()`关键服务图。 |
| `i18n`影响 | 插件`plugin.yaml`未启用`i18n.enabled: true`，本反馈仅补充 YAML 配置项、启动期内部错误和单元测试，不新增运行时 UI、菜单、按钮、前端文案、API 文档源文本、插件清单文案或语言包资源；不需要维护宿主或插件`manifest/i18n`。 |
| 缓存一致性影响 | 配置由宿主既有插件作用域配置服务按启动期读取；本反馈不新增进程内业务缓存、权限快照、插件状态快照或跨实例失效逻辑。配置变更运行期热更新不在本次范围内，当前语义为插件启动装配时生效。 |
| 数据权限影响 | 不新增或改变业务数据读写接口、下载、导出、聚合、批量操作、资源可见性或租户/用户边界；配置项只影响插件内部运行时连接、超时和有界容量。 |
| 数据库和 SQL 影响 | 不新增或修改 SQL、DAO、DO、Entity、软删除、时间字段或索引；无数据库迁移和 DAO 生成影响。 |
| 开发工具跨平台影响 | 不新增或修改长期维护工具、脚本、CI、Makefile、linactl 或平台专属入口；验证使用既有 Go/OpenSpec/git 命令。 |
| 测试策略 | 该反馈涉及插件启动装配和 runtime 配置消费，新增`backend/routes/config_test.go`覆盖默认值、业务配置读取和不支持 runtime mode 拒绝；运行插件 Go 编译门禁、OpenSpec 严格校验、静态扫描和`apps/lina-core`无 diff 检查闭环。 |
| 验证结果 | `env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`在`apps/lina-plugins/john-ai-agentbox`通过；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过；`git diff --name-status -- apps/lina-core`无输出；静态扫描确认插件启动装配不再保留`authsvc.Config{}`、固定`http.Client{Timeout: 20 * time.Second}`、无参`NewDockerRuntimeBackend()`或旧 runtime 常量引用。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`architecture.md`、`plugin.md`、`backend-go.md`、`api-contract.md`、`testing.md`、`i18n.md`、`data-permission.md`、`cache-consistency.md`、`database.md`、`dev-tooling.md`和`documentation.md`；插件目录无本地`AGENTS.md`。 |

## 修复 FB-4: 首登缺失工作台设置不应提示加载失败

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | `GET /x/john-ai-agentbox/api/v1/settings/workbench`在当前用户尚未保存工作台设置时按后端契约返回结构化`JOHN_AI_AGENTBOX_SETTING_NOT_FOUND`，用于让客户端初始化默认值；但前端`ApiError`只保存 HTTP status 和 GoFrame 数字 code，且`hydrateWorkbenchSettings`只用`status === 404`识别缺失设置。LinaPro 统一响应 envelope 下业务错误 HTTP status 可能仍为`200`，导致首次进入 AgentBox 时把可预期的“设置未初始化”当成加载失败 toast。 |
| 修复 | 扩展插件前端`ApiResponse`和`ApiError`以保留统一响应中的`errorCode`、`messageKey`和`messageParams`；`hydrateWorkbenchSettings`改为优先通过稳定业务错误码`JOHN_AI_AGENTBOX_SETTING_NOT_FOUND`识别缺失设置并静默使用本地或默认工作台设置，其他真实错误仍显示“加载工作台设置失败”。 |
| 前端影响 | 修改手写源码范围限于`apps/lina-plugins/john-ai-agentbox/frontend/src/api.ts`、`types.ts`和`App.tsx`，并通过前端构建重新生成`frontend/dist`静态资产。不改变后端设置 API、数据库表、用户设置存储语义或插件路径；只修正前端错误分类和首次初始化体验。 |
| E2E 覆盖 | 更新`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/TC001-portal-login.ts`的`TC001d`：测试先删除`usr-admin`的`workbench`用户设置，登录后断言`GET settings/workbench`返回`JOHN_AI_AGENTBOX_SETTING_NOT_FOUND`，并确认 AgentBox 应用壳可见且不会出现“加载工作台设置失败”toast。截图保存到`temp/20260612/144510-tc001d-app-shell.png`，未发现错误 toast、原始`i18n` key 或布局异常。 |
| API 契约影响 | 不新增或修改 HTTP 路由、方法、后端 DTO、`g.Meta`、权限标签或时间字段；前端类型补齐已有 envelope 元数据字段，保持后端契约不变。 |
| 数据权限影响 | 不新增或改变业务数据读写边界；工作台设置仍按插件自有`user_id + key`隔离。E2E 只清理本地测试库中`usr-admin/workbench`测试行以复现首登缺失状态。 |
| 数据库和 SQL 影响 | 不新增或修改 SQL、DAO、DO、Entity、索引、软删除或时间字段；无数据库迁移和 DAO 生成影响。 |
| 缓存一致性影响 | 不新增后端缓存、快照、派生状态、失效或跨实例同步逻辑；前端仍只使用既有`localStorage`作为当前浏览器默认设置兜底。 |
| `i18n`影响 | 插件`plugin.yaml`未启用`i18n.enabled: true`，本反馈不新增运行时文案、API 文档源文本、错误 fallback、插件清单文案或语言包资源；只是避免错误地展示既有中文 toast。 |
| 开发工具跨平台影响 | 不新增或修改长期维护工具、脚本、CI、Makefile、`linactl`或平台专属入口；验证使用既有`pnpm`、Playwright、OpenSpec 和 Git 检查。 |
| 测试策略 | 该反馈属于用户可观察行为修复，使用插件专属 E2E 覆盖复现场景；同时运行前端类型检查、前端构建、E2E 类型检查、E2E 组织校验、OpenSpec 严格校验和格式检查。 |
| 验证结果 | `pnpm -C apps/lina-plugins/john-ai-agentbox/frontend exec tsc --noEmit --skipLibCheck`通过；`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend run build`通过；`pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`通过；`node hack/tests/scripts/validate-e2e.mjs`通过；`PLAYWRIGHT_HTML_OPEN=never pnpm -C hack/tests test:module -- plugin:john-ai-agentbox --grep TC001`通过，`4 passed`；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`api-contract.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`data-permission.md`、`database.md`、`backend-go.md`、`cache-consistency.md`、`dev-tooling.md`和`goframe-v2`技能；插件目录无本地`AGENTS.md`。 |

## 修复 FB-5: 根路由直接作为 AgentBox 工作台入口

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | `apps/lina-plugins/john-ai-agentbox/backend/routes/portal.go`中的`servePortalIndex`和`redirectToLoginRoute`使用`Response.RedirectTo`返回`302`，导致浏览器地址栏从`/`或`/login`落到`/x-assets/john-ai-agentbox/0.1.0/index.html`或其 hash 登录路由。该实现把`/x-assets`静态资源托管路径暴露成用户入口，不符合“后端`/`根路由就是 AgentBox 工作台”的入口语义。 |
| 修复 | `plugin_embed.go`将插件嵌入资源`embeddedFiles`显式传入`routes.Register`；`routes.Register`在启动装配阶段调用`registerPortalRoutes`读取`frontend/dist/index.html`并构造固定 handler；`GET /`和`GET /login`现在直接返回`text/html; charset=utf-8`的 AgentBox SPA `index.html`，不再返回跳转响应。静态 JS/CSS/font 仍继续通过既有`/x-assets/john-ai-agentbox/0.1.0/...`加载。 |
| 前端调整 | `App.tsx`改为通过浏览器真实路径`/login`识别登录页，并保留对旧`#/login`的兼容识别；点击门户登录按钮后地址栏变为`/login`，登录成功后恢复为`/`。`TC001`和插件 POM 同步改为断言`/login`真实路径，而不是`#/login`。 |
| OpenSpec 调整 | `john-ai-agentbox-plugin`增量规范和`design.md`已改为要求`GET /`直接返回 AgentBox 工作台 SPA `index.html`并保持浏览器地址为`/`，`GET /login`直接返回同一 SPA 并保持地址为`/login`。 |
| 路由边界 | 插件仍只注册精确`GET /`和`GET /login`，不注册`/*`或其他通配 fallback；`/api/**`、`/x/**`、`/x-assets/**`、`/admin/**`和`/api.json`继续由宿主既有路由处理。 |
| DI 来源检查 | 新增依赖是插件自身嵌入资源`fs.FS`，owner 为`john-ai-agentbox`源码插件入口，创建位置为`plugin_embed.go`的`//go:embed`资源；传递路径为`plugin_embed.go`到`routes.Register`再到`registerPortalRoutes`。未新增关键运行期 service、缓存敏感服务或请求路径临时`New()`关键服务图。 |
| `i18n`影响 | 本反馈不新增或修改运行时用户可见文案、API 文档源文本、错误 fallback、插件清单文案或语言包资源；插件`plugin.yaml`仍未启用`i18n.enabled: true`，继续按单语言插件治理。 |
| 缓存一致性影响 | 不新增业务缓存、权限快照、插件状态快照、失效或跨实例同步逻辑；入口 handler 持有的是发布期嵌入的不可变`index.html`字节，随插件二进制发布更新。 |
| 数据权限影响 | 页面入口只返回公开 SPA HTML，不读取、写入或装配业务数据；业务 API 和资源访问仍由`agent_box_session`、插件用户归属和各 service 数据边界控制。 |
| 数据库和 SQL 影响 | 不新增或修改 SQL、DAO、DO、Entity、索引、软删除或时间字段；无数据库迁移和 DAO 生成影响。 |
| 开发工具跨平台影响 | 不新增或修改长期维护工具、脚本、CI、Makefile、`linactl`或平台专属入口；验证继续使用既有 Go、pnpm、Playwright 和 OpenSpec 命令。 |
| 测试策略 | 该反馈属于用户可观察入口行为修复，新增插件路由单测覆盖`/`和`/login`返回`200 text/html`且无`Location`跳转；更新插件专属 E2E 覆盖根入口、登录入口、门户登录链接和登录后回到工作台。 |
| 截图验证 | `PLAYWRIGHT_HTML_OPEN=never pnpm -C hack/tests test:module -- plugin:john-ai-agentbox --grep TC001`生成并人工检查`temp/20260612/150537-tc001a-root-auth-gate.png`、`150540-tc001b-login-page.png`、`150544-tc001c-login-from-portal.png`和`150551-tc001d-app-shell.png`，未发现错误跳转提示、原始`i18n` key、布局重叠或登录后停留在`/login`。 |
| 验证结果 | `env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./backend/routes -count=1`通过；`env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`在`apps/lina-plugins/john-ai-agentbox`通过；`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend exec tsc --noEmit --skipLibCheck`通过；`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend run build`通过；`pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`通过；`node hack/tests/scripts/validate-e2e.mjs`通过；`PLAYWRIGHT_HTML_OPEN=never pnpm -C hack/tests test:module -- plugin:john-ai-agentbox --grep TC001`通过，`4 passed`；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`documentation.md`、`architecture.md`、`plugin.md`、`api-contract.md`、`backend-go.md`、`frontend-ui.md`、`testing.md`和`goframe-v2`技能；`data-permission`、`database`、`cache-consistency`、`dev-tooling`和`i18n`已完成无影响评估，未触发对应规则文件强制读取场景。插件目录无本地`AGENTS.md`。 |

## 修复 FB-6: 初始化工作台设置 upsert 冲突键错误

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | 浏览器截图和后端日志显示首次登录后出现红色 toast：`初始化工作台设置失败：AgentBox setting storage is temporarily unavailable`。进一步排查发现`john_ai_agentbox_user_settings`真实主键为`(user_id, key)`，但`UpsertUserSetting`未显式声明冲突目标，GoFrame `Save()`生成了`ON CONFLICT (user_id)`，PostgreSQL 因`user_id`不是唯一约束而拒绝写入，导致前端默认工作台设置初始化失败。 |
| 修复 | 在`apps/lina-plugins/john-ai-agentbox/backend/internal/service/setting/setting_user.go`中为用户设置写入显式指定`OnConflict(dao.UserSettings.Columns().UserId, dao.UserSettings.Columns().Key).OnDuplicate(dao.UserSettings.Columns().Value).Save()`，与表主键`(user_id, key)`保持一致；继续使用`do.UserSettings`，不手写`created_at`或`updated_at`，保留 GoFrame 自动时间维护。 |
| E2E 覆盖 | 增强`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/TC001-portal-login.ts`的`TC001d`：测试先删除`usr-admin/workbench`设置，登录后断言`GET /settings/workbench`返回稳定业务错误码`JOHN_AI_AGENTBOX_SETTING_NOT_FOUND`，并等待默认设置初始化路径执行后确认页面不会出现`(加载|初始化|同步)工作台设置失败`toast。`TC001a`和`TC001b`同时增加`/`、`/login`响应级断言，确认入口直接返回插件 SPA `index.html`且没有跳转。 |
| 截图验证 | 本轮重新运行插件全量 E2E 并人工复核`temp/20260612/153438-tc001a-root-auth-gate.png`、`153440-tc001b-login-page.png`、`153447-tc001d-app-shell.png`、`153451-tc002a-admin-auth-boundary.png`、`153455-tc002b-agentbox-boundary-after-admin-login.png`、`153501-tc002c-scoped-logout-boundary.png`、`153503-tc003a-login-boundary-for-unauthenticated-api.png`、`153506-tc003b-cross-user-api-denied.png`和`153508-tc003c-cross-user-proxy-denied.png`。最新应用壳截图未出现旧的红色初始化失败 toast、原始`i18n` key、明显重叠或文本截断。 |
| API 契约影响 | 不新增或修改 HTTP 路由、HTTP 方法、DTO、`g.Meta`、权限标签或时间字段；只修复既有`PUT /settings/{key}`后端写入实现，使其符合已有用户设置表主键语义。 |
| 数据权限影响 | 不新增或改变业务资源可见性、租户边界或 LinaPro 角色数据权限接入；工作台设置继续按 AgentBox 自有`user_id + key`隔离，显式冲突键反而强化了同一用户不同 key 与不同用户同 key 的存储边界正确性。 |
| 数据库和 SQL 影响 | 不新增或修改 SQL、DAO、DO、Entity、索引、软删除或时间字段；本修复复用现有`john_ai_agentbox_user_settings`主键`(user_id, key)`，无数据库迁移和 DAO 生成影响。 |
| 缓存一致性影响 | 不新增后端缓存、快照、派生状态、失效或跨实例同步逻辑；设置权威数据源仍为`john_ai_agentbox_user_settings`，前端`localStorage`只作为当前浏览器默认设置兜底。 |
| `i18n`影响 | 插件`plugin.yaml`未启用`i18n.enabled: true`，本反馈不新增运行时文案、API 文档源文本、错误 fallback、插件清单文案或语言包资源；E2E 只扩大既有中文错误 toast 的负向断言。 |
| 开发工具跨平台影响 | 不新增或修改长期维护工具、脚本、CI、Makefile、`linactl`或平台专属入口；验证继续使用既有 Go、pnpm、Playwright、OpenSpec 和 Git 检查。 |
| DI 来源检查 | 不新增运行期依赖、构造函数参数、启动装配、插件 host service 或缓存敏感服务；修复位于既有`setting`service 方法内部，请求路径没有临时`New()`关键服务图。 |
| 测试策略 | 该反馈同时涉及后端写入行为和用户可观察 toast，使用插件后端包测试与插件专属 E2E 覆盖复现场景，并追加全量插件 E2E、前端构建、E2E 治理校验和截图复核。 |
| 验证结果 | `env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./backend/internal/service/setting ./backend/routes -count=1`通过；`env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`在`apps/lina-plugins/john-ai-agentbox`通过；`pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`通过；`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend exec tsc --noEmit --skipLibCheck`通过；`pnpm -C apps/lina-plugins/john-ai-agentbox/frontend run build`通过并通过资产预算、gzip 和 brotli 检查；`node hack/tests/scripts/validate-e2e.mjs`通过；`PLAYWRIGHT_HTML_OPEN=never pnpm -C hack/tests test:module -- plugin:john-ai-agentbox --grep TC001`通过，`4 passed`；`PLAYWRIGHT_HTML_OPEN=never pnpm -C hack/tests test:module -- plugin:john-ai-agentbox`通过，`10 passed`；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`documentation.md`、`architecture.md`、`data-permission.md`、`plugin.md`、`api-contract.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`dev-tooling.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`goframe-v2`技能、`lina-feedback`技能、`lina-e2e`技能、`playwright-cli`技能和`browser`技能；插件目录无本地`AGENTS.md`。 |

## 修复 FB-7: 插件 SQL 字段注释和 Makefile 生成入口

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | 插件安装 SQL 已创建 17 张`john_ai_agentbox_`表，但没有为表和字段维护 PostgreSQL `COMMENT ON`注释，导致 GoFrame `descriptionTag: true`重新生成 DAO/DO/Entity 时无法写入字段说明，生成结果中存在空`description`和空字段注释。同时插件根目录缺少参考`linapro-monitor-server`的`Makefile`，开发者不能在插件目录内直接执行`make dao`复用仓库统一生成入口。 |
| 修复 | 在`apps/lina-plugins/john-ai-agentbox/manifest/sql/001-john-ai-agentbox-schema.sql`为全部插件表和字段补充`COMMENT ON TABLE`与`COMMENT ON COLUMN`。新增插件根`Makefile`，与`linapro-monitor-server`保持同构，计算`PLUGIN_ROOT`和`REPO_ROOT`后 include 仓库统一`hack/makefiles/plugin.codegen.mk`，提供`make dao`和`make ctrl`包装入口。 |
| DAO 生成 | 启动临时 PostgreSQL 14 codegen 容器，执行插件安装 SQL 初始化`agentbox_plugin_codegen`库后，在`apps/lina-plugins/john-ai-agentbox`目录执行`make dao`通过。生成结果更新到插件自己的`backend/internal/dao/`和`backend/internal/model/{do,entity}/`，没有手工修改生成文件；临时容器生成后已停止并删除。 |
| SQL 验证 | 插件 SQL 在 PostgreSQL 14 上执行通过；重复执行同一 SQL 通过，`CREATE TABLE IF NOT EXISTS`、`CREATE INDEX IF NOT EXISTS`、`INSERT ... ON CONFLICT DO NOTHING`和`COMMENT ON`语句保持可重入。数据库元数据检查确认`john_ai_agentbox_%`业务表缺失表注释数为`0`、缺失字段注释数为`0`。 |
| 生成结果验证 | `rg -n 'description:""' apps/lina-plugins/john-ai-agentbox/backend/internal/model/entity apps/lina-plugins/john-ai-agentbox/backend/internal/model/do apps/lina-plugins/john-ai-agentbox/backend/internal/dao/internal -S`无输出；抽查`users`生成文件确认字段说明进入`description`标签、DO 字段注释和 DAO columns 注释。 |
| 数据库和 SQL 影响 | 本反馈修改的是仍处于活跃迁移变更中的插件首个安装 SQL，按用户明确要求补齐旧迭代 SQL 字段注释；不新增表、字段、索引、约束或 Seed 语义，不改变软删除、自动时间维护、自增主键写入或查询性能设计。SQL 数据分类仍为安装 DDL 与 Seed DML，Mock 数据未混入安装 SQL。 |
| 开发工具跨平台影响 | 新增插件根`Makefile`仅作为兼容包装层，业务逻辑复用跨平台 Go 工具`linactl`和共享`plugin.codegen.mk`；没有新增 Shell、PowerShell、Node 脚本或平台专属命令。已在 macOS 从插件目录执行`make dao`验证入口可用；跨平台等价入口为从仓库根目录执行`go run ./hack/tools/linactl dao dir=apps/lina-plugins/john-ai-agentbox/backend`。 |
| Go 编译门禁 | `env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`在`apps/lina-plugins/john-ai-agentbox`通过，覆盖插件 API、controller、service、routes 和重生成 DAO/DO/Entity 的使用包。 |
| API 契约影响 | 不新增或修改 HTTP 路由、HTTP 方法、DTO、`g.Meta`、权限标签、响应字段或前端调用契约；只是 SQL 注释、生成代码注释和开发期生成入口补齐。 |
| 数据权限影响 | 不新增或改变列表、详情、下载、写操作、聚合、批量操作、插件 host service 访问路径或用户资源可见性；所有业务表、外键、索引和用户归属边界保持不变。 |
| 缓存一致性影响 | 不新增或修改缓存、快照、派生状态、失效、刷新、预热、跨实例同步或运行时配置缓存；权威数据源和运行时状态边界保持不变。 |
| `i18n`影响 | 插件`plugin.yaml`未启用`i18n.enabled: true`，本反馈新增的是数据库注释和生成源码元数据，不新增运行时 UI、菜单、按钮、API 文档源文本、错误 fallback、插件清单文案或语言包资源；无需维护宿主或插件`manifest/i18n`。 |
| 测试策略 | 该反馈属于 SQL/代码生成治理和内部编译兼容修复，不改变用户可观察行为，因此不新增 E2E；使用 SQL 执行与幂等验证、DAO 重生成、生成结果静态扫描、插件 Go 编译门禁、OpenSpec 严格校验和格式检查闭环。 |
| 验证结果 | `docker exec -i agentbox-plugin-codegen-postgres psql -U postgres -d agentbox_plugin_codegen -v ON_ERROR_STOP=1 < apps/lina-plugins/john-ai-agentbox/manifest/sql/001-john-ai-agentbox-schema.sql`首次执行通过；同一命令重复执行通过；`make dao`在插件目录通过；表/字段注释缺失元数据查询结果为`tables_missing=0`和`columns_missing=0`；`rg -n 'description:""' ...`无输出；`env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`通过；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`documentation.md`、`architecture.md`、`data-permission.md`、`plugin.md`、`api-contract.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`dev-tooling.md`、`testing.md`、`i18n.md`、`goframe-v2`技能、`lina-feedback`技能和`lina-review`技能；插件目录无本地`AGENTS.md`。 |

## 修复 FB-8: 前端构建产物输出到`frontend/dist`

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | `index.html.br`和`index.html.gz`是前端构建后由`check-assets --compress`生成的 Brotli/Gzip 预压缩静态资源，属于发布构建产物；但 Vite `build.outDir`、资产压缩脚本、`plugin.yaml public_assets`、`plugin_embed.go`、门户入口读取路径和 TC004 静态资源夹具都沿用了`frontend/public`，导致构建产物落在源码公共目录。 |
| 修复 | 将前端构建输出目录、资产压缩检查目录、插件`public_assets.source`、源码插件`go:embed`、门户`index.html`读取路径和 TC004 静态资源夹具统一改为`frontend/dist`；删除旧`frontend/public/index.html`、`index.html.gz`和`index.html.br`。后续`FB-9`修正为不把`frontend/dist`纳入版本库，而是在编译前生成。 |
| 构建产物 | `pnpm -C apps/lina-plugins/john-ai-agentbox/frontend run build`已重新生成`frontend/dist/index.html`、`frontend/dist/index.html.gz`、`frontend/dist/index.html.br`和`frontend/dist/assets/**`中的 JS/CSS/font 及其预压缩文件；`frontend/public`目录已移除。 |
| DI 来源检查 | 不新增运行期 service、构造函数参数、缓存敏感服务或请求路径依赖；仍由`plugin_embed.go`嵌入插件自有`embed.FS`并传给`routes.Register`，只是发布资产路径从`frontend/public/index.html`改为`frontend/dist/index.html`。 |
| API 契约影响 | 不新增或修改 HTTP API、路由、HTTP 方法、DTO、`g.Meta`、权限标签、响应字段或前端接口调用契约；仅调整插件静态资产发布目录和构建产物路径。 |
| 数据权限影响 | 不新增或改变列表、详情、下载、写操作、聚合、批量操作、插件 host service 访问路径或用户资源可见性；页面入口只返回发布期静态 SPA HTML，不读取业务数据。 |
| 数据库和 SQL 影响 | 不新增或修改 SQL、DAO、DO、Entity、索引、软删除或时间字段；无数据库迁移和 DAO 生成影响。 |
| 缓存一致性影响 | 不新增或修改缓存、快照、派生状态、失效、刷新、预热或跨实例同步；发布资产仍是插件版本内不可变静态资源。 |
| `i18n`影响 | 插件`plugin.yaml`未启用`i18n.enabled: true`，本反馈不新增运行时 UI 文案、菜单、按钮、API 文档源文本、错误 fallback、插件清单文案或语言包资源；无需维护宿主或插件`manifest/i18n`。 |
| 开发工具跨平台影响 | 修改 Node/Vite 构建输出和 Node 资产检查脚本目标目录，继续使用既有跨平台`pnpm`、TypeScript、Vite 和 Node 入口；未新增 Shell、PowerShell、CI、Makefile 或平台专属脚本。 |
| 测试策略 | 该反馈属于构建产物目录和插件资源治理修复，不改变运行时业务语义，因此不新增浏览器 E2E；更新现有 TC004 静态资源夹具路径，并使用前端构建、插件 Go 编译门禁、E2E 类型检查、E2E 组织校验、OpenSpec 严格校验、格式检查和静态扫描闭环。 |
| 验证结果 | `pnpm -C apps/lina-plugins/john-ai-agentbox/frontend run build`通过并生成`frontend/dist`预压缩产物；`env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`在插件目录通过；`pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`通过；`node hack/tests/scripts/validate-e2e.mjs`通过；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过；静态扫描确认旧`frontend/public`路径只剩本反馈标题文本；文件检查确认`frontend/public`不存在且`frontend/dist/index.html.gz`和`frontend/dist/index.html.br`存在。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`documentation.md`、`architecture.md`、`data-permission.md`、`plugin.md`、`api-contract.md`、`backend-go.md`、`database.md`、`cache-consistency.md`、`dev-tooling.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`goframe-v2`技能、`lina-feedback`技能、`lina-e2e`技能、`karpathy-guidelines`技能和`lina-review`技能；插件目录无本地`AGENTS.md`。 |

### Lina 审查报告（FB-8）

**变更：** `migrate-john-ai-agentbox`
**范围：** 反馈级审查，覆盖`FB-8`触达的构建目录、插件资源声明、源码插件嵌入、门户静态入口、README、OpenSpec 记录、E2E/POM 类型边界、旧`frontend/public`删除和未跟踪`frontend/dist/**`构建产物；范围来源为`git status --short`、`git ls-files --others --exclude-standard`和`openspec status --change migrate-john-ai-agentbox --json`。`apps/lina-plugins/john-ai-agentbox/AGENTS.md`不存在。
**已读取规则文件：** `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/plugin.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/database.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/frontend-ui.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`和`.agents/instructions/markdown-format.instructions.md`。

#### 发现的问题

- 未发现阻塞问题。

#### 规则域结论

- OpenSpec：通过。活跃变更仍位于`openspec/changes/migrate-john-ai-agentbox`，`FB-8`仅在实现和验证完成后标记为`[x]`。
- 插件和架构：通过。静态资源发布目录统一为插件自有`frontend/dist`，未修改`apps/lina-core`宿主能力或工作台边界。
- 后端 Go 和 API：通过。`plugin_embed.go`和`portal.go`只调整嵌入资产路径，不新增运行期依赖、路由、DTO 或 API 契约；Go 编译门禁覆盖插件全包。
- 开发工具：通过。修改既有 Node/Vite 构建输出和资产检查目录，未新增平台专属脚本；`pnpm`构建通过。
- 前端/UI 和 E2E：通过。未改变用户可观察业务语义；TC004 静态夹具路径改为`frontend/dist`，POM/fixture 类型边界通过 TypeScript 和 E2E 组织校验。
- 文档：通过。`README.md`、`README.zh-CN.md`和 OpenSpec 记录同步使用`frontend/dist`。
- `i18n`：无影响。插件未启用`i18n.enabled: true`，未新增运行时文案、API 文档源文本或语言包资源。
- 缓存一致性：无影响。发布资产仍是插件版本内不可变静态文件，不新增缓存、失效或跨实例同步。
- 数据权限：无影响。仅调整静态资源构建和发布路径，不读取或暴露业务数据。
- 数据库：无影响。无 SQL、DAO、DO、Entity、索引、软删除或时间字段变更。

#### 验证证据

- `pnpm -C apps/lina-plugins/john-ai-agentbox/frontend run build`：通过，生成`frontend/dist/index.html`、`.gz`、`.br`和`dist/assets/**`。
- `env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`，工作目录`apps/lina-plugins/john-ai-agentbox`：通过。
- `pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`：通过。
- `node hack/tests/scripts/validate-e2e.mjs`：通过。
- `openspec validate migrate-john-ai-agentbox --strict`：通过。
- `git diff --check`：通过。
- 旧路径静态扫描：代码和配置无`frontend/public`发布入口残留；仅反馈记录保留问题说明文本。

#### 摘要

- 严重：`0`
- 警告：`0`
- 剩余风险：本次未重跑浏览器 E2E；该反馈是构建产物目录治理，已通过构建、类型检查、Go 编译、组织校验、静态扫描和 OpenSpec 校验覆盖。

## 修复 FB-9: `frontend/dist`编译前生成并保持 Git 忽略

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | `.gitignore`曾为`apps/lina-plugins/john-ai-agentbox/frontend/dist`和`dist/assets`添加反向例外，把 Go `//go:embed`的编译期输入误当成需要提交的版本库产物。这样会让构建产物进入 Git，且干净检出时不能明确说明`frontend/dist`应由构建流程在 Go 编译前生成。 |
| 修复 | 移除`.gitignore`中针对该插件`frontend/dist`和`dist/assets`的反向例外，让仓库已有`**/dist/`继续生效；插件根`Makefile`维护插件自己的`build`目标和`PLUGIN_BUILD_STEP_1`，该步骤转调插件前端构建；共享`hack/makefiles/plugin.codegen.mk`只保留`ctrl`、`dao`等通用代码生成目标，不再定义通用`build`。 |
| 框架构建顺序 | `linactl build`在插件启用模式下、宿主 Go `go build`前扫描`apps/lina-plugins/*/Makefile`中的`PLUGIN_BUILD_STEP_*`声明，并从插件根目录直接执行这些步骤。框架工具不再扫描插件内部`frontend/package.json`，也不假设插件必须使用`pnpm run build`；插件自行决定`build`步骤如何生成所需产物。 |
| Go 嵌入边界 | `plugin_embed.go`仍声明`//go:embed frontend/dist/**/* frontend/dist/*`，但`frontend/dist`不进入版本库；`linactl build`会在 Go 编译前通过插件级 build hook 生成该目录，随后 Go 编译按当前工作树中的生成文件完成嵌入。 |
| 开发者入口 | 插件目录内可执行`make build`，该入口直接执行插件`Makefile`中维护的`PLUGIN_BUILD_STEP_1`。跨平台默认构建路径不直接依赖系统`make`，`linactl build`读取同一个`PLUGIN_BUILD_STEP_*`声明后直接执行；无构建内容的源码插件若没有声明构建步骤，`linactl build dir=<plugin>`会输出无构建钩子并成功返回。 |
| DI 来源检查 | 不新增运行期 service、构造函数参数、插件 host service、缓存敏感服务或请求路径依赖；所有变更位于构建工具、构建入口、Git 忽略和文档记录。 |
| API 契约影响 | 不新增或修改 HTTP API、路由、HTTP 方法、DTO、`g.Meta`、权限标签、响应字段或前端调用契约。 |
| 数据权限影响 | 不新增或改变列表、详情、下载、写操作、聚合、批量操作、插件 host service 数据访问路径、用户资源可见性或租户边界；构建产物生成不读取业务数据。 |
| 数据库和 SQL 影响 | 不新增或修改 SQL、DAO、DO、Entity、索引、软删除、时间字段或 DAO 生成输入。 |
| 缓存一致性影响 | 不新增或修改缓存、快照、派生状态、失效、刷新、预热、跨实例同步或运行时配置缓存；发布资产仍是编译期嵌入的不可变静态资源。 |
| `i18n`影响 | 插件`plugin.yaml`未启用`i18n.enabled: true`；本反馈只修改构建入口、README 和 OpenSpec 记录，不新增运行时 UI 文案、菜单、按钮、API 文档源文本、错误 fallback、插件清单文案或语言包资源。 |
| 开发工具跨平台影响 | 修改`linactl build`和插件根`Makefile`。长期维护默认构建调度仍在 Go 工具`linactl`中，插件构建命令由插件`Makefile`维护，根构建读取`PLUGIN_BUILD_STEP_*`后直接执行，不依赖`make -C`或系统`make`。新增单元测试覆盖插件构建钩子在后端 Go 编译前执行，以及没有插件构建步骤时不会误触发。 |
| 文档影响 | `README.md`和`README.zh-CN.md`同步补充`make build`入口，并说明`frontend/dist`是构建生成产物、继续由 Git 忽略。OpenSpec 反馈记录同步修正`FB-8`中关于 Git 跟踪的过时描述。 |
| 测试策略 | 该反馈属于构建治理和编译顺序修复，不改变用户可观察业务行为，因此不新增浏览器 E2E；使用`linactl`单元测试、插件根构建、插件 Go 编译门禁、E2E 类型检查、E2E 组织校验、OpenSpec 严格校验、Git 忽略检查和格式检查闭环。 |
| 验证结果 | `go test ./hack/tools/linactl -run 'TestRunBuildRunsPluginBuildHookBeforeBackendCompile|TestDiscoverPluginBuildHookRootsSkipsPluginsWithoutBuildScript|TestOfficialPluginBuildEnvSeparatesHostOnlyAndPluginFullModes|TestResolveOfficialPluginBuildModeAutoDetectsWorkspace|TestEnsurePackedPublicPlaceholderCreatesGitkeep' -count=1`通过；`make -C apps/lina-plugins/john-ai-agentbox -n build`输出插件本地`pnpm --dir ".../frontend" run build`命令，确认插件目录入口不再是空目标；`make -C apps/lina-plugins/john-ai-agentbox build`通过，确认插件根`Makefile build`能生成`frontend/dist`和预压缩资产；`go run ./hack/tools/linactl build dir=apps/lina-plugins/john-ai-agentbox`通过，确认根构建调度读取插件`Makefile`步骤；`env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`在插件目录通过；`pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`通过；`node hack/tests/scripts/validate-e2e.mjs`通过；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过；`git check-ignore -v apps/lina-plugins/john-ai-agentbox/frontend/dist/index.html apps/lina-plugins/john-ai-agentbox/frontend/dist/assets/index-Dlev1m0S.js apps/lina-plugins/john-ai-agentbox/frontend/dist/assets/index-DIsS3CrO.css`确认`frontend/dist`由仓库既有`**/dist/`规则忽略；`git ls-files apps/lina-plugins/john-ai-agentbox/frontend/dist`无输出，确认无`frontend/dist`产物进入索引；`git ls-files --others --exclude-standard | rg '^apps/lina-plugins/john-ai-agentbox/frontend/dist' || true`无输出，确认构建产物不会出现在未跟踪候选中。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`documentation.md`、`architecture.md`、`data-permission.md`、`plugin.md`、`cache-consistency.md`、`dev-tooling.md`、`testing.md`、`i18n.md`和`.agents/instructions/markdown-format.instructions.md`；插件目录无本地`AGENTS.md`。 |

### Lina 审查报告（FB-9）

**变更：** `migrate-john-ai-agentbox`
**范围：** 反馈级审查，覆盖`.gitignore`、`hack/tools/linactl/command_build.go`、`hack/tools/linactl/main_test.go`、`hack/makefiles/plugin.codegen.mk`、`apps/lina-plugins/john-ai-agentbox/Makefile`、`README.md`、`README.zh-CN.md`和本任务记录；范围来源为`git status --short`、`git ls-files --others --exclude-standard`、`git ls-files apps/lina-plugins/john-ai-agentbox/frontend/dist`和`openspec status --change migrate-john-ai-agentbox --json`。`apps/lina-plugins/john-ai-agentbox/AGENTS.md`不存在。
**已读取规则文件：** `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/plugin.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/backend-go.md`和`.agents/instructions/markdown-format.instructions.md`。

#### 发现的问题

- 未发现阻塞问题。

#### 规则域结论

- OpenSpec：通过。活跃变更仍位于`openspec/changes/migrate-john-ai-agentbox`，验证和审查证据已记录后再标记`FB-9`完成。
- 插件和架构：通过。插件资产仍由插件自有`frontend/dist`生成并由`plugin_embed.go`编译期嵌入；未扩大`apps/lina-core`宿主协议或工作台边界。
- 开发工具：通过。长期维护默认构建顺序在 Go 工具`linactl build`中实现；插件 build hook 识别插件根`Makefile`中的`PLUGIN_BUILD_STEP_*`声明，不扫描或理解插件内部`frontend/package.json`。
- 后端 Go 编译门禁：通过。`linactl`变更通过目标单元测试覆盖，插件全包`go test`验证`frontend/dist`生成后当前工作区可被 Go `embed`编译。
- 文档：通过。英文和中文`README`同步说明`make build`入口、构建前生成和 Git 忽略语义；OpenSpec 记录同步修正`FB-8`中的过时跟踪描述。
- 测试策略：通过。本反馈为构建治理和编译顺序修复，无用户可观察业务行为变化，不新增浏览器 E2E；已使用工具单测、构建、Go 编译、E2E 治理校验、OpenSpec 校验、Git 忽略检查和格式检查闭环。
- `i18n`：无影响。插件未启用`i18n.enabled: true`，未新增运行时 UI 文案、API 文档源文本、错误 fallback、插件清单文案或语言包资源。
- 缓存一致性：无影响。只调整编译前静态资产生成和 Git 忽略策略，不新增缓存、快照、失效或跨实例同步。
- 数据权限：无影响。构建产物生成不读取业务数据，不改变读写接口、资源可见性或租户边界。
- 数据库：无影响。无 SQL、DAO、DO、Entity、索引、软删除或时间字段变更。

#### 验证证据

- `go test ./hack/tools/linactl -run 'TestRunBuildRunsPluginBuildHookBeforeBackendCompile|TestDiscoverPluginBuildHookRootsSkipsPluginsWithoutBuildScript|TestOfficialPluginBuildEnvSeparatesHostOnlyAndPluginFullModes|TestResolveOfficialPluginBuildModeAutoDetectsWorkspace|TestEnsurePackedPublicPlaceholderCreatesGitkeep' -count=1`：通过。
- `make -C apps/lina-plugins/john-ai-agentbox -n build`：通过，输出插件本地`pnpm --dir ".../frontend" run build`命令。
- `make -C apps/lina-plugins/john-ai-agentbox build`：通过，生成`frontend/dist`、gzip 和 Brotli 产物。
- `go run ./hack/tools/linactl build dir=apps/lina-plugins/john-ai-agentbox`：通过，确认根构建调度读取插件`Makefile`步骤。
- `env GOWORK=off GOCACHE=/Users/john/Workspace/github/gqcn/agentbox/temp/go-build-cache go test ./... -count=1`，工作目录`apps/lina-plugins/john-ai-agentbox`：通过。
- `pnpm -C hack/tests exec tsc --noEmit -p tsconfig.json`：通过。
- `node hack/tests/scripts/validate-e2e.mjs`：通过。
- `openspec validate migrate-john-ai-agentbox --strict`：通过。
- `git diff --check`：通过。
- `git check-ignore -v apps/lina-plugins/john-ai-agentbox/frontend/dist/index.html ...`：通过，命中`.gitignore:60:**/dist/`。
- `git ls-files apps/lina-plugins/john-ai-agentbox/frontend/dist`和`git ls-files --others --exclude-standard | rg '^apps/lina-plugins/john-ai-agentbox/frontend/dist' || true`：均无输出。

#### 摘要

- 严重：`0`
- 警告：`0`
- 剩余风险：本次未运行完整浏览器 E2E；该反馈不改变运行时业务路径，已由构建链路、编译门禁和治理检查覆盖。

## 修复 FB-10: `make build`跨平台目录定向构建

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | `linactl build`只有仓库全量构建模式，根`Makefile`也未透传目录选择参数；`make -C <dir>`虽可作为本地便利入口，但不覆盖仓库`make.cmd`的 Windows 跨平台入口，不能作为框架默认构建契约。 |
| 修复 | 根`hack/makefiles/build.mk`透传`dir=<path>`到`linactl build`；`linactl build`新增跨平台`dir`参数。未传`dir`时继续全量构建宿主框架后端、默认管理工作台前端、宿主`manifest`资源和所有启用插件；传入`dir`时按目录类型定向构建`apps/lina-vben`、`apps/lina-core`、`apps/lina-plugins/<plugin-id>`或带非空`package.json`构建脚本的普通包目录。 |
| 默认全量构建语义 | 根目录`make build`和`make.cmd build`仍由`linactl build`承载。插件工作区存在插件清单时，默认启用插件完整模式，执行插件根`Makefile`中声明的`PLUGIN_BUILD_STEP_*`步骤，构建动态插件`Wasm`产物，并在宿主 Go 编译前完成资源准备。 |
| 定向构建语义 | `dir=apps/lina-vben`只构建默认工作台前端并刷新宿主`internal/packed/public`；`dir=apps/lina-core`先构建默认工作台前端、准备宿主`manifest`嵌入资源，再编译宿主后端；`dir=apps/lina-plugins/<plugin-id>`使用和全量构建一致的官方插件构建环境执行该插件`Makefile`声明的构建步骤，动态插件目录则走`runWasm plugin_dir=<dir>`；插件目录内`make build`直接执行同一个`PLUGIN_BUILD_STEP_*`命令。 |
| 跨平台影响 | 长期维护构建分流逻辑在 Go 工具`linactl`中实现，根`Makefile`和`make.cmd`仍是薄包装入口。`dir`参数使用仓库相对路径或绝对路径，目录解析、插件识别、插件`Makefile`读取和构建分流均使用 Go 标准库，不依赖`make -C`、Shell 或平台专属命令。 |
| DI 来源检查 | 不新增运行期 service、构造函数参数、插件 host service、缓存敏感服务或请求路径依赖；所有变更位于开发工具、构建入口、文档和 OpenSpec 记录。 |
| API 契约影响 | 不新增或修改 HTTP API、路由、HTTP 方法、DTO、`g.Meta`、权限标签、响应字段或前端调用契约。 |
| 架构和插件影响 | 不修改`apps/lina-core`核心领域契约或工作台展示适配；仅调整框架构建工具对宿主、默认工作台和插件目录的编译编排。插件生命周期资源仍留在插件目录内，源码插件仍通过编译期`go:embed`嵌入自己的构建产物。 |
| 数据权限影响 | 不新增或改变业务数据读取、列表、详情、下载、写操作、聚合、批量操作、插件 host service 数据访问路径、用户资源可见性或租户边界；构建命令只读源码和构建资源。 |
| 数据库和 SQL 影响 | 不新增或修改 SQL、DAO、DO、Entity、索引、软删除、时间字段或 DAO 生成输入。 |
| 缓存一致性影响 | 不新增或修改缓存、快照、派生状态、失效、刷新、预热、跨实例同步或运行时配置缓存。 |
| `i18n`影响 | 本反馈只修改构建工具帮助文档、英文/中文`linactl`说明和 OpenSpec 记录，不新增运行时 UI 文案、菜单、按钮、API 文档源文本、错误 fallback、插件清单文案或语言包资源。 |
| 文档影响 | `hack/tools/linactl/README.md`和`README.zh-CN.md`同步补充`dir`参数示例和默认全量构建语义，命令帮助同步展示`dir=apps/lina-plugins/<plugin-id>`；插件目录`README.md`和`README.zh-CN.md`继续说明可执行`make build`。 |
| 测试策略 | 该反馈属于构建工具行为修复，无用户可观察业务 UI 变化，不新增浏览器 E2E；使用`linactl`单元测试覆盖插件定向构建、默认工作台定向构建、宿主后端定向构建、插件构建钩子默认全量顺序，以及`make -n`、帮助命令、OpenSpec 校验和格式检查闭环。 |
| 验证结果 | `go test ./hack/tools/linactl -run 'TestRunBuildRunsPluginBuildHookBeforeBackendCompile|TestRunBuildDirBuildsSelectedPluginOnly|TestRunBuildDirBuildsHostFrontendOnly|TestRunBuildDirBuildsHostBackendWithPreparedAssets|TestDiscoverPluginBuildHookRootsSkipsPluginsWithoutBuildScript|TestEnsurePackedPublicPlaceholderCreatesGitkeep' -count=1`通过；`make -n build dir=apps/lina-plugins/john-ai-agentbox`输出`go run . build dir=apps/lina-plugins/john-ai-agentbox verbose=0`，确认根`Makefile`透传；`make -C apps/lina-plugins/john-ai-agentbox -n build`输出插件本地`pnpm --dir ".../frontend" run build`命令；`go run ./hack/tools/linactl build dir=apps/lina-plugins/john-ai-agentbox`通过，确认根构建入口读取插件`Makefile`步骤；`go run ./hack/tools/linactl help build`展示`dir=apps/lina-plugins/<plugin-id>`；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过。`go test ./hack/tools/linactl -count=1`未完全通过，失败项为既有 dev/status 服务测试对 PID 文件名的预期不一致，与本次`build`变更无关。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`documentation.md`、`architecture.md`、`data-permission.md`、`plugin.md`、`cache-consistency.md`、`dev-tooling.md`、`testing.md`、`backend-go.md`、`i18n.md`和`.agents/instructions/markdown-format.instructions.md`；插件目录无本地`AGENTS.md`。 |

### Lina 审查报告（FB-10）

**变更：** `migrate-john-ai-agentbox`
**范围：** 反馈级审查，覆盖`hack/makefiles/build.mk`、`hack/makefiles/plugin.codegen.mk`、`apps/lina-plugins/john-ai-agentbox/Makefile`、`hack/tools/linactl/command_build.go`、`hack/tools/linactl/command.go`、`hack/tools/linactl/main_test.go`、`hack/tools/linactl/README.md`、`hack/tools/linactl/README.zh-CN.md`和本任务记录；范围来源为`git status --short`、`git diff -- hack/makefiles/build.mk hack/makefiles/plugin.codegen.mk apps/lina-plugins/john-ai-agentbox/Makefile hack/tools/linactl/command_build.go hack/tools/linactl/command.go hack/tools/linactl/main_test.go hack/tools/linactl/README.md hack/tools/linactl/README.zh-CN.md openspec/changes/migrate-john-ai-agentbox/tasks.md`和`openspec validate migrate-john-ai-agentbox --strict`。`apps/lina-plugins/john-ai-agentbox/AGENTS.md`不存在。
**已读取规则文件：** `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/plugin.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/testing.md`、`.agents/rules/backend-go.md`、`.agents/rules/i18n.md`和`.agents/instructions/markdown-format.instructions.md`。

#### 发现的问题

- 未发现阻塞问题。

#### 规则域结论

- OpenSpec：通过。`FB-10`已记录根因、修复、影响分析、验证证据和剩余风险；活跃变更仍位于`openspec/changes/migrate-john-ai-agentbox`。
- 开发工具：通过。跨平台构建契约收敛到`linactl build dir=...`，根`Makefile`和插件共享`Makefile`片段只透传参数，`make.cmd`天然转发同一参数，不依赖`make -C`作为默认跨平台契约。
- 后端 Go 编译门禁：通过目标包测试覆盖`command_build.go`和新增测试；完整`go test ./hack/tools/linactl -count=1`存在与本次无关的 dev/status PID 文件预期失败，已记录剩余风险。
- 插件和架构：通过。插件构建由插件根`Makefile`维护，未把插件生命周期资源回流宿主；未修改`apps/lina-core`核心领域契约或工作台展示适配。
- 文档：通过。`linactl`英文/中文 README 同步说明`dir`参数和默认全量构建语义，命令帮助同步更新。
- 测试策略：通过。本反馈为构建工具行为修复，无用户可观察业务 UI 变化，不触发 E2E 质量审查；单元测试覆盖目录构建分流，治理验证覆盖 Makefile 透传、帮助文本、OpenSpec 和格式。
- `i18n`：无影响。未新增运行时 UI 文案、API 文档源文本、错误 fallback、插件清单文案或语言包资源。
- 缓存一致性：无影响。未新增缓存、快照、失效、刷新或跨实例同步。
- 数据权限：无影响。未新增业务数据读取或写入路径，不改变租户、组织或用户资源可见性。
- 数据库：无影响。无 SQL、DAO、DO、Entity、索引、软删除或时间字段变更。

#### 验证证据

- `go test ./hack/tools/linactl -run 'TestRunBuildRunsPluginBuildHookBeforeBackendCompile|TestRunBuildDirBuildsSelectedPluginOnly|TestRunBuildDirBuildsHostFrontendOnly|TestRunBuildDirBuildsHostBackendWithPreparedAssets|TestDiscoverPluginBuildHookRootsSkipsPluginsWithoutBuildScript|TestEnsurePackedPublicPlaceholderCreatesGitkeep' -count=1`：通过。
- `make -n build dir=apps/lina-plugins/john-ai-agentbox`：通过，输出`go run . build dir=apps/lina-plugins/john-ai-agentbox verbose=0`。
- `make -C apps/lina-plugins/john-ai-agentbox -n build`：通过，输出插件本地`pnpm --dir ".../frontend" run build`命令。
- `go run ./hack/tools/linactl build dir=apps/lina-plugins/john-ai-agentbox`：通过，确认根构建入口读取插件`Makefile`步骤。
- `go run ./hack/tools/linactl help build`：通过，输出包含`dir=apps/lina-plugins/<plugin-id>`。
- `openspec validate migrate-john-ai-agentbox --strict`：通过。
- `git diff --check`：通过。
- `go test ./hack/tools/linactl -count=1`：未通过，失败项为`TestPrintStatusTableIncludesDevelopmentServiceDetails`和`TestRunDevStartsServicesAsAsyncProcessesAndPrintsFinalStatus`，表现为 dev/status PID 文件名预期不一致；该失败不涉及本次`build`路径。

#### 摘要

- 严重：`0`
- 警告：`0`
- 剩余风险：完整`linactl`包测试仍有与本次构建变更无关的 dev/status 测试失败；本次构建入口已由目标单测和治理验证覆盖。

## 修复 FB-11: 插件构建逻辑下沉到插件`Makefile`

反馈修复记录：

| 项目 | 记录 |
|------|------|
| 根因 | `hack/makefiles/plugin.codegen.mk`在共享片段中定义了通用`build`目标，并转调`linactl build dir="$(PLUGIN_ROOT)"`。这会把所有源码插件的`make build`固定为同一个框架分发入口，且容易形成`linactl build dir=<plugin>`与插件`make build`之间的递归设计；同时不同插件的构建技术栈和命令可能不同，不能在共享片段中固定。 |
| 修复 | 删除共享`plugin.codegen.mk`中的`build`目标，仅保留`ctrl`、`dao`等通用代码生成入口。`apps/lina-plugins/john-ai-agentbox/Makefile`新增插件自有`PLUGIN_BUILD_STEP_1 := pnpm --dir "$(PLUGIN_ROOT)/frontend" run build`和`build`目标，插件目录内`make build`直接执行该插件前端构建。 |
| 框架构建调度 | `linactl build`在源码插件目录中读取插件根`Makefile`的`PLUGIN_BUILD_STEP_*`声明，并由 Go 工具直接执行这些步骤。这样插件自己的`Makefile`维护构建事实来源，根构建仍保留跨平台`linactl build dir=<plugin>`入口，不需要系统通过`make -C`运行插件钩子。 |
| 跨平台影响 | `make -C apps/lina-plugins/john-ai-agentbox build`仍是本地插件开发便利入口；仓库根`make build`、`make.cmd build`和`go run ./hack/tools/linactl build`由 Go 工具直接解析插件`Makefile`中的简单变量并执行命令，不依赖`GNU Make`、POSIX Shell、PowerShell 或平台专属路径语义。当前解析范围刻意只覆盖`PLUGIN_BUILD_STEP_*`、`$(PLUGIN_ROOT)`和`$(REPO_ROOT)`，避免把`linactl`变成通用 Makefile 解释器。 |
| DI 来源检查 | 不新增运行期 service、构造函数参数、插件 host service、缓存敏感服务或请求路径依赖；所有变更位于开发工具、构建入口、文档和 OpenSpec 记录。 |
| API 契约影响 | 不新增或修改 HTTP API、路由、HTTP 方法、DTO、`g.Meta`、权限标签、响应字段或前端调用契约。 |
| 架构和插件影响 | 不修改`apps/lina-core`核心领域契约或工作台展示适配；插件生命周期资源仍留在插件目录内，插件构建命令由插件根`Makefile`维护，共享框架只读取窄构建步骤声明。 |
| 数据权限影响 | 不新增或改变业务数据读取、列表、详情、下载、写操作、聚合、批量操作、插件 host service 数据访问路径、用户资源可见性或租户边界；构建命令只读源码和构建资源。 |
| 数据库和 SQL 影响 | 不新增或修改 SQL、DAO、DO、Entity、索引、软删除、时间字段或 DAO 生成输入。 |
| 缓存一致性影响 | 不新增或修改缓存、快照、派生状态、失效、刷新、预热、跨实例同步或运行时配置缓存。 |
| `i18n`影响 | 本反馈只修改构建入口、英文/中文`linactl`说明和 OpenSpec 记录，不新增运行时 UI 文案、菜单、按钮、API 文档源文本、错误 fallback、插件清单文案或语言包资源。 |
| 文档影响 | `hack/tools/linactl/README.md`和`README.zh-CN.md`同步说明源码插件通过插件根`Makefile`的`PLUGIN_BUILD_STEP_*`维护构建步骤；插件 README 已说明可执行`make build`且`frontend/dist`是生成产物。 |
| 测试策略 | 该反馈属于构建工具行为修复，无用户可观察业务 UI 变化，不新增浏览器 E2E；使用`linactl`单元测试、插件本地 dry-run、插件真实构建、根`linactl build dir=<plugin>`真实执行、OpenSpec 严格校验和格式检查闭环。 |
| 验证结果 | `go test ./hack/tools/linactl -run 'TestRunBuildRunsPluginBuildHookBeforeBackendCompile|TestRunBuildDirBuildsSelectedPluginOnly|TestRunBuildDirBuildsHostFrontendOnly|TestRunBuildDirBuildsHostBackendWithPreparedAssets|TestDiscoverPluginBuildHookRootsSkipsPluginsWithoutBuildScript|TestEnsurePackedPublicPlaceholderCreatesGitkeep' -count=1`通过；`make -n build dir=apps/lina-plugins/john-ai-agentbox`输出根`linactl build dir=...`调用；`make -C apps/lina-plugins/john-ai-agentbox -n build`输出插件本地`pnpm --dir ".../frontend" run build`命令；`make -C apps/lina-plugins/john-ai-agentbox build`通过；`go run ./hack/tools/linactl build dir=apps/lina-plugins/john-ai-agentbox`通过；`go run ./hack/tools/linactl help build`通过；`openspec validate migrate-john-ai-agentbox --strict`通过；`git diff --check`通过。 |
| 已读取规则 | 已按`AGENTS.md`读取`openspec.md`、`documentation.md`、`architecture.md`、`data-permission.md`、`plugin.md`、`cache-consistency.md`、`dev-tooling.md`、`testing.md`、`backend-go.md`、`i18n.md`和`.agents/instructions/markdown-format.instructions.md`；插件目录无本地`AGENTS.md`。 |

### Lina 审查报告（FB-11）

**变更：** `migrate-john-ai-agentbox`
**范围：** 反馈级审查，覆盖`apps/lina-plugins/john-ai-agentbox/Makefile`、`hack/makefiles/plugin.codegen.mk`、`hack/tools/linactl/command_build.go`、`hack/tools/linactl/main_test.go`、`hack/tools/linactl/README.md`、`hack/tools/linactl/README.zh-CN.md`和本任务记录；范围来源为`git status --short`、`git diff -- apps/lina-plugins/john-ai-agentbox/Makefile hack/makefiles/plugin.codegen.mk hack/tools/linactl/command_build.go hack/tools/linactl/main_test.go hack/tools/linactl/README.md hack/tools/linactl/README.zh-CN.md openspec/changes/migrate-john-ai-agentbox/tasks.md`和`openspec validate migrate-john-ai-agentbox --strict`。`apps/lina-plugins/john-ai-agentbox/AGENTS.md`不存在。
**已读取规则文件：** `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/data-permission.md`、`.agents/rules/plugin.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/testing.md`、`.agents/rules/backend-go.md`、`.agents/rules/i18n.md`和`.agents/instructions/markdown-format.instructions.md`。

#### 发现的问题

- 未发现阻塞问题。

#### 规则域结论

- OpenSpec：通过。`FB-11`记录根因、修复、影响分析、验证证据和审查结论；任务在验证完成后标记。
- 开发工具：通过。插件构建事实来源下沉到插件根`Makefile`，根构建仍由 Go 工具`linactl`直接执行插件声明的`PLUGIN_BUILD_STEP_*`，不依赖`make -C`作为跨平台默认路径。
- 插件和架构：通过。共享`plugin.codegen.mk`不再固定插件构建命令；`john-ai-agentbox`插件构建步骤留在插件目录内，未修改宿主核心领域契约。
- 后端 Go 编译门禁：通过。`linactl`构建分流和 Makefile 步骤解析由目标单元测试覆盖；新增 Go 私有函数保持注释和窄职责。
- 文档：通过。`linactl`英文/中文 README 同步说明插件`Makefile`构建步骤声明。
- 测试策略：通过。本反馈为构建工具行为修复，无用户可观察业务 UI 变化，不触发 E2E 质量审查；使用工具单测、插件本地构建、根构建定向执行、OpenSpec 校验和格式检查覆盖。
- `i18n`：无影响。未新增运行时 UI 文案、API 文档源文本、错误 fallback、插件清单文案或语言包资源。
- 缓存一致性：无影响。未新增缓存、快照、失效、刷新或跨实例同步。
- 数据权限：无影响。未新增业务数据读取或写入路径，不改变租户、组织或用户资源可见性。
- 数据库：无影响。无 SQL、DAO、DO、Entity、索引、软删除或时间字段变更。

#### 验证证据

- `go test ./hack/tools/linactl -run 'TestRunBuildRunsPluginBuildHookBeforeBackendCompile|TestRunBuildDirBuildsSelectedPluginOnly|TestRunBuildDirBuildsHostFrontendOnly|TestRunBuildDirBuildsHostBackendWithPreparedAssets|TestDiscoverPluginBuildHookRootsSkipsPluginsWithoutBuildScript|TestEnsurePackedPublicPlaceholderCreatesGitkeep' -count=1`：通过。
- `make -n build dir=apps/lina-plugins/john-ai-agentbox`：通过，输出根`linactl build dir=...`调用。
- `make -C apps/lina-plugins/john-ai-agentbox -n build`：通过，输出插件本地`pnpm --dir ".../frontend" run build`命令。
- `make -C apps/lina-plugins/john-ai-agentbox build`：通过。
- `go run ./hack/tools/linactl build dir=apps/lina-plugins/john-ai-agentbox`：通过。
- `go run ./hack/tools/linactl help build`：通过。
- `openspec validate migrate-john-ai-agentbox --strict`：通过。
- `git diff --check`：通过。

#### 摘要

- 严重：`0`
- 警告：`0`
- 剩余风险：`linactl`只解析插件`Makefile`中用于构建调度的简单`PLUGIN_BUILD_STEP_*`变量，不解释任意 Makefile 语法；这是刻意限制的跨平台契约。
