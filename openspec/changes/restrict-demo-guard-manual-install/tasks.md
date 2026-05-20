## Feedback

- [x] **FB-1**: 演示控制源码插件允许通过插件管理页面安装，容易误启用全局只读保护
- [x] **FB-2**: 演示控制插件页面安装被拒绝时错误提示夹带生命周期内部信息和英文 fallback，不够友好

## Implementation

- [x] 1. 为源码插件 `BeforeInstall` 生命周期传递启动自动启用安装上下文。
- [x] 2. 为 `linapro-ops-demo-guard` 注册 `BeforeInstall` 回调并拒绝非 `plugin.autoEnable` 安装。
- [x] 3. 补充 Go 单元测试覆盖普通安装拒绝、启动自动启用允许和插件回调判断。
- [x] 4. 运行 OpenSpec、Go 编译门禁和审查。
- [x] 5. 本地化生命周期 veto 原因并优化演示控制插件页面安装拒绝文案。

## Verification

- [x] `openspec validate restrict-demo-guard-manual-install --strict`
- [x] `cd apps/lina-core && go test ./pkg/pluginhost -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/plugin -run 'TestSourceLifecycleBeforeInstall(RejectsManualWhenStartupAutoEnableRequired|ReceivesStartupAutoEnableFlag)' -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/plugin -run 'TestSourceLifecyclePreconditionLocalizesReasonParams|TestSourceLifecycleBeforeInstallRejectsManualWhenStartupAutoEnableRequired|TestSourceLifecycleBeforeInstallReceivesStartupAutoEnableFlag' -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/plugin -count=1`
- [x] `cd apps/lina-core && go test ./internal/service/i18n -count=1`
- [x] `cd apps/lina-core && go test ./internal/cmd -count=1`
- [x] `tmpdir=$(mktemp -d) && (cd "$tmpdir" && go work init /Users/john/Workspace/github/linaproai/linapro/apps/lina-core /Users/john/Workspace/github/linaproai/linapro/apps/lina-plugins/linapro-ops-demo-guard) && GOWORK="$tmpdir/go.work" go test ./backend -count=1` from `apps/lina-plugins/linapro-ops-demo-guard`
- [x] `go run ./hack/tools/linactl i18n.check`
- [x] `go run ./hack/tools/linactl test.scripts`

## Review Notes

- [x] i18n 影响：新增并优化 `linapro-ops-demo-guard` 插件运行时错误文案，同步宿主生命周期错误模板、插件 `manifest/i18n/en-US/error.json` 与 `manifest/i18n/zh-CN/error.json`，并通过 `linactl i18n.check`。
- [x] 缓存一致性：本次不新增缓存；启动自动启用仍复用既有 registry 收敛、启用快照刷新、runtime cache 修订通知和集群主节点共享副作用控制。
- [x] 数据权限影响：本次不新增或扩大数据操作接口；插件安装仍为平台插件治理写操作，由既有平台治理权限和生命周期前置条件控制。
- [x] 开发工具脚本影响：未新增或修改开发脚本；`linactl test.scripts` 已通过。
- [x] `/lina-review`：审查范围包含源码插件生命周期输入、启动自动启用安装路径、生命周期 veto 原因本地化、演示控制插件 `BeforeInstall`、panic allowlist、插件 i18n JSON、单元测试和 OpenSpec 记录；未发现阻塞问题。
