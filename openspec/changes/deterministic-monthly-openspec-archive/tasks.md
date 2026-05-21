## 1. Deterministic Archive Workflow

- [x] 1.1 Add a shared monthly OpenSpec deterministic archive composite action.
- [x] 1.2 Wire Codex, Claude Code, and GitHub Copilot reusable workflows to use deterministic archive before AI consolidation.
- [x] 1.3 Keep final failure signaling for completed active changes that remain after partial archive success.
- [x] 1.4 Upgrade artifact upload workflow actions away from the Node 20 runtime generation.

## 2. OpenSpec Archive Blocker Fix

- [x] 2.1 Fix `remove-sqlite-support` so `openspec archive -y remove-sqlite-support` can apply against the current baseline.

## 3. Verification

- [x] 3.1 Run OpenSpec validation for this change and `remove-sqlite-support`.
- [x] 3.2 Run workflow YAML/action validation and shell syntax checks for modified CI files.
- [x] 3.3 Run a temporary-copy deterministic archive smoke that covers partial success and the fixed `remove-sqlite-support` blocker.
- [x] 3.4 Record i18n, cache, data permission, REST API, E2E, and Go production code impact.
- [x] 3.5 Run `lina-review` for the CI/OpenSpec governance change.

## Feedback

- [x] **FB-1**: Monthly OpenSpec archive failed because AI auto archive returned success while completed active changes remained unarchived.
- [x] **FB-2**: `remove-sqlite-support` cannot be archived because its OpenSpec delta removes a requirement header that no longer exists in the baseline spec.

## Verification Notes

- FB-1 修复：新增 `.github/actions/monthly-openspec-auto-archive`，使用固定 OpenSpec CLI 版本执行 `openspec list --json`，按名称稳定排序 completed active changes，逐个运行 `openspec archive -y <change>`。当 OpenSpec 提供任务计数且 `completedTasks != totalTasks` 时直接记录失败，不执行归档。每个候选归档后重新执行 `openspec list --json`，确认该 change 已离开活跃列表；即使 OpenSpec CLI 在 `Aborted` 场景下返回 0，也会被识别为失败。
- FB-1 workflow 接入：`.github/workflows/monthly-openspec-archive-{codex,cc,copilot}.yml` 先执行确定性归档，再检测 OpenSpec diff；只有存在 diff 时才准备对应 AI runtime 并执行 archive consolidation。`monthly-openspec-assert-archive-complete` 移到 PR finalization 之后，并仅在 deterministic archive 报告失败时运行，从而允许成功归档部分先写入归档 PR，同时保留失败 job 信号。
- FB-1 prompt 清理：基础自动归档不再通过 AI prompt 执行，已删除废弃的 `.github/prompts/monthly-openspec-auto-archive.zh-CN.md`；保留 archive consolidation prompt 供 AI 聚合阶段使用。
- FB-1 artifact 升级：所有 `actions/upload-artifact@v4` 已升级为 `actions/upload-artifact@v7`，静态扫描确认 `.github` 中不再存在 `upload-artifact@v4`。
- FB-2 修复：将 `remove-sqlite-support` 中与当前主规范不匹配的 REMOVED/MODIFIED header 调整为现有 baseline requirement，删除已经被当前 baseline 吸收且不存在的 SQLite 专属 REMOVED delta；新增 header mismatch 检查确认该变更所有 MODIFIED/REMOVED requirement 标题均存在于当前主规范。
- 验证通过：`openspec validate remove-sqlite-support --strict`。
- 验证通过：`openspec validate deterministic-monthly-openspec-archive --strict`。
- 验证通过：`ruby -e 'require "yaml"; ARGV.each { |f| YAML.load_file(f); puts "ok #{f}" }' .github/workflows/*.yml .github/actions/*/action.yml`。
- 验证通过：`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/monthly-openspec-archive.yml .github/workflows/monthly-openspec-archive-codex.yml .github/workflows/monthly-openspec-archive-cc.yml .github/workflows/monthly-openspec-archive-copilot.yml .github/workflows/reusable-e2e-tests.yml .github/workflows/reusable-host-only-build-smoke.yml .github/workflows/reusable-redis-cluster-smoke.yml`。
- 验证通过：对 `remove-sqlite-support` 执行 Node header mismatch 扫描，所有 MODIFIED/REMOVED requirement 标题均匹配当前 `openspec/specs/<capability>/spec.md`。
- 验证通过：在 `/tmp/linapro-action-smoke` 临时副本中抽取 `.github/actions/monthly-openspec-auto-archive/action.yml` 的实际 `run` 脚本执行。结果：`archived-count=8`、`failed-count=0`、`had-failures=false`、`failed-json=[]`、`openspec list --json` 剩余 `changes=[]`、`openspec validate --all` 为 `92 passed, 0 failed`。
- 验证通过：`git diff --check -- .github openspec/changes/deterministic-monthly-openspec-archive openspec/changes/remove-sqlite-support/specs`。
- i18n 影响：本次仅修改 GitHub Actions workflow/composite action、OpenSpec 变更文档和 OpenSpec delta，不新增、修改或删除用户运行时可见文案、前端语言包、宿主/插件 `manifest/i18n` 或 apidoc i18n JSON。
- 缓存一致性影响：本次不修改运行时业务缓存、缓存键、失效触发、分布式协调或跨实例一致性逻辑；新增的 CI action 只在 GitHub runner 工作区内执行 OpenSpec 归档，不涉及生产缓存。
- 数据权限影响：本次不新增或修改 HTTP/API 数据操作接口、服务数据访问路径、插件宿主服务适配器或聚合统计，不影响角色数据权限边界。
- REST API 影响：本次不新增或修改 REST API。
- E2E 影响：本次为 CI/OpenSpec 治理修复，不涉及用户可观察页面、路由、表单、表格或端到端业务流程；使用 OpenSpec、workflow/actionlint、header mismatch 扫描和临时归档 smoke 作为治理验证。
- Go 生产代码影响：本次不新增或修改 Go 生产代码，不触发后端 Go 编译门禁。`actionlint` 通过 `go run` 执行属于外部验证工具，不改变仓库 Go 源码。
- Review：已按 `lina-review` 口径完成审查。审查范围来源包括 `git status --short`、`git ls-files --others --exclude-standard`、`openspec status --change deterministic-monthly-openspec-archive --json`、`openspec status --change remove-sqlite-support --json`、`.github` 与目标 OpenSpec 文件 diff、OpenSpec strict 校验、workflow/action YAML 解析、actionlint、`remove-sqlite-support` header mismatch 扫描和临时 action smoke。确认 `.github/actions/monthly-openspec-auto-archive` 会在归档后复查 active list，能识别 OpenSpec CLI `Aborted` 但退出码为 0 的场景；三条 monthly reusable workflow 均先执行确定性归档，再按需准备 AI runtime 做聚合，且 finalization 后保留剩余 completed active change 的失败信号；`upload-artifact@v4` 已清理。严重问题 0；警告 0。当前工作区仍存在与本次无关的 Go、前端、测试与其他 OpenSpec 改动，本次未修改或回退。
