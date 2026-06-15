## ADDED Requirements

### Requirement: `john-ai-agentbox`必须作为源码插件维护

系统 SHALL 将外部`agent-box`项目迁移为源码插件`john-ai-agentbox`，插件源码根目录 MUST 为`apps/lina-plugins/john-ai-agentbox/`。插件目录 MUST 遵守 LinaPro 源码插件同构目录结构，并维护`plugin.yaml`、`plugin_embed.go`、`backend/`、`frontend/`、`manifest/`、`README.md`和`README.zh-CN.md`。插件 ID MUST 与目录名、`plugin.yaml id`、URL path、资产路径、权限命名空间、数据库资源标识和测试目录保持一致。

#### Scenario: 插件目录和清单身份一致
- **WHEN** 实现`john-ai-agentbox`插件
- **THEN** 插件目录为`apps/lina-plugins/john-ai-agentbox/`
- **AND** `plugin.yaml`中的`id`为`john-ai-agentbox`
- **AND** 源码插件注册、public asset、API、权限和测试资源均使用该插件 ID

#### Scenario: 插件资源不回流宿主
- **WHEN** 迁移 AgentBox 后端、前端、SQL、配置、文档或测试资源
- **THEN** 资源 MUST 位于`apps/lina-plugins/john-ai-agentbox/`下的对应目录
- **AND** 不得把 AgentBox 业务页面、SQL、DAO、Service 或运行时资源放入`apps/lina-core`作为内建功能

### Requirement: AgentBox 插件必须提供并读取业务配置

系统 SHALL 在`apps/lina-plugins/john-ai-agentbox/manifest/config/config.yaml`中维护插件运行所需的业务配置默认值，并在`config.example.yaml`中提供同字段配置模板。插件启动装配 SHALL 通过宿主既有插件作用域配置服务读取`john-ai-agentbox`配置，并把认证会话、provider 远程请求、AI 调用、Docker runtime、workspace 和 service discovery 的纯值配置显式传入插件内部 service 或 runtime backend。该能力 MUST 闭环在`john-ai-agentbox`插件内部，不得为本插件新增或修改`apps/lina-core`主框架配置契约、插件清单协议或宿主 service 语义。

#### Scenario: manifest/config 包含插件业务配置
- **WHEN** 查看`apps/lina-plugins/john-ai-agentbox/manifest/config/config.yaml`
- **THEN** 文件包含`auth.sessionTtl`、`providers.requestTimeout`、`providers.remoteModelSyncLimit`、`ai.requestTimeout`、`runtime.mode`、`runtime.docker.*`、`runtime.workspace.*`和`runtime.services.discoveryLimit`
- **AND** `config.example.yaml`提供同字段模板

#### Scenario: 启动装配读取插件作用域配置
- **WHEN** `john-ai-agentbox`注册源码插件 HTTP 路由
- **THEN** 插件通过`registrar.Services().Plugins().Config()`读取插件作用域配置
- **AND** 认证 TTL、provider/AI timeout、Docker host、容器日志行数、停止超时、workspace 根路径、上传/预览/技能上限和服务发现上限均来自该配置或插件内部默认值
- **AND** 请求处理路径不得临时创建第二套关键 service 图

#### Scenario: 配置变更不修改主框架
- **WHEN** 补齐或调整`john-ai-agentbox`业务配置项
- **THEN** 修改范围 MUST 保持在`apps/lina-plugins/john-ai-agentbox/`和当前 OpenSpec 记录内
- **AND** 不得修改`apps/lina-core`主框架代码，除非先向用户说明不可避免原因并获得确认

### Requirement: AgentBox 页面入口必须闭环在插件内部

系统 SHALL 在`john-ai-agentbox`安装并启用后，通过插件自身源码路由注册精确页面入口`GET /`和`GET /login`。入口实现 MUST 闭环在`apps/lina-plugins/john-ai-agentbox/`内，不得要求新增`lina-core`清单字段、宿主 SPA fallback、宿主缓存或主框架路由协议。`GET /` MUST 直接返回 AgentBox 工作台 SPA 的`index.html`内容，浏览器地址 MUST 保持为`/`，不得通过`302`、`303`或等价跳转把用户带到`/x-assets/john-ai-agentbox/{version}/index.html`。默认 LinaPro 管理工作台 MUST 继续通过`/admin`访问。AgentBox 页面入口不得自动投影为 LinaPro 管理工作台菜单、角色权限节点或 OpenAPI 路径。

#### Scenario: 根路径打开 AgentBox 功能界面
- **WHEN** `john-ai-agentbox`已安装并启用
- **AND** 浏览器访问`/`
- **THEN** 插件精确入口直接返回 AgentBox 工作台 SPA 的`index.html`
- **AND** 浏览器地址保持为`/`
- **AND** 不通过`/x-assets/john-ai-agentbox/{version}/index.html`作为用户可见入口
- **AND** 不返回 LinaPro 管理工作台 SPA

#### Scenario: `/login`打开 AgentBox 登录界面
- **WHEN** `john-ai-agentbox`已安装并启用
- **AND** 浏览器访问`/login`
- **THEN** 插件精确入口直接返回 AgentBox SPA 的`index.html`并渲染独立登录页
- **AND** 浏览器地址保持为`/login`
- **AND** 该登录界面不依赖 LinaPro 管理工作台登录状态

#### Scenario: `/admin`继续打开管理工作台
- **WHEN** 浏览器访问`/admin`或`/admin/*`
- **THEN** 系统返回 LinaPro 默认管理工作台
- **AND** AgentBox 不得注册通配页面路由吞掉该路径

#### Scenario: 插件门户不进入工作台治理
- **WHEN** `john-ai-agentbox`通过`/`或`/login`提供页面
- **THEN** 管理工作台菜单不得自动新增 AgentBox 门户入口
- **AND** 角色授权不得因为门户路由自动新增权限节点

### Requirement: AgentBox 后端接口必须使用插件 API 命名空间

系统 SHALL 将 AgentBox JSON API 迁移到`/x/john-ai-agentbox/api/v1/...`。源码插件 API MUST 通过插件 registrar 的`APIPrefix()`或等价能力注册，不得挂载到宿主`/api`或`/api/v1`命名空间。API DTO MUST 遵守 RESTful 方法语义、`g.Meta`文档标签、`dc`和`eg`字段要求，所有公开时间点响应字段 MUST 使用 Unix 毫秒时间戳。

#### Scenario: 登录 API 使用插件命名空间
- **WHEN** 前端提交 AgentBox 登录请求
- **THEN** 请求路径为`POST /x/john-ai-agentbox/api/v1/auth/sessions`
- **AND** 系统不得使用`POST /api/auth/sessions`作为生产入口

#### Scenario: Agent 资源 API 使用插件命名空间
- **WHEN** 前端查询 Agent 列表
- **THEN** 请求路径为`GET /x/john-ai-agentbox/api/v1/agents`
- **AND** 响应包含当前页面所需的最小 Agent 投影和必要关联投影
- **AND** 前端不得为了渲染列表对每个 Agent 自动补查详情

#### Scenario: 旧根级 API 不保留兼容入口
- **WHEN** 浏览器或客户端请求`/api/agents`、`/api/providers`或`/api/ai/capability-tiers`
- **THEN** 这些路径不得作为 AgentBox 生产 API 入口
- **AND** 宿主`/api`命名空间继续只归 LinaPro 控制面所有

#### Scenario: 时间字段使用 Unix 毫秒
- **WHEN** AgentBox API 返回`createdAt`、`updatedAt`、`lastLoginAt`、`expiresAt`、`startedAt`或`endedAt`等时间点字段
- **THEN** JSON 字段值 MUST 为 Unix timestamp in milliseconds
- **AND** 不得返回格式化时间字符串作为时间点字段

### Requirement: WebSocket、preview 和服务代理必须归属插件路径

系统 SHALL 将 AgentBox WebSocket、workspace preview、Agent runtime service proxy 和 TCP tunnel 等非普通 JSON API 入口迁移到`/x/john-ai-agentbox/api/v1/...`下的插件自有路径。此类入口 MUST 复用插件启用检查、AgentBox 独立会话鉴权和目标资源归属校验，不得注册根级`/ws`、`/s`或其他宿主级通配代理路径。

#### Scenario: Chat WebSocket 使用插件路径
- **WHEN** 前端连接某个 Agent Chat session
- **THEN** WebSocket URL 使用`/x/john-ai-agentbox/api/v1/ws/agents/{agentId}/chat/sessions/{sessionId}`或等价插件路径
- **AND** 系统校验`agent_box_session`和目标 Agent/session 归属后才建立连接

#### Scenario: Shell WebSocket 使用插件路径
- **WHEN** 前端连接某个 Agent shell
- **THEN** WebSocket URL 使用`/x/john-ai-agentbox/api/v1/ws/agents/{agentId}/shell`
- **AND** 系统不得使用根级`/ws/agents/{agentId}/shell`作为生产入口

#### Scenario: Agent runtime 服务代理使用插件路径
- **WHEN** 用户打开 Agent runtime 暴露的 HTTP 服务
- **THEN** 代理 URL MUST 位于`/x/john-ai-agentbox/api/v1/proxy/...`或等价插件子路径
- **AND** 代理处理前 MUST 校验当前 AgentBox 用户对目标 Agent、service 和 bridge 的可见性

### Requirement: AgentBox 登录态必须与 LinaPro 管理工作台隔离

系统 SHALL 由`john-ai-agentbox`插件独立维护登录页面、用户表、密码哈希、会话表、会话 token 哈希、Cookie 和鉴权中间件。AgentBox 登录成功 MUST 写入独立 HttpOnly Cookie`agent_box_session`。LinaPro 管理工作台登录状态不得使用户自动登录 AgentBox，AgentBox 登录状态也不得使用户自动登录`/admin`。

#### Scenario: AgentBox 登录写入独立 Cookie
- **WHEN** 用户通过`/login`提交有效 AgentBox 凭据
- **THEN** 系统创建插件自有用户会话
- **AND** 响应写入 HttpOnly Cookie`agent_box_session`
- **AND** 数据库仅保存该 opaque token 的哈希值

#### Scenario: 管理工作台登录不等于 AgentBox 登录
- **WHEN** 用户只登录 LinaPro 管理工作台
- **AND** 浏览器访问`/`
- **THEN** AgentBox 仍按`agent_box_session`判断是否已登录
- **AND** 不得因管理工作台 token 存在而视为 AgentBox 已认证

#### Scenario: AgentBox 登录不等于管理工作台登录
- **WHEN** 用户只登录 AgentBox
- **AND** 浏览器访问`/admin`
- **THEN** LinaPro 管理工作台仍按自身认证机制判断登录状态
- **AND** 不得因`agent_box_session`存在而视为管理工作台已认证

#### Scenario: 注销只清理 AgentBox 会话
- **WHEN** 用户调用`DELETE /x/john-ai-agentbox/api/v1/auth/session`
- **THEN** 系统撤销 AgentBox 插件会话并清理`agent_box_session`
- **AND** 不得清理 LinaPro 管理工作台登录状态

### Requirement: AgentBox AI 数据必须使用插件自有表空间

系统 SHALL 由`john-ai-agentbox`插件独立维护`AI`渠道、模型、能力档位、调用日志、Agent、Chat、Prompt、Workspace、Terminal 和用户设置等数据。所有插件表名、索引名、外键名和 DAO 生成输入 MUST 使用`john_ai_agentbox_`前缀或与该前缀一致的资源命名。插件不得读取、写入或复用宿主`AI`表、`linapro-ai-core`插件表或宿主`DAO/DO/Entity`作为自身业务存储。

#### Scenario: 插件安装创建自有表
- **WHEN** 安装或启用`john-ai-agentbox`
- **THEN** 插件 SQL 创建`john_ai_agentbox_ai_providers`、`john_ai_agentbox_provider_models`、`john_ai_agentbox_coding_agents`、`john_ai_agentbox_agent_chat_sessions`、`john_ai_agentbox_users`和其他插件自有表
- **AND** 不创建或修改无`john_ai_agentbox_`前缀的 AgentBox 业务表

#### Scenario: AI 渠道数据不复用宿主
- **WHEN** AgentBox 管理 AI provider、provider model、能力档位或调用日志
- **THEN** 数据 MUST 写入`john_ai_agentbox_`前缀表
- **AND** 不得写入`linapro-ai-core`或宿主文本`AI`能力的渠道、模型、档位和调用日志表

#### Scenario: 插件 DAO 归属插件目录
- **WHEN** 为 AgentBox 表生成 DAO、DO 或 Entity
- **THEN** 生成结果 MUST 位于`apps/lina-plugins/john-ai-agentbox/backend/internal/dao/`和`backend/internal/model/`
- **AND** 生产代码不得 import 宿主核心 DAO 或其他插件 DAO 作为 AgentBox 存储入口

### Requirement: AgentBox 数据访问必须按插件用户和资源归属隔离

系统 SHALL 使用 AgentBox 插件自有用户 ID 和资源归属作为业务数据可见性边界。读取列表、详情、下载、workspace 文件、Chat message、Terminal session、服务代理和写操作时，系统 MUST 在数据库查询阶段或动作执行前校验当前 AgentBox 用户对目标资源的归属关系。不可见资源 MUST 与不存在资源保持等价拒绝语义。

#### Scenario: Agent 列表只返回当前 AgentBox 用户资源
- **WHEN** 已登录 AgentBox 用户查询`GET /x/john-ai-agentbox/api/v1/agents`
- **THEN** 系统只返回该 AgentBox 用户可见的 Agent
- **AND** 数据库查询不得先读取所有用户 Agent 后在内存过滤

#### Scenario: 详情操作校验目标归属
- **WHEN** 已登录 AgentBox 用户读取、更新、启动、停止或删除某个 Agent
- **THEN** 系统在执行动作前校验该 Agent 属于当前 AgentBox 用户
- **AND** 不可见 Agent 不得泄露名称、状态、provider、runtime 或 workspace 信息

#### Scenario: 批量或关联资源不泄露其他用户数据
- **WHEN** 用户查询 Chat sessions、workspace resources、terminal sessions 或 service bridges
- **THEN** 查询 MUST 先限定到当前用户可见 Agent 或 session
- **AND** 分页总数、候选项、排序和错误信息不得泄露其他用户资源存在性

### Requirement: AgentBox 高频接口必须具备有界装配和索引支撑

系统 SHALL 在 AgentBox 列表、详情批量、workspace tree、Chat 历史、服务发现、调用日志和下拉候选等高频接口中控制数据规模和数据库访问次数。接口 MUST 具备分页、范围过滤、数量上限或异步刷新边界；后端 MUST 使用集合化查询、投影查询、批量关联、缓存或快照避免`N+1`查询。

#### Scenario: Provider 列表批量装配模型摘要
- **WHEN** 前端查询 provider 列表
- **THEN** 后端使用集合化查询或聚合查询装配当前页 provider 的 model 摘要
- **AND** 不得为每个 provider 循环查询 model 列表

#### Scenario: Agent 列表批量装配关联投影
- **WHEN** 前端查询 Agent 列表
- **THEN** 后端一次性装配当前页所需 provider、model、image、runtime 和 activity 投影
- **AND** 不得要求前端对每行 Agent 再发起详情查询

#### Scenario: 调用日志按数据库侧过滤分页
- **WHEN** 用户查询 AI 调用日志
- **THEN** 后端 MUST 在数据库侧按时间范围、状态、档位、provider、model 和来源过滤、排序、分页
- **AND** 表结构 MUST 提供支撑主要过滤和排序路径的索引

### Requirement: AgentBox 前端必须使用插件路径配置

系统 SHALL 将 AgentBox React/Vite 前端迁移为插件 public asset 前端，并通过集中配置或常量生成 API base、WebSocket base、public asset base 和登录路径。前端组件不得硬编码生产根级`/api`、`/ws`、`/assets`或服务代理路径。`/`和`/login`入口 MUST 由插件精确后端路由处理，不依赖宿主 SPA fallback。

#### Scenario: 前端请求使用 API base
- **WHEN** 前端调用 AgentBox API
- **THEN** 请求基于`/x/john-ai-agentbox/api/v1`构造
- **AND** 组件内部不得散落硬编码`/api`生产路径

#### Scenario: 前端 WebSocket 使用 WebSocket base
- **WHEN** 前端创建 Chat 或 Shell WebSocket
- **THEN** URL 基于`/x/john-ai-agentbox/api/v1/ws`构造
- **AND** 不得硬编码根级`/ws`

#### Scenario: 页面刷新保留门户体验
- **WHEN** 用户在浏览器刷新`/`或`/login`
- **THEN** 插件精确入口直接返回 AgentBox SPA 的`index.html`
- **AND** 前端根据当前路径渲染功能界面或登录界面

### Requirement: AgentBox 插件必须按插件 i18n 配置治理

系统 SHALL 对`john-ai-agentbox`的运行时文案、菜单、API 文档和错误消息执行插件维度`i18n`影响评估。若`plugin.yaml`未声明`i18n.enabled: true`，该插件视为单语言插件，系统 MUST 不要求补齐插件`manifest/i18n`或`apidoc`翻译资源，也不得把插件文案翻译键写入`lina-core`语言资源。若后续启用插件`i18n`，翻译资源 MUST 维护在插件自己的`manifest/i18n/`目录。

#### Scenario: 单语言插件不维护 i18n 资源
- **WHEN** `john-ai-agentbox`未启用`i18n.enabled: true`
- **THEN** 审查记录 MUST 明确说明单语言插件判断
- **AND** 实现不得为了插件文案向`apps/lina-core/manifest/i18n`写入翻译键

#### Scenario: 启用 i18n 后使用插件资源
- **WHEN** `john-ai-agentbox`后续声明`i18n.enabled: true`
- **THEN** 插件运行时和 API 文档翻译资源 MUST 位于`apps/lina-plugins/john-ai-agentbox/manifest/i18n/`
- **AND** E2E MUST 断言关键用户可见文案显示为目标语言翻译而非原始 key

### Requirement: AgentBox 插件必须提供专属测试覆盖

系统 SHALL 为`john-ai-agentbox`补充插件专属单元测试、后端编译门禁、SQL 静态或初始化验证、前端构建验证和 Playwright E2E。源码插件专属 E2E MUST 位于`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/`，TC ID 从`TC001`开始连续递增。E2E MUST 覆盖`/`入口、`/login`登录入口、独立会话、`/admin`隔离、插件 API 鉴权失败和核心 Agent 工作流。

#### Scenario: 插件 E2E 放在插件目录
- **WHEN** 为 AgentBox 编写 E2E
- **THEN** 测试文件位于`apps/lina-plugins/john-ai-agentbox/hack/tests/e2e/`
- **AND** 文件命名使用`TC{NNN}-{brief-name}.ts`
- **AND** 插件专属 POM 和 helper 位于同插件`hack/tests/pages/`和`hack/tests/support/`

#### Scenario: TC001 覆盖门户和登录
- **WHEN** 执行`TC001-portal-login.ts`
- **THEN** 测试验证`/`未认证时进入 AgentBox 认证流程
- **AND** 验证`/login`入口显示 AgentBox 独立登录页面
- **AND** 验证登录成功后`/`显示 AgentBox 功能界面

#### Scenario: TC002 覆盖管理工作台隔离
- **WHEN** 执行`TC002-admin-isolation.ts`
- **THEN** 测试验证 AgentBox 登录不会让`/admin`自动认证
- **AND** 验证 LinaPro 管理工作台登录不会让 AgentBox 自动认证

#### Scenario: TC003 覆盖 API 和数据隔离
- **WHEN** 执行`TC003-agentbox-api-data-isolation.ts`
- **THEN** 测试验证未带`agent_box_session`访问插件受保护 API 返回认证失败
- **AND** 验证当前用户无法访问其他 AgentBox 用户的 Agent、Chat session、workspace resource 或服务代理
