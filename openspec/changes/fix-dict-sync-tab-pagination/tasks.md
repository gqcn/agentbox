## Feedback

- [x] **FB-1**: 字典标签文字更新后已打开列表未同步显示最新标签
- [x] **FB-2**: 切换工作台 Tab 后前一个列表页分页状态丢失
- [x] **FB-3**: `plugin-full` Go 单测在动态插件样例 manifest 资源断言中失败
- [x] **FB-4**: 动态插件样例不应将运行配置契约从 `config.yaml` 改为 `config.example.yaml`

### FB-1 根因

字典 Store 的 `resetCache()` 直接清空 `dictOptionsMap`，但已挂载的 `DictTag` 持有的是此前 `getDictOptions()` 返回的数组引用。清空 `Map` 不会更新这些旧数组，导致已打开列表继续显示旧标签。

### FB-1 影响分析

- 修改范围：`apps/lina-vben/apps/web-antd/src/utils/dict.ts`、字典数据新增/更新/删除触发路径、字典类型更新时新旧类型 key 刷新边界、相关单元测试和 E2E。
- i18n 影响：无新增运行时文案、菜单、按钮、表格标题、API 文档源文本或语言包资源；字典标签为用户可编辑业务数据。
- 缓存一致性影响：涉及前端字典进程内缓存；权威数据源为 `/dict/data/type/{dictType}`；字典数据写入成功后按 `dictType` 精确刷新并就地更新数组；刷新失败保留旧数组并允许后续重试，避免写入成功后的页面流程因缓存刷新失败中断；跨浏览器不共享前端内存缓存，刷新页面可恢复。
- 数据权限影响：无新增或修改数据操作接口；沿用现有字典接口权限和可见性边界。
- API 契约影响：无新增 HTTP API、路由、DTO、响应字段或前后端接口契约变化。
- 后端 Go 影响：无 Go 生产代码、Controller、Service、DAO、运行期依赖或 DI 变更。
- 数据库影响：无 DDL、DML、索引或 DAO 生成影响。
- 开发工具跨平台影响：无 Makefile、脚本、CI、代码生成或 `linactl` 变更。
- 插件影响：未修改 `apps/lina-plugins/<plugin-id>/` 下文件；无源码插件、动态插件或插件生命周期资源影响。
- 已读取规则：`AGENTS.md`、`openspec.md`、`documentation.md`、`architecture.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`cache-consistency.md`、`api-contract.md`、`data-permission.md`。

### FB-1 验证记录

- `pnpm -C apps/lina-vben exec vitest run apps/web-antd/src/utils/dict.test.ts --dom`：通过，覆盖刷新时保持数组引用稳定、标签和样式同步更新、并发请求去重。
- `pnpm -C apps/lina-vben --filter @lina/web-antd typecheck`：通过。
- `cd hack/tests && npx tsc --noEmit --project tsconfig.json`：通过。
- `cd hack/tests && npx playwright test e2e/settings/dict/TC010-dict-label-sync-and-tab-pagination.ts`：通过，`TC-10a` 覆盖已打开菜单列表同步显示更新后的字典标签。

### FB-2 根因

工作台 Tab 缓存只缓存 `meta.keepAlive=true` 的路由，而 `/menus/all` 的 `meta.keepAlive` 来自 `sys_menu.is_cache`。内建列表菜单 seed 当前为 `is_cache=0`，列表页切出 Tab 后重新挂载，表格分页回到默认页。

### FB-2 影响分析

- 修改范围：宿主原始菜单 Seed DML、相关 E2E。用户已明确要求不考虑兼容性，且项目顶层规范说明全新项目无历史负担，因此直接修正 `006-menu-role-management.sql` 与 `011-scheduled-job-management.sql` 的源头 seed，不保留额外补丁 SQL。
- i18n 影响：无新增运行时文案、菜单标题、API 文档源文本或语言包资源；仅修改已有菜单行的 `is_cache` 标志。
- 缓存一致性影响：涉及工作台页面实例 keep-alive，不涉及后端分布式缓存；权威数据源为 `sys_menu.is_cache`，`/menus/all` 投影 `meta.keepAlive`；前端路由缓存按该投影生效。
- 数据权限影响：无新增或修改数据读取、写入或权限判断；菜单可见性和权限仍由现有菜单/角色逻辑控制。
- API 契约影响：无新增 HTTP API、路由、DTO、响应字段或前后端接口契约变化；`/menus/all` 继续投影既有 `meta.keepAlive` 字段。
- 开发工具跨平台影响：无 Makefile、脚本、CI、代码生成或 `linactl` 变更；`make db.init` 仅作为验证入口。
- SQL 影响：修改宿主 Seed DML 中内建可分页 routed 菜单的 `is_cache` 值为 `1`；使用既有 `INSERT ... ON CONFLICT DO NOTHING` 幂等模式，不写自增 ID，不修改 DDL、索引或 DAO。
- 后端 Go 影响：无 Go 生产代码、Controller、Service、DAO 或运行期依赖变更。
- 插件影响：未修改 `apps/lina-plugins/<plugin-id>/` 下文件；插件管理宿主菜单只是宿主 seed 行，未改变插件资源或插件接口契约。
- 已读取规则：`AGENTS.md`、`openspec.md`、`documentation.md`、`architecture.md`、`frontend-ui.md`、`testing.md`、`i18n.md`、`cache-consistency.md`、`api-contract.md`、`database.md`、`backend-go.md`、`data-permission.md`、`plugin.md`、`goframe-v2`、`lina-e2e`。

### FB-2 验证记录

- `make db.init confirm=init`：通过，执行到 `012-distributed-cache-consistency.sql`，无额外 `013` 补丁 SQL。
- `cd apps/lina-core && go test ./pkg/dialect -run 'Test.*SQL|TestHostSQL|TestSeed' -count=1`：通过。
- `cd hack/tests && npx playwright test e2e/settings/dict/TC010-dict-label-sync-and-tab-pagination.ts`：通过，`TC-10b` 覆盖内建菜单 `isCache=1` 和 Tab 切回后字典数据分页仍在第 2 页。
- `git diff --check`：通过。
- `openspec validate fix-dict-sync-tab-pagination --strict`：通过。

### FB-3 根因

GitHub Actions `Go unit tests (plugin-full)` 失败于 `linactl/internal/wasmbuilder.TestPluginDemoDynamicRuntimeArtifactEmbedsReviewedAssets`。动态插件样例的 `plugin.yaml` 和测试断言要求打包 `manifest/config/config.yaml`，该文件按插件配置规范是本地或部署环境的实际运行配置，默认被插件工作区 `.gitignore` 规则 `**/manifest/config/config.yaml` 忽略，CI 干净 checkout 中不存在。正确修复是保持样例契约和测试断言继续使用 `config/config.yaml`，并由单测在构建官方动态插件样例前从 `config.example.yaml` 准备临时 `config.yaml` fixture。

### FB-3 影响分析

- 修改范围：最终代码差异集中在 `hack/tools/linactl/internal/wasmbuilder` 官方动态插件样例单测 fixture 准备逻辑；`apps/lina-plugins/linapro-demo-dynamic` 样例文件已恢复并确认继续使用 `config/config.yaml` 契约。
- i18n 影响：未产生最终插件 i18n 资源差异；已确认显式启用 `i18n` 的插件 API 文档源文本和 `zh-CN` apidoc 翻译资源继续描述 `config/config.yaml`。
- 缓存一致性影响：无运行时缓存、翻译缓存失效逻辑或分布式缓存变更；仅调整动态插件样例打包资源路径和测试期望。
- 数据权限影响：无新增或修改数据读取、写入、聚合、下载或数据可见性边界；manifest host service 仍只读插件自身已授权打包资源。
- 开发工具跨平台影响：涉及 `linactl` Go 单测 fixture 准备逻辑和断言，不新增脚本或平台专属入口；验证使用 Go 标准库文件读写、Go 工具链和根 `make test.go` 包装入口。
- 插件本地规范影响：已检查 `apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`，不存在插件本地规范入口。
- 后端 Go 影响：修改 `linactl` Go 单测文件；动态插件 Go 源文件最终无差异；无新增运行期依赖、Controller 构造函数、Service 构造函数、DAO、事务或数据库查询路径。
- API 契约影响：响应字段结构不变，`configPath` 示例值和文档描述保持插件配置规范要求的 `config/config.yaml`。
- 已读取规则：`openspec.md`、`documentation.md`、`testing.md`、`backend-go.md`、`dev-tooling.md`、`plugin.md`、`api-contract.md`、`i18n.md`、`architecture.md`、`data-permission.md`、`frontend-ui.md`、`lina-e2e`、`goframe-v2`。

### FB-3 验证记录

- `go test ./internal/wasmbuilder -run TestPluginDemoDynamicRuntimeArtifactEmbedsReviewedAssets -count=1`：通过。
- `go test -race ./internal/wasmbuilder -count=1`：通过。
- `GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins go test ./... -count=1`：通过。
- `LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' make test.go plugins=1 race=true verbose=true`：通过，覆盖 `plugin-full` Go 单测路径。
- `openspec validate fix-dict-sync-tab-pagination --strict`：通过。
- `git diff --check`、`git -C apps/lina-plugins diff --check`：通过。

### FB-4 根因

上一轮修复为了让 CI 干净 checkout 中缺失的 `manifest/config/config.yaml` 不再导致单测失败，将动态插件样例的 `manifest` host-service 授权、后端返回路径、API 文档、README 和 E2E 断言改成了 `config/config.example.yaml`。这会违背插件配置规范：`config.example.yaml` 只是可提交模板，插件运行期配置和动态 artifact 默认配置应保持 `config.yaml` 文件名。

### FB-4 影响分析

- 修改范围：确认 `apps/lina-plugins/linapro-demo-dynamic` 样例契约保持 `config/config.yaml`，并调整 `hack/tools/linactl/internal/wasmbuilder` 单测在缺少被忽略配置文件时临时准备 fixture。
- i18n 影响：插件 API 文档源文本和 `zh-CN` apidoc 翻译继续使用 `config/config.yaml` 配置路径描述，不新增语言或翻译键。
- 缓存一致性影响：无运行时缓存、翻译缓存失效逻辑或分布式缓存变更。
- 数据权限影响：无新增或修改数据读取、写入、聚合、下载或数据可见性边界；manifest host service 仍只读插件自身已授权打包资源。
- 开发工具跨平台影响：涉及 `linactl` Go 单测，使用 Go 标准库准备和清理临时文件，不新增脚本或平台专属入口。
- 插件本地规范影响：已检查 `apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`，不存在插件本地规范入口。
- 后端 Go 影响：仅修改 `linactl` Go 单测；动态插件 Go 源文件和 API DTO 文档源文本最终无差异；无新增运行期依赖、Controller 构造函数、Service 构造函数、DAO、事务或数据库查询路径。
- API 契约影响：响应字段结构不变，仅恢复 `configPath` 示例值和文档描述为 `config/config.yaml`。
- 已读取规则：`AGENTS.md`、`openspec.md`、`documentation.md`、`testing.md`、`backend-go.md`、`dev-tooling.md`、`plugin.md`、`api-contract.md`、`i18n.md`、`architecture.md`、`data-permission.md`、`cache-consistency.md`、`lina-e2e`、`goframe-v2`。

### FB-4 验证记录

- `go test ./internal/wasmbuilder -run TestPluginDemoDynamicRuntimeArtifactEmbedsReviewedAssets -count=1`：通过。
- `go test -race ./internal/wasmbuilder -count=1`：通过。
- `GOWORK=/Users/john/Workspace/github/linaproai/linapro/temp/go.work.plugins go test ./... -count=1`：通过。
- `LINA_TEST_PGSQL_LINK='pgsql:postgres:postgres@tcp(127.0.0.1:5432)/linapro?sslmode=disable' make test.go plugins=1 race=true verbose=true`：通过，覆盖 `plugin-full` Go 单测入口。
- `openspec validate fix-dict-sync-tab-pagination --strict`：通过。
- `git diff --check`、`git -C apps/lina-plugins diff --check`：通过。

## Tasks

- [x] 修复字典 Store：支持按类型刷新并保持已缓存数组引用稳定
- [x] 调整字典数据写入路径：新增、更新和删除后按当前 `dictType` 精确刷新，字典类型 key 变更时同步刷新新旧 key
- [x] 修正宿主菜单 Seed：将内建可分页 routed 菜单配置为 `is_cache=1`
- [x] 更新 E2E：覆盖字典标签同步与 Tab 分页状态保持
- [x] 修复动态插件样例 manifest host-service 资源路径并更新单测、文档和 E2E 断言
- [x] 纠正动态插件样例配置契约：保持 `config/config.yaml` 并让 CI 单测准备被忽略的本地配置 fixture
- [x] 运行验证：前端类型检查、单元测试、E2E 定向验证、SQL 初始化与静态检查、`openspec validate fix-dict-sync-tab-pagination --strict`
