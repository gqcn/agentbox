## 1. 后端路由契约调整

- [x] 1.1 在动态插件 runtime 中定义 public prefix `/x`，并让路由匹配只支持该入口。
- [x] 1.2 调整 HTTP 路由绑定，使 `/x/{pluginId}/...` 挂载宿主统一中间件和动态插件鉴权/权限中间件，且不依赖 `/api/v1` 路由组。
- [x] 1.3 移除 `/api/v1/extensions/{pluginId}/...` 动态插件分发入口，确保旧路径不再执行动态插件 bridge 路由。
- [x] 1.4 更新动态插件 public path 元数据构建，确保 bridge 请求、审计上下文和中间件拿到实际命中的 public path。

## 2. OpenAPI、文档与 i18n 同步

- [x] 2.1 将动态插件 OpenAPI 投影和插件资源列表 public path 生成切换到 `/x/{pluginId}/...`。
- [x] 2.2 更新 apidoc 动态插件路径识别逻辑，使 `/x/{pluginId}/...` 生成和读取稳定插件翻译 key。
- [x] 2.3 同步更新核心 API DTO 文档、动态示例插件 API 注释、`manifest/i18n` 和 packed apidoc i18n JSON 中关于动态插件公开路径的文案。
- [x] 2.4 静态扫描非归档代码和文档中的 `/api/v1/extensions` 引用，确认仅保留迁移说明或明确断言旧路径不再分发的测试引用。

## 3. 动态示例插件与测试更新

- [x] 3.1 将 `linapro-demo-dynamic` 前端请求基地址切换为 `/x/${pluginId}`，并确保插件自有 API 版本可在内部路径中表达。
- [x] 3.2 更新动态插件 runtime、OpenAPI、controller、pluginbridge 编解码相关单元测试，覆盖 `/x` 路径和旧路径不再分发。
- [x] 3.3 更新官方动态插件 E2E 用例，验证 `/x/{pluginId}/...` 可完成已登录动态插件请求。
- [x] 3.4 更新性能审计脚本或端点扫描脚本中的动态插件路径识别和 public path 生成逻辑。

## 4. 验证与审查

- [x] 4.1 运行 `openspec validate rename-dynamic-plugin-route-prefix --strict`。
- [x] 4.2 运行后端 Go 编译/测试门禁，至少覆盖 `apps/lina-core/internal/cmd`、动态插件 runtime、OpenAPI 投影、apidoc 和 pluginbridge 相关变更包。
- [x] 4.3 运行动态示例插件相关测试或 E2E smoke，覆盖新 `/x` 路径的用户可观察行为。
- [x] 4.4 记录 i18n 影响判断：本次需要同步 apidoc i18n JSON 和动态示例插件文案，不新增运行时菜单翻译键时也需明确说明。
- [x] 4.5 记录缓存一致性判断：本次不新增缓存，必须确认 `/x` 路由仍经过插件 runtime freshness、启用状态和派生缓存失效逻辑。
- [x] 4.6 执行 `/lina-review` 审查，修复发现的问题后再标记任务完成。

## 5. 验证记录

- [x] OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] 后端 Go：`cd apps/lina-core && go test ./internal/cmd ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/openapi ./internal/service/apidoc ./internal/controller/plugin ./pkg/pluginbridge/... ./pkg/pluginservice/contract -count=1`
- [x] 路由回归：`cd apps/lina-core && go test ./internal/cmd -run TestDynamicPluginRootRoutesPrecedeSPAFallback -count=1`
- [x] 动态示例插件 E2E smoke：重启本地开发服务到当前工作区代码后，`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests test:module -- plugin:linapro-demo-dynamic -- --grep "TC-1j"`；验证启用 `linapro-demo-dynamic` 后 `/x/linapro-demo-dynamic/backend-summary` 返回真实 Wasm bridge 响应。
- [x] E2E 治理：`pnpm -C hack/tests test:validate`
- [x] 路径静态扫描：`rg -n '/api/v1/extensions|api/v1/extensions|extensions/\\$\\{|`extensions/|/x//|LegacyRoutePublicPrefix|bindLegacyDynamicPluginAPIRoutes|parts\\[0\\] == "extensions"|/extensions' apps/lina-core apps/lina-plugins/linapro-demo-dynamic hack/tests .agents/skills/lina-perf-audit/scripts -g '!node_modules' -g '!dist' -g '!build' -g '!public/stoplight/**'`；仅剩服务依赖基线中的非路由文件名 `extensions.go` 和旧路径不再分发的负向单元测试引用。
- [x] i18n 影响：已同步核心 API DTO 文档、核心 `manifest/i18n`、核心 packed apidoc i18n、动态示例插件 API 注释和动态示例插件 `manifest/i18n` 中的公开路径文案；本次未新增运行时菜单翻译键。
- [x] 缓存一致性：本次不新增缓存；`/x` 路由继续通过 `PrepareDynamicRouteMiddleware`、`AuthenticateDynamicRouteMiddleware`、`prepareDynamicRouteRuntime` 和 `ensureRuntimeCacheFresh` 相关路径，仍受插件启用状态、runtime freshness、运行时修订号和派生缓存失效机制约束。
- [x] `/lina-review`：审查范围限定为本变更相关路由、runtime、OpenAPI/apidoc、插件示例、E2E 引用、性能审计脚本和 OpenSpec 文档；未发现阻塞问题。已补充 `TestDynamicPluginRootRoutesPrecedeSPAFallback` 防止 `/x/...` 被前端 SPA fallback 接管，并修正旧兼容措辞注释。

## Feedback

- [x] **FB-1**: 将动态插件 API DTO 中的 `gmeta.Meta` 统一改为 `g.Meta`
- [x] **FB-2**: 优化插件管理列表的列对齐、运行时状态列位置、列宽和运行时状态说明
- [x] **FB-3**: 调整插件管理列表列标题居中、名称/版本列宽和基础信息列顺序
- [x] **FB-4**: 将动态插件 apidoc i18n key 生成从硬编码 `/api/v1` 路径 slug 调整为显式 operation identity 优先、通用 dotted path 兜底，且不保留旧 key 兼容
- [x] **FB-5**: 将动态插件 OpenAPI 操作翻译扩展字段从 Lina 专用名称泛化为 `x-i18n-key`
- [x] **FB-6**: 复用 apidoc service 中的 OpenAPI i18n extension 字段定义，移除动态插件 OpenAPI 投影内的重复常量
- [x] **FB-7**: 删除动态插件 OpenAPI `x-i18n-key` 扩展字段，改为将 OpenAPI `operationId` 作为 apidoc i18n key 生成规则的元素
- [x] **FB-8**: 移除 `/api/v1` 作为插件路由强制前缀的实现约束，确保仅 `/x/{pluginId}` 是宿主硬性路由前缀
- [x] **FB-9**: 将动态插件 API DTO 的 `/api/v1` 路径段迁移到构建期路由分组前缀，DTO 仅声明插件本地资源路径
- [x] **FB-10**: 将动态插件 apidoc i18n key 收敛为 `method + public path` 派生，不再使用 `operationId` 或 `RequestType` 作为翻译身份
- [x] **FB-11**: 将动态插件路由分组前缀从生成 API 文件迁移到后端 `RegisterRoutes` 注册声明
- [x] **FB-12**: 补充动态插件 `RegisterRoutes` 和 `registrar.Group(...)` 的中英文注释说明
- [x] **FB-13**: `linactl wasm out=temp/output` 在 `apps/lina-plugins` 下执行时把动态插件产物写入插件工作区 `temp/output`

## Feedback Verification

- [x] FB-1 后端 Go：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/api/... -count=1`
- [x] FB-1 WASM API 编译烟测：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off GOOS=wasip1 GOARCH=wasm go test ./backend/api/dynamic/v1 -run '^$' -count=1`
- [x] FB-1 动态插件产物构建：`cd apps/lina-plugins && go run ../../hack/tools/linactl wasm p=linapro-demo-dynamic out=temp/output`
- [x] FB-1 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-1 i18n 影响：本次仅统一 Go DTO 元数据嵌入类型，未新增、修改或删除用户可见文案、菜单、路由、API 文档源文本或 apidoc i18n JSON。
- [x] FB-1 缓存一致性：本次不新增或修改缓存逻辑，不影响动态插件 runtime freshness、启用状态或派生缓存失效机制。
- [x] FB-1 `/lina-review`：审查范围限定为动态插件 API DTO、独立模块依赖元数据和 OpenSpec 反馈记录；未发现阻塞问题。`g.Meta` 是 `gmeta.Meta` 的类型别名，路由元数据、权限标签和 apidoc 源文本未发生语义变化；`frame/g` 引入的依赖变更已通过普通 Go 测试和实际 `wasip1/wasm` 产物构建验证。
- [x] FB-2 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-2 前端 i18n/E2E 治理：`pnpm -C hack/tests test:validate`
- [x] FB-2 前端类型检查：`pnpm -C apps/lina-vben -F @lina/web-antd typecheck`
- [x] FB-2 E2E：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/hack/tests/e2e/extension/plugin/TC013-plugin-management-table-layout.ts --grep "TC-13a" --workers=1`
- [x] FB-2 i18n 影响：本次新增插件管理页运行时状态列表头说明文案，已同步 `zh-CN` 和 `en-US` 前端运行时语言包；不涉及菜单、路由、manifest i18n 或 apidoc i18n。
- [x] FB-2 缓存一致性：本次仅调整前端表格列展示、列宽和 tooltip 文案，不新增或修改缓存逻辑，不影响插件 runtime freshness、启用状态或派生缓存失效机制。
- [x] FB-2 数据权限影响：本次不新增、修改或扩大任何数据操作接口；插件列表查询仍复用既有接口与权限边界。
- [x] FB-2 `/lina-review`：审查范围包含插件管理页列配置、中英文运行时语言包、宿主插件页 POM、新增 E2E 用例和 OpenSpec 反馈记录；未发现阻塞问题。变更不涉及 Go 生产代码、REST API、后端数据权限或缓存逻辑，Go 编译门禁不适用。
- [x] FB-3 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-3 前端 i18n/E2E 治理：`pnpm -C hack/tests test:validate`
- [x] FB-3 前端类型检查：`pnpm -C apps/lina-vben -F @lina/web-antd typecheck`
- [x] FB-3 E2E：`E2E_BROWSER_CHANNEL=chrome pnpm -C hack/tests exec playwright test /Users/john/Workspace/github/linaproai/linapro/hack/tests/e2e/extension/plugin/TC013-plugin-management-table-layout.ts --grep "TC-13a" --workers=1`
- [x] FB-3 i18n 影响：本次将插件管理列表描述列表头统一为插件专属“插件描述/Plugin Description”，已同步 `zh-CN` 和 `en-US` 前端运行时语言包；不涉及菜单、路由、manifest i18n 或 apidoc i18n。
- [x] FB-3 缓存一致性：本次仅调整前端表格列顺序、表头对齐和列宽，不新增或修改缓存逻辑，不影响插件 runtime freshness、启用状态或派生缓存失效机制。
- [x] FB-3 数据权限影响：本次不新增、修改或扩大任何数据操作接口；插件列表查询仍复用既有接口与权限边界。
- [x] FB-3 `/lina-review`：审查范围包含插件管理页列配置、中英文运行时语言包、宿主插件页 POM、插件管理列表布局 E2E 和 OpenSpec 反馈记录；未发现阻塞问题。变更不涉及 Go 生产代码、REST API、后端数据权限或缓存逻辑，Go 编译门禁不适用。
- [x] FB-4 后端 Go：`cd apps/lina-core && go test ./internal/service/apidoc ./internal/service/plugin/internal/openapi ./internal/service/plugin/internal/runtime ./internal/service/pluginhostservices ./pkg/pluginbridge/... ./pkg/pluginservice/contract -count=1`
- [x] FB-4 启动绑定：`cd apps/lina-core && go test ./internal/cmd -count=1`
- [x] FB-4 操作日志插件编译烟测：使用临时 `go.work` 包含 `apps/lina-core` 与 `apps/lina-plugins/linapro-monitor-operlog` 后运行 `GOWORK="$tmpdir/go.work" go test lina-plugin-linapro-monitor-operlog/backend/internal/service/middleware -run '^$' -count=1`
- [x] FB-4 动态插件 API 编译：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/api/... -count=1`、`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off GOOS=wasip1 GOARCH=wasm go test ./backend/api/dynamic/v1 -run '^$' -count=1`
- [x] FB-4 动态插件产物构建：`cd apps/lina-plugins && go run ../../hack/tools/linactl wasm p=linapro-demo-dynamic out=temp/output`
- [x] FB-4 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-4 静态验证：`python3 -m json.tool apps/lina-plugins/linapro-demo-dynamic/manifest/i18n/zh-CN/apidoc/plugin-api-main.json >/dev/null`、`rg -n "api_v1_|paths\\.(get|post|put|delete)\\.api_v1|/api/v1/extensions" apps/lina-core apps/lina-plugins/linapro-demo-dynamic hack/tools/linactl openspec/changes/rename-dynamic-plugin-route-prefix -g '!node_modules' -g '!dist' -g '!build'`；代码与动态插件 apidoc 资源无旧 `api_v1_*` key，`/api/v1/extensions` 仅存在于当前 OpenSpec 变更的迁移说明、负向规范和验证记录中。
- [x] FB-4 i18n 影响：本次迁移动态插件 apidoc i18n key contract，已同步 `linapro-demo-dynamic` 的 `zh-CN` apidoc JSON；`en-US` apidoc 继续为空占位并使用英文源文本；不新增前端运行时语言包或菜单翻译键。
- [x] FB-4 缓存一致性：本次仅改变 route contract 元数据、OpenAPI/apidoc key 生成与操作日志 routeDocKey 选择，不新增或修改缓存逻辑，不影响动态插件 runtime freshness、启用状态或派生缓存失效机制。
- [x] FB-4 数据权限影响：本次不新增、修改或扩大任何数据操作接口；动态插件路由仍复用既有登录鉴权、权限校验和 runtime dispatch 数据边界。
- [x] FB-4 `/lina-review`：审查范围包含动态插件 route contract `operationId`、wasm builder 提取、OpenAPI/apidoc key 生成、操作日志动态路由 doc key 选择、示例动态插件 apidoc 资源、单元测试和 OpenSpec 规范记录；未发现阻塞问题。
- [x] FB-5 后端 Go：`cd apps/lina-core && go test ./internal/service/apidoc ./internal/service/plugin/internal/openapi -count=1`
- [x] FB-5 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-5 静态验证：`rg -n "x-lina-apidoc-operation-key|openAPIOperationKeyExtension|x-i18n-key|openAPII18nKeyExtension|legacyOpenAPIOperationKeyExtension" apps/lina-core openspec/changes/rename-dynamic-plugin-route-prefix -S`；确认代码只使用 `x-i18n-key`，不再保留旧 `x-lina-apidoc-operation-key` 兼容读取、常量或测试。
- [x] FB-5 i18n 影响：本次仅泛化 OpenAPI extension 字段名，不新增、修改或删除 apidoc i18n JSON、前端运行时语言包、菜单翻译键或 manifest i18n 资源；现有 key base 仍为 `plugins.{pluginId}.operations.{operationId}`。
- [x] FB-5 缓存一致性：本次不新增或修改缓存读写、失效、刷新、运行时修订号或跨实例同步逻辑，不影响动态插件 runtime freshness 与启用状态缓存边界。
- [x] FB-5 数据权限影响：本次不新增、修改或扩大 REST API、插件数据面执行路径或业务数据访问，动态插件路由仍复用既有登录鉴权、权限校验和 runtime dispatch 边界。
- [x] FB-5 开发工具脚本影响：本次不新增或修改开发工具、脚本、构建入口或跨平台执行路径。
- [x] FB-5 `/lina-review`：审查范围包含动态插件 OpenAPI 投影扩展名、apidoc 本地化读取、单元测试和 OpenSpec 规范记录；未发现阻塞问题。生产投影与读取端均只使用 `x-i18n-key`，不保留旧 `x-lina-apidoc-operation-key` 兼容分支。
- [x] FB-6 后端 Go：`cd apps/lina-core && go test ./internal/service/apidoc ./internal/service/plugin/internal/openapi -count=1`
- [x] FB-6 插件服务编译烟测：`cd apps/lina-core && go test ./internal/service/plugin -run '^$' -count=1`
- [x] FB-6 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-6 静态验证：`rg "openAPII18nKeyExtension|OpenAPII18nKeyExtension|x-i18n-key" apps/lina-core/internal/service/apidoc apps/lina-core/internal/service/plugin/internal/openapi -n`；确认 `x-i18n-key` 仅在 `apidoc.OpenAPII18nKeyExtension` 中定义，动态插件 OpenAPI 投影只引用 apidoc 包导出的定义。
- [x] FB-6 i18n 影响：本次仅复用 OpenAPI extension 字段名常量，不新增、修改或删除 apidoc i18n JSON、前端运行时语言包、菜单翻译键或 manifest i18n 资源；现有 key base 仍为 `plugins.{pluginId}.operations.{operationId}`。
- [x] FB-6 缓存一致性：本次不新增或修改缓存读写、失效、刷新、运行时修订号或跨实例同步逻辑，不影响动态插件 runtime freshness 与启用状态缓存边界。
- [x] FB-6 数据权限影响：本次不新增、修改或扩大 REST API、插件数据面执行路径或业务数据访问，动态插件路由仍复用既有登录鉴权、权限校验和 runtime dispatch 边界。
- [x] FB-6 开发工具脚本影响：本次不新增或修改开发工具、脚本、构建入口或跨平台执行路径。
- [x] FB-6 `/lina-review`：审查范围包含 apidoc OpenAPI extension 常量导出、动态插件 OpenAPI 投影引用、相关单元测试和 OpenSpec 反馈记录；未发现阻塞问题。`apidoc` 包不依赖 `plugin` 包，`plugin/internal/openapi` 引用 `apidoc.OpenAPII18nKeyExtension` 已通过包测试和 `plugin` 编译烟测验证无 import cycle。
- [x] FB-7 后端 Go：`cd apps/lina-core && go test ./internal/service/apidoc ./internal/service/plugin/internal/openapi -count=1`
- [x] FB-7 启动绑定：`cd apps/lina-core && go test ./internal/cmd -count=1`
- [x] FB-7 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-7 静态验证：`rg -n "OpenAPII18nKeyExtension|x-i18n-key|x-lina-apidoc-operation-key|operationExtensionString|buildRouteOpenAPIDocOperationKey" apps/lina-core -S`；确认核心代码不再定义、输出或读取动态插件 OpenAPI i18n 扩展字段。
- [x] FB-7 i18n 影响：本次仅调整动态插件 apidoc key 生成机制，key base 仍为 `plugins.{pluginId}.operations.{operationId}` 或路径兜底；不新增、修改或删除 apidoc i18n JSON、前端运行时语言包、菜单翻译键或 manifest i18n 资源。
- [x] FB-7 缓存一致性：本次不新增或修改缓存读写、失效、刷新、运行时修订号或跨实例同步逻辑，不影响动态插件 runtime freshness 与启用状态缓存边界。
- [x] FB-7 数据权限影响：本次不新增、修改或扩大 REST API、插件数据面执行路径或业务数据访问，动态插件路由仍复用既有登录鉴权、权限校验和 runtime dispatch 边界。
- [x] FB-7 开发工具脚本影响：本次不新增或修改开发工具、脚本、构建入口或跨平台执行路径。
- [x] FB-7 `/lina-review`：审查范围包含动态插件 OpenAPI 投影、apidoc 本地化 key 选择、相关单元测试和 OpenSpec 规范记录；未发现阻塞问题。动态插件 OpenAPI operation 不再输出 `x-i18n-key`，apidoc 本地化从 `/x/{pluginId}/...` 路径解析插件标识并从 OpenAPI `operationId` 剥离插件前缀生成 canonical operation key，解析失败时保留路径兜底。
- [x] FB-8 后端 Go：`cd apps/lina-core && go test ./pkg/pluginbridge/... ./pkg/pluginhost ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/openapi ./internal/controller/plugin -count=1`
- [x] FB-8 启动绑定：`cd apps/lina-core && go test ./internal/cmd -count=1`
- [x] FB-8 源码插件编译烟测：使用临时 `go.work` 包含 `apps/lina-core` 与 8 个内建源码插件后运行 `GOWORK="$tmpdir/go.work" go test lina-plugin-linapro-content-notice/backend lina-plugin-linapro-demo-source/backend lina-plugin-linapro-org-core/backend lina-plugin-linapro-tenant-core/backend lina-plugin-linapro-monitor-server/backend lina-plugin-linapro-monitor-online/backend lina-plugin-linapro-monitor-loginlog/backend lina-plugin-linapro-monitor-operlog/backend -run '^$' -count=1`
- [x] FB-8 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-8 静态验证：`rg -n 'dynamic route path must start with /api/v1|must start with /api/v1|PluginAPIVersionPath|PluginAPIVersionPrefix|must include the /api/v1|必须.*api/v1.*段|fixed.*x/\{pluginId\}/api/v1|固定以 /x/\{pluginId\}/api/v1' apps/lina-core apps/lina-plugins openspec/changes/rename-dynamic-plugin-route-prefix -g'*.go' -g'*.md' -g'*.json' -g'*.yaml'`；确认无实现或文档继续把 `/api/v1` 描述为插件路由强制前缀。
- [x] FB-8 i18n 影响：本次更新动态插件路由公开路径相关 API 文档源文本、核心 `zh-CN` apidoc JSON、packed apidoc JSON、动态示例插件 `zh-CN` apidoc JSON 和中英文 README；不新增前端运行时菜单或页面翻译键。
- [x] FB-8 缓存一致性：本次不新增或修改缓存读写、失效、刷新、运行时修订号或跨实例同步逻辑；`/x/{pluginId}` 路由仍经过动态插件 runtime freshness、启用状态、菜单可见性和权限治理链路。
- [x] FB-8 数据权限影响：本次不新增、修改或扩大业务数据操作接口；动态插件和源码插件路由仍复用既有登录鉴权、权限校验、租户上下文和插件启用状态边界。
- [x] FB-8 开发工具脚本影响：本次不新增或修改开发工具、脚本、构建入口或跨平台执行路径。
- [x] FB-8 `/lina-review`：审查范围包含动态插件 contract 校验、runtime `/x/{pluginId}` 路由解析、OpenAPI/route review 投影、源码插件 `APIPrefix()` 语义、内建源码插件注册路径、动态示例插件文档与 apidoc 文案、相关单元测试和 OpenSpec 记录；未发现阻塞问题。`/api/v1` 仅保留为示例和内建源码插件主动选择的插件自有路径段，不再作为宿主强制前缀。
- [x] FB-9 后端 Go：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/api/... -count=1`
- [x] FB-9 WASM API 编译烟测：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off GOOS=wasip1 GOARCH=wasm go test ./backend/api/dynamic/v1 -run '^$' -count=1`
- [x] FB-9 linactl 构建器单元测试：`go test ./hack/tools/linactl/internal/wasmbuilder -run 'TestBuildRuntimeWasmArtifactFromSourceEmbedsDeclaredAssets|TestBuildRuntimeWasmArtifactFromSourceSkipsHiddenEmbeddedDirectoryEntries' -count=1`、`go test ./hack/tools/linactl/internal/wasmbuilder -count=1`
- [x] FB-9 开发工具门禁：`go test ./hack/tools/linactl/... -count=1`
- [x] FB-9 动态插件产物构建：`cd apps/lina-plugins && go run ../../hack/tools/linactl wasm p=linapro-demo-dynamic out=temp/output`
- [x] FB-9 动态路由回归：`cd apps/lina-core && go test ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/openapi ./internal/controller/plugin ./pkg/pluginbridge/... -count=1`
- [x] FB-9 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-9 静态验证：`rg -n 'path:"/api/v1|/api/v1/extensions|api_v1_|paths\.(get|post|put|delete)\.api_v1|route convention|路由约定' apps/lina-plugins/linapro-demo-dynamic hack/tools/linactl/internal/wasmbuilder openspec/changes/rename-dynamic-plugin-route-prefix -g '!node_modules' -g '!dist' -g '!build'`；代码、示例插件 DTO 和动态插件 apidoc 资源中无 DTO 级 `/api/v1` path 前缀或旧 `api_v1_*` key，`/api/v1/extensions` 仅存在于当前 OpenSpec 变更的迁移说明和负向规范中。
- [x] FB-9 i18n 影响：本次同步动态示例插件 API 文档源文本、`zh-CN` apidoc JSON 和中英文 README，将示例的 `/api/v1` 描述为路由分组；不新增前端运行时菜单或页面翻译键，`en-US` apidoc 继续使用英文源文本。
- [x] FB-9 缓存一致性：本次不新增或修改缓存读写、失效、刷新、运行时修订号或跨实例同步逻辑；最终动态路由契约和公开路径仍通过 `/x/{pluginId}`、runtime freshness、启用状态和权限治理链路执行。
- [x] FB-9 数据权限影响：本次不新增、修改或扩大业务数据操作接口；动态插件 demo 记录接口的权限、租户上下文和既有数据访问边界不变。
- [x] FB-9 开发工具脚本影响：本次修改 `linactl` Go 构建器，不新增 shell/PowerShell 脚本；已通过 `go test ./hack/tools/linactl/... -count=1` 和实际 `linactl wasm` 构建验证跨平台 Go 工具入口。
- [x] FB-9 `/lina-review`：审查范围包含动态插件 DTO 路径、构建期路由分组组合、构建器测试、示例 apidoc/README 文案、OpenSpec 规范与任务记录；未发现阻塞问题。DTO 只声明资源本地路径，构建器将插件自有路由分组前缀组合进 route contract，公开路径保持 `/x/{pluginId}/api/v1/...`。
- [x] FB-10 后端 Go：`cd apps/lina-core && go test ./internal/service/apidoc ./internal/service/plugin/internal/openapi ./internal/service/plugin/internal/runtime ./internal/service/pluginhostservices ./pkg/pluginbridge/... ./pkg/pluginservice/contract -count=1`
- [x] FB-10 启动绑定：`cd apps/lina-core && go test ./internal/cmd -count=1`
- [x] FB-10 插件服务编译烟测：`cd apps/lina-core && go test ./internal/service/plugin -run '^$' -count=1`
- [x] FB-10 动态示例插件 API 编译：`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/api/... -count=1`
- [x] FB-10 linactl 构建器：`cd hack/tools/linactl && go test ./internal/wasmbuilder -count=1`
- [x] FB-10 操作日志插件编译烟测：使用临时 `go.work` 包含 `apps/lina-core` 与 `apps/lina-plugins/linapro-monitor-operlog` 后运行 `GOWORK="$tmpdir/go.work" go test lina-plugin-linapro-monitor-operlog/backend/internal/service/middleware -run '^$' -count=1`
- [x] FB-10 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-10 静态验证：`python3 -m json.tool apps/lina-plugins/linapro-demo-dynamic/manifest/i18n/zh-CN/apidoc/plugin-api-main.json >/dev/null`、`rg -n '\bOperationID\b|operationId:\"|plugins\.linapro_demo_dynamic\.operations|\"operations\"' apps/lina-core apps/lina-plugins/linapro-demo-dynamic apps/lina-plugins/linapro-monitor-operlog hack/tools/linactl/internal/wasmbuilder -S`；确认动态插件 route contract、示例 DTO、示例 apidoc JSON 和操作日志路径不再使用 operation identity，剩余 `OperationID` 仅为静态 OpenAPI 字段或测试断言动态路由不会输出 operationId。
- [x] FB-10 i18n 影响：本次将动态插件 apidoc key 收敛为 `plugins.{pluginId}.paths.{method}.{dottedRoutePath}`，已同步 `linapro-demo-dynamic` 的 `zh-CN` apidoc JSON；`en-US` apidoc 继续为空占位并使用英文源文本；不新增前端运行时菜单或页面翻译键。
- [x] FB-10 缓存一致性：本次不新增或修改缓存读写、失效、刷新、运行时修订号或跨实例同步逻辑；动态插件路由仍通过 `/x/{pluginId}`、runtime freshness、启用状态和权限治理链路执行。
- [x] FB-10 数据权限影响：本次不新增、修改或扩大 REST API、插件数据面执行路径或业务数据访问；动态插件路由仍复用既有登录鉴权、权限校验和 runtime dispatch 边界。
- [x] FB-10 开发工具脚本影响：本次修改 `linactl` Go 构建器的 route contract 提取，不新增 shell/PowerShell 脚本；已通过构建器包测试验证。
- [x] FB-10 `/lina-review`：审查范围包含动态插件 route contract、wasm builder route 提取、OpenAPI 投影、apidoc 本地化 key 选择、操作日志动态 routeDocKey、示例插件 DTO/apidoc JSON、单元测试和 OpenSpec 规范记录；未发现阻塞问题。动态插件翻译身份最终收敛为已有唯一契约 `method + public path`，不再维护 `operationId` 或 `RequestType` 第二身份。
- [x] FB-11 后端 Go：`go test ./apps/lina-core/pkg/pluginbridge/... -count=1`、`cd apps/lina-plugins/linapro-demo-dynamic && GOWORK=off go test ./backend/api/... -count=1`
- [x] FB-11 动态插件后端编译烟测：使用临时 `go.work` 包含 `apps/lina-core` 与 `apps/lina-plugins/linapro-demo-dynamic` 后运行 `GOWORK="$tmpdir/go.work" go test lina-plugin-linapro-demo-dynamic/backend -run '^$' -count=1`
- [x] FB-11 linactl 构建器：`go test ./hack/tools/linactl/internal/wasmbuilder -count=1`
- [x] FB-11 开发工具门禁：`go test ./hack/tools/linactl/... -count=1`
- [x] FB-11 动态插件产物构建：`cd apps/lina-plugins && go run ../../hack/tools/linactl wasm p=linapro-demo-dynamic out=temp/output`
- [x] FB-11 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-11 静态验证：`rg -n 'RouteGroupPrefix|path:"/api/v1|api_v1_|paths\.(get|post|put|delete)\.api_v1' apps/lina-plugins/linapro-demo-dynamic/backend/api apps/lina-plugins/linapro-demo-dynamic/manifest/i18n hack/tools/linactl/internal/wasmbuilder -S`；确认示例插件生成 API 接口文件和 DTO 不再声明路由分组前缀或 DTO 级 `/api/v1` path，剩余命中仅为构建器内部 group-prefix helper 命名和防回归测试断言。
- [x] FB-11 i18n 影响：本次仅调整动态插件后端路由分组声明位置和中英文 README 说明，不新增、修改或删除运行时前端语言包、manifest i18n 或 apidoc i18n JSON；API 文档源文本不变。
- [x] FB-11 缓存一致性：本次不新增或修改缓存读写、失效、刷新、运行时修订号或跨实例同步逻辑；动态插件最终 route contract 和公开路径仍通过 `/x/{pluginId}`、runtime freshness、启用状态和权限治理链路执行。
- [x] FB-11 数据权限影响：本次不新增、修改或扩大业务数据操作接口；动态插件 demo 记录接口的权限、租户上下文和既有数据访问边界不变。
- [x] FB-11 开发工具脚本影响：本次修改 `linactl` Go 构建器，不新增 shell/PowerShell 脚本；已通过 `go test ./hack/tools/linactl/... -count=1` 和实际 `linactl wasm` 构建验证跨平台 Go 工具入口。
- [x] FB-11 `/lina-review`：审查范围包含动态插件 `RegisterRoutes` 注册声明、`pluginbridge.DynamicRouteRegistrar` 契约、`linactl` 构建器 route group 提取、示例插件 API 约束测试、README 和 OpenSpec 记录；未发现阻塞问题。生成的 `backend/api/dynamic/dynamic.go` 不再被手工修改，动态插件路由分组职责回到后端注册入口，新增 `/api/v2` 或 `/interface/m1` 时应新增独立 API 包并在 `RegisterRoutes` 中绑定对应分组。
- [x] FB-12 后端 Go：使用临时 `go.work` 包含 `apps/lina-core` 与 `apps/lina-plugins/linapro-demo-dynamic` 后运行 `GOWORK="$tmpdir/go.work" go test lina-plugin-linapro-demo-dynamic/backend -run '^$' -count=1`
- [x] FB-12 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-12 i18n 影响：本次仅补充 Go 源码注释，不新增、修改或删除用户可见页面文案、运行时语言包、manifest i18n 或 apidoc i18n JSON。
- [x] FB-12 缓存一致性：本次仅补充注释，不新增或修改缓存读写、失效、刷新、运行时修订号或跨实例同步逻辑。
- [x] FB-12 数据权限影响：本次仅补充注释，不新增、修改或扩大业务数据操作接口，动态插件 demo 记录接口的数据权限边界不变。
- [x] FB-12 开发工具脚本影响：本次不新增或修改开发工具、脚本、构建入口或跨平台执行路径。
- [x] FB-12 `/lina-review`：审查范围限定为动态插件 `RegisterRoutes` 注释和 OpenSpec 反馈记录；未发现阻塞问题。变更仅解释构建期约定入口和 `registrar.Group(dynamicAPIV1GroupPrefix, "dynamic/v1")` 的参数含义，不改变路由契约生成、运行时请求分发、i18n 资源、缓存或数据权限行为。
- [x] FB-13 linactl 精准回归：`go test ./hack/tools/linactl -run TestRunWasmResolvesExplicitRelativeOutputFromRepositoryRoot -count=1`
- [x] FB-13 linactl 构建器：`go test ./hack/tools/linactl/internal/wasmbuilder -count=1`
- [x] FB-13 开发工具门禁：`go test ./hack/tools/linactl/... -count=1`
- [x] FB-13 动态插件产物构建：`cd apps/lina-plugins && go run ../../hack/tools/linactl wasm p=linapro-demo-dynamic out=temp/output`；确认输出为仓库根 `temp/output/linapro-demo-dynamic.wasm`，并清理旧的 `apps/lina-plugins/temp/output` 遗留产物。
- [x] FB-13 OpenSpec：`openspec validate rename-dynamic-plugin-route-prefix --strict`
- [x] FB-13 i18n 影响：本次仅更新 `linactl` 双语 README 中 `out` 参数说明，不新增、修改或删除用户可见运行时文案、前端语言包、manifest i18n 或 apidoc i18n JSON。
- [x] FB-13 缓存一致性：本次仅调整开发工具产物路径解析和文档说明，不新增或修改运行时缓存、失效、刷新、修订号或跨实例同步逻辑。
- [x] FB-13 数据权限影响：本次不新增、修改或扩大业务数据操作接口，不影响动态插件 demo 记录接口的数据权限边界。
- [x] FB-13 开发工具脚本影响：本次修改 `linactl` Go 工具，不新增 shell/PowerShell 脚本；已通过 `go test ./hack/tools/linactl/... -count=1` 和实际 `linactl wasm` 构建验证跨平台 Go 工具入口。
- [x] FB-13 `/lina-review`：审查范围包含 `linactl wasm` 输出目录解析、回归单测、双语 README 参数说明和 OpenSpec 反馈记录；未发现阻塞问题。`git status --short --ignored --ignore-submodules=none` 中存在无关的 `apps/lina-core/internal/cmd/*` 与 `decouple-workspace-plugin-routes/tasks.md` 变更，未纳入本次审查。本次不涉及运行时后端服务、REST API、前端运行时文案、manifest i18n、apidoc i18n、缓存或数据权限逻辑。
