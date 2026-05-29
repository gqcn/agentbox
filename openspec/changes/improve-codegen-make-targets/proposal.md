## Why

当前 `make dao` 和 `make ctrl` 的实际生成能力已经由 `linactl` 内嵌 `GoFrame CLI` 承载，但默认目标目录固定为 `apps/lina-core`。这使根目录和聚合入口无法直接对插件后端执行同一套生成流程，开发者仍需要进入宿主或插件后端目录才能得到正确的 `hack/config.yaml`、`api/` 和 `internal/` 路径解析结果。

同时，`apps/lina-core/Makefile` 仍包含部分历史本地命令和旧 `gf` 安装心智，插件根目录也缺少与宿主一致的薄 `Makefile` 入口。项目是全新项目，不需要保留这些历史入口，应将代码生成实现集中维护在 `linactl` 和根 `Makefile`，宿主与插件目录只保留一致的转发层。

## What Changes

- 扩展 `linactl ctrl` 和 `linactl dao`，支持显式指定生成目标目录，并默认继续指向 `apps/lina-core`。
- 根 `Makefile` 新增或统一 `ctrl`、`dao` 目标，支持默认宿主生成，也支持通过插件 ID 定向插件后端生成。
- 清理 `apps/lina-core/Makefile` 中与代码生成无关或旧 `gf` 安装相关的本地命令，只保留转发到根工具链的薄入口。
- 为官方插件根目录增加一致的 `Makefile`，通过根目录共享的`hack/makefiles/plugin.codegen.mk`提供`ctrl`和`dao`入口，并转发到仓库统一`linactl`。
- 保持 `linactl` 内嵌 `GoFrame CLI`、隐藏子进程隔离和白名单生成命令边界，不重新依赖外部 `gf` 可执行文件。

## Impact

- 影响仓库开发工具入口：根 `Makefile`、`apps/lina-core/Makefile`、插件根 `Makefile`、`hack/tools/linactl`。
- 影响 GoFrame 生成命令的目标目录解析，不修改生成产物、SQL 迁移、HTTP API、运行时服务或前端 UI。
- 插件目录影响仅为新增根目录薄`Makefile`并引入共享`hack/makefiles/plugin.codegen.mk`，不修改插件业务源码、插件清单、生命周期资源或运行时授权边界。
- `i18n` 影响：无运行时用户可见文案、菜单、API 文档源文本、语言包或翻译缓存影响。
- 缓存一致性影响：无缓存、快照、失效、刷新或集群一致性影响。
- 数据权限影响：无业务数据读取、写入、列表、详情、导出、聚合或租户/组织可见性影响。
- SQL 影响：不新增或修改 SQL 文件、迁移账本、DAO 生成输入内容；仅调整 DAO 生成入口的工作目录解析。
