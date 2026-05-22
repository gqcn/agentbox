## 1. 配置与契约准备

- [x] 1.1 将默认管理工作台入口路径确定为 `/admin`，并在配置设计中明确工作台入口不得配置为根路径 `/`。
- [x] 1.2 在后端配置结构、默认配置、配置模板和配置元数据中新增工作台入口路径配置及校验规则。
- [x] 1.3 更新 public frontend 配置投影或等价前端启动配置，使前端可以读取管理工作台入口路径；同时明确工作台 `basePath` 以启动期配置为权威，不作为运行期热修改配置。
- [x] 1.4 明确本变更不新增插件 HTTP 路由 manifest 声明，不新增 `public_routes`、`portal_routes` 或工作台 API 路由分组字段；同时在 `plugin.yaml` 增加可选 `public_assets` 严格 schema 声明边界，用于源码插件和动态插件通过 `/x-assets/{plugin-id}/{version}/...` 托管可公开静态资源，且不保留 `/plugin-assets` 兼容入口。
- [x] 1.5 明确插件 API base 契约：宿主控制面 API 使用 `/api/v1`，源码插件和动态插件 API 统一使用 `/x/{plugin-id}/api/v1`；`/x` 表示统一插件 API 命名空间，不再表示动态插件专属数据面，且 `/x/**` 不允许承载插件非 API 公开路由。
- [x] 1.6 记录 i18n 影响判断；若新增用户可见配置项、接口文档或页面文案，则同步运行时 i18n 与 apidoc i18n 资源。
- [x] 1.7 记录缓存一致性影响判断；确认本变更不新增业务缓存，且工作台 fallback 与源码插件路由注册不依赖跨节点可变状态；public asset URL 以插件版本为缓存边界，同一版本资源内容必须不可变。

## 2. 后端路由边界

- [x] 2.1 调整后端前端静态资源处理器，使默认管理工作台静态资源与 SPA fallback 仅匹配配置后的工作台入口路径。
- [x] 2.2 保护宿主已注册控制面 API 路由、统一插件 API 命名空间 `/x`、统一插件资产命名空间 `/x-assets` 和工作台入口路径；源码插件 API 必须注册到 `/x/{plugin-id}/api/v1`，源码插件公开页面、门户、静态资源或自管 fallback 仍可注册其他非保留路径但不得注册到 `/x`。
- [x] 2.3 保持源码插件 HTTP registrar 代码注册模型，不要求插件在 `plugin.yaml` 中声明 HTTP 路由或路由分组语义；新增 `APIPrefix()` 或等价方法返回 `/x/{plugin-id}/api/v1`。
- [x] 2.4 增加后端测试覆盖：工作台 basePath 可访问、非工作台根路径不返回管理台 SPA、源码插件根路径可注册并命中插件 handler。
- [x] 2.5 增加后端测试覆盖：源码插件与宿主保留路径或其他插件发生 GoFrame 路由冲突时启动或注册失败。
- [x] 2.6 增加插件 public asset resolver 和校验测试，覆盖源码插件 embedded FS 与动态插件 artifact 只暴露 `plugin.yaml public_assets` 声明的资源，动态插件既有 frontend assets 迁移到 `/x-assets/{plugin-id}/{version}/...` 声明式模型后才能访问，非法 schema 阻止插件校验/安装/启用，插件禁用或租户不可用时返回 404，声明路径不得通过 `../`、绝对路径、URL 或符号链接逃逸插件自身资源边界，旧 `/plugin-assets` 入口不可访问。
- [x] 2.7 调整动态插件 route contract 示例、OpenAPI 投影和路由分发测试，使动态插件 API 公开路径使用 `/x/{plugin-id}/api/v1/...`，并确认动态插件 `/x` 分发不会吞掉源码插件 `/x/{source-plugin-id}/api/v1/...` 具体路由。
- [x] 2.8 若修改 Go 生产代码，运行覆盖变更包的 `go test`，并对涉及启动路由绑定的变更运行 `cd apps/lina-core && go test ./internal/cmd -count=1` 或更窄等价测试。

## 3. 前端工作台入口

- [x] 3.1 调整管理工作台 Vue Router base、登录跳转、刷新路径和默认首页解析，使其使用配置后的工作台入口路径。
- [x] 3.2 调整 public frontend 配置同步逻辑，确保在非根路径工作台入口下仍能正确请求 `/api/v1/config/public/frontend`。
- [x] 3.3 增加或调整插件前端 API base helper，区分宿主 `/api/v1` 和统一插件 API `/x/{plugin-id}/api/v1`。
- [x] 3.4 调整开发代理和构建配置，使本地开发以后端公开地址作为统一入口，并由后端在工作台入口路径下代理或转发到 Vite；同时不吞掉后端源码插件根路由或 `/x/*` 统一插件 API 请求。
- [x] 3.5 更新前端单元测试，覆盖工作台入口路径规范化、登录重定向、插件 API base helper 和刷新行为。
- [x] 3.6 运行相关前端单元测试和类型检查，确认工作台入口路径变更不破坏现有管理台页面。

## 4. 插件与菜单边界

- [x] 4.1 确认现有源码插件 `plugin.yaml` 的 `menus` 仍作为工作台菜单与权限事实来源，不新增 HTTP 路由声明字段。
- [x] 4.2 增加或调整源码插件测试夹具，验证插件可通过代码注册公开 HTTP 路由并自行返回内容，同时可通过 `APIPrefix()` 注册 `/x/{plugin-id}/api/v1/...` API。
- [x] 4.3 验证插件 HTTP 路由不会自动生成菜单、权限节点、OpenAPI 文档或插件资源引用。
- [x] 4.4 验证工作台菜单声明不会让主框架推断同路径 HTTP handler 存在。
- [x] 4.5 为源码插件和动态插件补充 `plugin.yaml public_assets` 声明示例或夹具，验证声明式托管不影响插件自行注册 HTTP 静态资源路由。
- [x] 4.6 迁移官方源码插件和动态插件示例的 API 路径文档、前端调用与测试夹具，源码插件和动态插件示例均使用 `/x/{plugin-id}/api/v1/...`。
- [x] 4.7 迁移动态插件 hosted frontend 菜单，继续使用 `system/plugin/dynamic-page` 或等价工作台承载组件，并通过菜单 query 或等价参数引用 `/x-assets/{plugin-id}/{version}/...`；不得把 `/x-assets/.../index.html` 直接作为管理工作台菜单路由替代动态页组件。

## 5. E2E 与文档

- [x] 5.1 按 lina-e2e 规范新增或更新 E2E，用配置后的工作台入口访问登录页和管理台页面。
- [x] 5.2 新增或更新 E2E，验证源码插件公开根路由或非保留公开路径不被管理台 SPA fallback 吞掉，并验证源码插件统一插件 API 路径 `/x/{plugin-id}/api/v1/...` 可用。
- [x] 5.3 新增或更新 E2E，验证动态插件 API 路径 `/x/{plugin-id}/api/v1/...` 可用，旧式 `/x/{plugin-id}/{resource}` 示例和断言已迁移。
- [x] 5.4 更新 README 或开发文档中的默认管理工作台访问路径；若新增目录说明文档，同步英文 `README.md` 与中文 `README.zh-CN.md`。
- [x] 5.5 更新插件开发文档，说明源码插件 HTTP 路由是插件实现细节，源码插件和动态插件 API 均必须使用 `/x/{plugin-id}/api/v1/...`，`/x` 下不得注册非 API 路由，工作台权限入口仍通过 `plugin.yaml` 菜单声明，并说明 `plugin.yaml public_assets` 声明只用于 `/x-assets/{plugin-id}/{version}/...` 托管可公开静态资源，动态插件既有 frontend assets 访问路径也需迁移到该声明式模型并继续由动态页组件承载，且不提供 `/plugin-assets` 兼容路径。

## 6. 验证与审查

- [x] 6.1 运行 `openspec validate decouple-workspace-plugin-routes --strict`。
- [x] 6.2 运行后端路由绑定、配置校验、插件注册、`APIPrefix()`、源码插件 `/x/{plugin-id}/api/v1` 和动态插件 `/x/{plugin-id}/api/v1` 相关 Go 测试。
- [x] 6.3 运行前端工作台入口、插件 API base helper 相关单元测试和必要的 E2E 回归。
- [x] 6.4 运行静态检查，确认未新增将源码插件 HTTP 路由写入 `plugin.yaml` 的治理字段或文档要求，确认 `public_assets` 声明文档以插件自身资源边界和路径安全校验为准，确认文档不再使用旧源码插件 API 前缀，确认动态插件示例不再使用旧式 `/x/{plugin-id}/{resource}` API 路径，确认 `/x` 下非 API 路由未被作为允许示例，并确认 public asset 文档不再把 `/plugin-assets` 作为可用入口。
- [x] 6.5 执行 `/lina-review` 审查，重点检查主框架边界、插件 API base、工作台菜单治理、i18n 判断、缓存一致性判断、Go 编译门禁和 E2E 覆盖。

## Feedback

- [x] **FB-1**: 需求澄清为全新项目且不考虑根路径管理工作台兼容性，更新提案、设计、增量规范和任务清单，移除 `/` 兼容与回滚表述。
- [x] **FB-2**: 需求澄清为源码插件和动态插件均可声明 public assets，由宿主通过声明式资产入口托管可公开静态资源；保留源码插件 HTTP 路由代码注册事实来源，并明确 SQL、manifest、i18n、apidoc 与后端资源不得被公开。
- [x] **FB-3**: 需求澄清为源码插件公开页面、门户、静态资源和自管 fallback 可继续自由注册非保留 HTTP 路由；public assets 声明放在 `plugin.yaml public_assets`，并明确动态插件既有 frontend assets 访问路径迁移到该声明式模型。
- [x] **FB-4**: 需求澄清为源码插件 API 前缀需要从建议性描述升级为规范性契约；源码插件仍保留非 API 公开 HTTP 路由的自由注册能力。
- [x] **FB-5**: 需求澄清为不考虑兼容性，源码插件和动态插件 API 统一使用 `/x/{plugin-id}/api/v1` 插件 API 命名空间；`/x` 不再作为动态插件专属含义，源码插件非 API 公开 HTTP 路由仍保留自由注册能力。
- [x] **FB-6**: 需求澄清为不考虑 `/plugin-assets` 兼容入口，插件 public assets 统一改为 `/x-assets/{plugin-id}/{version}/...`，保留插件版本作为静态资源缓存和升级隔离边界。
- [x] **FB-7**: 根据需求澄清完善官网能力边界文档，说明 `/admin` 工作台入口、统一插件 API 命名空间 `/x/{plugin-id}/api/v1`、`public_assets` 与 `/x-assets/{plugin-id}/{version}/...` 的边界，并同步英文镜像。
- [x] **FB-8**: 需求澄清为开发模式以后端公开地址作为统一入口并由后端代理 Vite；工作台 `basePath` 为启动期权威配置；`/x/**` 只承载插件 API；源码插件 OpenAPI 只投影插件 API；`public_assets` 使用严格 schema、默认匿名可读但受插件启用与租户可用性约束；动态插件页面继续通过 `system/plugin/dynamic-page` 引用 `/x-assets` 资源。
- [x] **FB-9**: 需求澄清为 `plugin.yaml` 静态资源声明字段必须使用 `public_assets`，替代 camelCase 旧写法，以保持配置项命名规则一致。
- [x] **FB-10**: 根据需求澄清改进官网能力边界文档中“门户路由与管理路由的区分”示例代码，明确公开门户入口使用非保留路径，插件 API 使用 `/x/{plugin-id}/api/v1`，工作台入口仍由 `plugin.yaml` 的 `menus` 管理，并同步英文镜像。
- [x] **FB-11**: 将 `public_assets` 目录默认文件从硬编码 `index.html` 调整为每个声明项的可选 `index` 配置；未配置时仍默认 `index.html`，并校验 `index` 为安全相对文件名。
- [x] **FB-12**: 将插件宿主路由、托管资产和动态页菜单的重复硬编码契约收敛到 `pkg/pluginhost` 主文件维护，避免 controller、cmd 和插件前端服务重复定义同一宿主插件边界。
- [x] **FB-13**: 需求澄清为 `public_assets` 完全信任插件作者对插件内资源的显式发布声明；宿主只做路径安全、声明存在性、声明目录读取和插件资源边界校验，不再维护治理目录黑名单或前端目录白名单。
- [x] **FB-14**: 参考 `linapro-content-notice` 源码插件的路由注册编码风格，统一其他官方源码插件的路由注册代码结构。
- [x] **FB-15**: 修复已有根通配路由 `/*` 存在时工作台 `/admin` 静态资源 fallback 被抢占并返回 404 的单元测试失败。

## Verification Notes

- `cd apps/lina-core && go test ./internal/service/plugin -count=1` 当前受本地数据库缺少 `plugin_linapro_tenant_core_user_membership` 表阻断；已运行变更相关定向测试和子包测试覆盖插件路由、public assets、动态插件运行时升级回滚与启动路由绑定。
- `pnpm -C hack/tests test:validate` 当前受多个既有官方源码插件四位 TC 文件名阻断；本次新增/调整的 `linapro-demo-source` 插件 E2E 已按 `TC001`/`TC002` 命名并串行运行通过。
- FB-11 已验证：`cd apps/lina-core && go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/frontend ./internal/service/plugin/internal/runtime -count=1`、`cd apps/lina-core && go test ./internal/cmd -count=1`、`cd hack/tools/linactl && go test ./internal/wasmbuilder -count=1`、`openspec validate decouple-workspace-plugin-routes --strict`、`git diff --check`、`git -C apps/lina-plugins diff --check`、`rg -n "publicAssets" openspec apps hack -g '!node_modules' -g '!dist' -g '!*.map'` 均通过或无命中。
- FB-12 为项目治理类反馈，不新增用户可见文案、接口语义、缓存或运行时行为；i18n 与缓存资源无需变更。已验证：`cd apps/lina-core && go test ./pkg/pluginhost ./internal/controller/user ./internal/controller/menu ./internal/service/plugin/internal/frontend ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/openapi ./internal/service/apidoc ./internal/cmd -count=1`、`openspec validate decouple-workspace-plugin-routes --strict`、`git diff --check -- apps/lina-core/pkg/pluginhost/pluginhost.go apps/lina-core/internal/controller/user/user_v1_get_info.go apps/lina-core/internal/controller/user/user_v1_get_info_test.go apps/lina-core/internal/controller/menu/menu_v1_all.go apps/lina-core/internal/service/plugin/internal/frontend/contract.go apps/lina-core/internal/service/plugin/internal/frontend/contract_test.go apps/lina-core/internal/service/plugin/internal/frontend/frontend_runtime.go apps/lina-core/internal/cmd/cmd_http_frontend.go apps/lina-core/pkg/pluginhost/pluginhost_route_registrar.go apps/lina-core/internal/service/plugin/internal/runtime/route.go apps/lina-core/internal/service/plugin/internal/runtime/route_dispatch_test.go apps/lina-core/internal/service/plugin/internal/runtime/route_test.go apps/lina-core/internal/service/plugin/internal/openapi/openapi.go apps/lina-core/internal/service/plugin/internal/openapi/openapi_paths.go apps/lina-core/internal/service/apidoc/apidoc_i18n.go openspec/changes/decouple-workspace-plugin-routes/tasks.md`。
- FB-13 不新增用户可见运行时文案、接口字段、数据访问接口或业务缓存；i18n、数据权限和缓存资源无需变更。已验证：`cd apps/lina-core && go test ./pkg/pluginfs ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/frontend -count=1`、`openspec validate decouple-workspace-plugin-routes --strict`、`git diff --check -- apps/lina-core/pkg/pluginfs/pluginfs.go apps/lina-core/internal/service/plugin/internal/catalog/manifest_validate.go apps/lina-core/internal/service/plugin/internal/catalog/catalog_test.go apps/lina-core/internal/service/plugin/internal/frontend/frontend_runtime.go apps/lina-core/internal/service/plugin/internal/frontend/public_asset_test.go openspec/changes/decouple-workspace-plugin-routes/proposal.md openspec/changes/decouple-workspace-plugin-routes/design.md openspec/changes/decouple-workspace-plugin-routes/specs/plugin-ui-integration/spec.md openspec/changes/decouple-workspace-plugin-routes/tasks.md apps/lina-plugins/README.md apps/lina-plugins/README.zh-CN.md`。
- FB-14 为源码插件路由注册代码风格治理，未新增用户可见文案、接口字段、数据操作接口、业务缓存、开发工具或脚本；i18n、数据权限和缓存资源无需变更。已验证：使用临时 `go.work` 包含 `apps/lina-core` 与相关源码插件后运行 `GOWORK="$tmpdir/go.work" go test lina-plugin-linapro-content-notice/backend lina-plugin-linapro-demo-source/backend lina-plugin-linapro-monitor-loginlog/backend lina-plugin-linapro-monitor-online/backend lina-plugin-linapro-monitor-operlog/backend lina-plugin-linapro-monitor-server/backend lina-plugin-linapro-org-core/backend lina-plugin-linapro-tenant-core/backend -run '^$' -count=1`、`cd apps/lina-plugins && GOWORK=off go test ./... -run '^$' -count=1`、`cd apps/lina-core && go test ./pkg/pluginhost -count=1`、`cd apps/lina-core && go test ./internal/cmd -count=1`、`openspec validate decouple-workspace-plugin-routes --strict`。
- FB-14 `/lina-review`：审查范围限定为官方源码插件 `backend/plugin.go` 路由注册结构与本任务记录；未发现阻塞问题。
- FB-15 为后端工作台与插件公开资产静态路由绑定修复，不新增用户可见文案、接口字段、数据操作接口、业务缓存、开发工具或脚本；i18n、数据权限和缓存资源无需变更。已验证：`cd apps/lina-core && go test ./internal/cmd -run '^(TestFrontendAssetFallbackIsScopedToWorkspaceBasePath|TestFrontendAssetFallbackClaimsHostedPluginAssetNamespace)$' -count=1 -v`、`cd apps/lina-core && go test ./internal/cmd -count=1`、`openspec validate decouple-workspace-plugin-routes --strict`、`git diff --check -- apps/lina-core/internal/cmd/cmd_http_frontend.go apps/lina-core/internal/cmd/cmd_http_test.go openspec/changes/decouple-workspace-plugin-routes/tasks.md`。
