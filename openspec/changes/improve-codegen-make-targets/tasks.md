## Tasks

- [x] 1. 记录反馈根因、影响范围和命中规则域。
- [x] 2. 扩展 `linactl ctrl` 和 `linactl dao`，支持宿主和插件后端目标目录。
- [x] 3. 收敛根、宿主和插件目录 `Makefile`，提供一致的 `ctrl` 与 `dao` 入口。
- [x] 4. 补充工具测试和治理验证，确认默认宿主与插件定向生成路径正确。
- [x] 5. 运行 OpenSpec 严格校验和格式检查，并完成反馈审查记录。

## Feedback

- [x] **FB-1**: `make dao`和`make ctrl`只能在宿主或插件后端目录下正确生成
- [x] **FB-2**: 插件根目录`Makefile`硬编码插件后端路径

### FB-1 处理记录

- 根因：`linactl` 的 `goframecli.Run` 和隐藏 `RunEmbedded` 固定使用仓库根目录下的 `apps/lina-core` 作为 GoFrame 项目目录，`apps/lina-core/hack/hack.mk` 也只转发到不带目标参数的 `linactl ctrl/dao`。因此生成器只能解析宿主的 `hack/config.yaml`、`api/` 和 `internal/`，无法从根目录或插件根目录直接定向到插件 `backend/`。同时 `apps/lina-core/Makefile` 仍包含历史 `cli`、`cli.install`、`up`、`image` 等本地命令，插件根目录没有与宿主一致的薄 Makefile 入口。
- 影响分析：本反馈修改开发工具、代码生成入口和插件根目录薄 Makefile；不修改 HTTP API、运行时服务、前端 UI、SQL 内容、DAO/DO/Entity 生成产物或数据库访问路径。
- `i18n` 影响：无运行时用户可见文案、菜单、路由、API 文档源文本、错误消息、插件清单、语言包或翻译缓存影响。
- 缓存一致性影响：无缓存、快照、失效、刷新、预热、集群一致性或关键运行时状态影响。
- 数据权限影响：无业务数据读写、列表、详情、导出、聚合统计、下载、授权关系或租户/组织边界影响。
- SQL 影响：不新增或修改 SQL 迁移、Seed DML、Mock 数据、插件安装/卸载 SQL、幂等逻辑或索引；仅调整 DAO 生成命令的工作目录解析。
- 开发工具跨平台影响：命中 `Makefile`、`linactl` 和代码生成入口；实现必须继续由 Go 工具承载，Makefile 仅作为薄转发层，Windows 入口继续通过 `make.cmd` 转发到 `linactl`。
- 插件规范影响：新增插件根目录 `Makefile`，仅作为开发工具 wrapper；无插件业务源码、`plugin.yaml`、生命周期资源、host service 授权或运行时桥接变更。插件根目录未发现本地 `AGENTS.md`。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/database.md`、`.agents/rules/plugin.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/i18n.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`。
- 修复：`linactl ctrl` 和 `linactl dao` 现在支持 `p=<plugin-id>`、`plugin=<plugin-id>`、`dir=<backend-dir>` 和 `target=<backend-dir>` 目标选择；不传目标时默认宿主 `apps/lina-core`。公开命令负责解析目标目录，隐藏 `__goframe` 子命令仍只接受 `gen ctrl` 或 `gen dao`，并在父进程指定的工作目录中运行内嵌 GoFrame CLI。`dao` 会提前校验目标 `hack/config.yaml`，`ctrl` 只要求目标后端目录存在，避免把 DAO 配置前置条件错误套到只生成 controller 的插件上。
- Makefile 收敛：根 `Makefile` 通过 `hack/makefiles/database.mk` 暴露统一 `ctrl` 和 `dao` 目标；`apps/lina-core/Makefile` 删除旧 `hack/hack-cli.mk`、`hack/hack.mk` 依赖和旧 `gf` 安装/镜像重复入口，只保留宿主相关薄转发；`apps/lina-plugins/Makefile` 新增需要 `p=<plugin-id>` 的聚合 `ctrl` 和 `dao`；所有官方插件根目录新增薄 `Makefile`，在插件目录内执行 `make ctrl` 或 `make dao` 时定向自身 `backend/`。
- 文档：已同步更新 `apps/lina-plugins/README.md`、`apps/lina-plugins/README.zh-CN.md`、`hack/tools/linactl/README.md` 和 `hack/tools/linactl/README.zh-CN.md`，说明插件本地 Makefile 和 `linactl` 目标参数。
- 验证：`go test ./internal/goframecli -count=1` 通过；`go test . -run 'TestRun(Ctrl|Dao)|TestEmbeddedGoFrameCtrlSmokeWithoutExternalGF|TestCommandRegistry|TestPrintHelp|TestHiddenGoFrame|TestRunEmbeddedGoFrameRejectsParameters' -count=1` 通过；`go test ./... -count=1` 在 `hack/tools/linactl` 下通过；`make -n ctrl`、`make -n ctrl p=linapro-content-notice`、`make -n dao dir=apps/lina-plugins/linapro-content-notice/backend`、`make -n -C apps/lina-core ctrl`、`make -n -C apps/lina-plugins/linapro-demo-dynamic ctrl`、`make -n -C apps/lina-plugins/linapro-content-notice dao` 输出预期转发命令；`make -n -C apps/lina-plugins ctrl` 在缺少 `p` 时按预期报错；官方插件根目录 Makefile 存在性检查通过；旧外部 `gf` 安装入口静态扫描仅命中 `main_test.go` 中验证旧命令不再注册的测试；`git diff --check` 和 `git -C apps/lina-plugins diff --check` 通过；`openspec validate improve-codegen-make-targets --strict` 通过。
- 审查记录：已按命中规则做反馈级自查，未发现阻塞问题。开发工具逻辑集中在 `hack/tools/linactl`，Makefile 仍为薄转发层；未新增 shell 脚本或平台专属默认入口；`make.cmd` 未修改且继续作为 Windows 跨平台入口转发到 `linactl`。后端生成产物未手工修改；没有 API、SQL、缓存、数据权限、运行时依赖、插件 host service 授权、前端 UI 或 i18n 资源运行时变更。文档中英文镜像已同步，OpenSpec 任务状态只在实现和验证完成后标记完成。

### FB-2 处理记录

- 根因：上一版官方插件根目录`Makefile`在`ctrl`和`dao`目标中直接写入`dir=apps/lina-plugins/<plugin-id>/backend`。这会把可复用的插件代码生成包装层耦合到具体插件 ID，新增、复制或重命名插件时容易产生路径漂移，也让同一套插件代码生成逻辑无法集中维护。
- 影响分析：本反馈仅调整插件开发工具入口和治理文档；不修改`linactl` Go 代码、HTTP API、运行时服务、前端 UI、SQL 内容、DAO/DO/Entity 生成产物或数据库访问路径。
- `i18n`影响：无运行时用户可见文案、菜单、路由、API 文档源文本、错误消息、插件清单、语言包或翻译缓存影响。
- 缓存一致性影响：无缓存、快照、失效、刷新、预热、集群一致性或关键运行时状态影响。
- 数据权限影响：无业务数据读写、列表、详情、导出、聚合统计、下载、授权关系或租户/组织边界影响。
- SQL 影响：不新增或修改 SQL 迁移、Seed DML、Mock 数据、插件安装/卸载 SQL、幂等逻辑或索引。
- 开发工具跨平台影响：命中`Makefile`和`hack/makefiles/`；实现继续由`linactl` Go 工具承载，新增`hack/makefiles/plugin.codegen.mk`仅作为 Make 薄转发层，使用`go -C "<repo>/hack/tools/linactl" run .`避免依赖 shell `cd`语义。
- 插件规范影响：官方插件根目录`Makefile`改为声明当前插件根目录并引入根目录共享`hack/makefiles/plugin.codegen.mk`；共享片段通过`$(PLUGIN_ROOT)/backend`推导目标后端目录，插件`Makefile`不再硬编码具体插件 ID 或`apps/lina-plugins/<plugin-id>/backend`。插件根目录未发现本地`AGENTS.md`。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/plugin.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/database.md`。
- 修复：新增根目录共享片段`hack/makefiles/plugin.codegen.mk`，统一维护插件`ctrl`和`dao`目标；所有官方插件根目录`Makefile`改为计算自身`PLUGIN_ROOT`和`REPO_ROOT`后 include 共享片段；`apps/lina-plugins/README.md`、`apps/lina-plugins/README.zh-CN.md`和增量规范同步记录共享片段约束。
- 验证：`make -n -C apps/lina-plugins/linapro-content-notice ctrl`、`make -n -C apps/lina-plugins/linapro-content-notice dao`、`make -n -C apps/lina-plugins/linapro-demo-dynamic ctrl`和`make -n -f apps/lina-plugins/linapro-content-notice/Makefile ctrl`均输出预期`linactl`转发命令；所有官方插件根目录`ctrl`和`dao` dry-run 循环通过；`rg -n 'apps/lina-plugins/.*/backend' apps/lina-plugins -g Makefile`无命中；`rg -n 'dir=.*linapro-|linapro-[a-z0-9-]+/backend|go run ../../../hack/tools/linactl' apps/lina-plugins -g Makefile`无命中；所有官方插件根目录`Makefile`共享 include 检查通过；`openspec validate improve-codegen-make-targets --strict`通过；`git diff --check`和`git -C apps/lina-plugins diff --check`通过。
- 审查记录：已按命中规则做反馈级自查，未发现阻塞问题。本次为治理类开发工具反馈，不改变运行时行为，因此无需新增单元测试或 E2E；使用 Make dry-run、静态扫描、OpenSpec 严格校验和格式检查作为匹配治理验证。无新增运行期依赖、DI、缓存、数据权限、API、SQL、插件 host service 授权、前端 UI 或 i18n 资源变更。
