## Why

`agent-box`已经具备独立的`AI` Agent 产品能力，但它目前位于仓库外部，无法纳入 LinaPro 插件生命周期、路由边界、安装 SQL、权限治理和持续交付流程。将其以插件`john-ai-agentbox`迁入当前项目，可以在不污染`apps/lina-core`核心宿主和不复用宿主已有`AI`数据表的前提下，把插件入口、独立登录和 AgentBox 自有`AI`数据管理纳入统一框架治理。

## What Changes

- 新增源码插件`john-ai-agentbox`，目录为`apps/lina-plugins/john-ai-agentbox/`，插件 ID 固定为`john-ai-agentbox`。
- 将外部项目`/Users/john/Workspace/github/gqcn/agent-box`的一体化前后端能力迁移到插件目录，保留其独立产品体验，不改造成默认管理工作台页面。
- 支持插件内部精确页面入口：访问`/`进入 AgentBox 功能界面，访问`/login`进入 AgentBox 独立登录界面；默认 LinaPro 管理工作台继续位于`/admin`。
- 将 AgentBox 后端 API、WebSocket 和服务代理统一迁移到`/x/john-ai-agentbox/api/v1/...`及插件自有子路径，不再占用宿主`/api/v1`或根级`/ws`、`/s`。
- AgentBox 登录态独立维护，使用插件自有会话、Cookie、鉴权中间件和用户表，不复用 LinaPro 管理工作台 token、用户会话或权限登录状态。
- AgentBox 自有`AI`渠道、模型、Agent、会话、日志等数据表统一使用`john_ai_agentbox_`前缀，不读取、不写入、不复用宿主或`linapro-ai-core`既有`AI`功能表。
- 页面入口闭环在`john-ai-agentbox`插件内部，通过现有源码插件`http.route.register`能力注册精确`GET /`和`GET /login`入口，不新增`lina-core`主框架门户协议、fallback、缓存或清单字段。
- 补充插件专属单元测试和 E2E 测试计划，覆盖门户路由、独立登录、API 鉴权、数据隔离和管理工作台隔离。

## Capabilities

### New Capabilities

- `john-ai-agentbox-plugin`：定义`john-ai-agentbox`源码插件的目录、路由、独立认证、插件自有`AI`数据、API 迁移、前端门户和测试治理要求。

### Modified Capabilities

- 无。根路径和登录页入口由`john-ai-agentbox`插件通过现有源码插件路由能力闭环，本变更不修改`lina-core`宿主能力规范。

## Impact

- 影响插件源码目录：新增`apps/lina-plugins/john-ai-agentbox/`下的`plugin.yaml`、`plugin_embed.go`、`backend/`、`frontend/`、`manifest/`和插件测试资产。
- 不影响宿主通用路由能力：本轮不修改`apps/lina-core`主框架代码、清单协议或插件 public asset 解析能力。
- 影响 API 契约：AgentBox 原`/api`、`/ws`、`/s`路径需要迁移到插件命名空间和 RESTful 契约，公开时间点响应字段使用 Unix 毫秒时间戳。
- 影响数据库：新增插件自有 PostgreSQL SQL、DAO 生成配置、安装 SQL、卸载 SQL和可选 mock 数据；所有表、索引和唯一键必须与`john_ai_agentbox_`前缀绑定。
- 影响前端：迁移 AgentBox React/Vite 前端为插件门户静态资产或插件自管 HTTP 页面，`/`和`/login`不进入 LinaPro 管理工作台菜单与权限模型。
- 影响测试和工具：需要补充插件后端编译门禁、SQL 初始化或静态校验、插件专属 E2E，以及 host-only/plugin-full 构建或开发入口验证。
- 影响治理记录：实现和审查必须记录`i18n`、缓存一致性、数据权限、开发工具跨平台、DI 来源和测试覆盖判断；当前提案阶段不修改运行时代码。
