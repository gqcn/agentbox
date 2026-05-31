## 1. 规则更新

- [x] 1.1 读取并确认本变更命中的 `openspec.md`、`documentation.md` 和 `testing.md` 规则。
- [x] 1.2 更新 `.agents/rules/testing.md`，新增 E2E 质量审查要求，内容只约束测试结果和审查证据，不规定具体测试实现方式。
- [x] 1.3 在任务记录中说明 `i18n`、缓存一致性、数据权限、开发工具跨平台、插件、API、后端 Go 和运行时行为无影响判断。

### 任务 1 影响分析

- `i18n`：无运行时用户可见文案、菜单、路由、按钮、表单、表格、提示信息、语言包、API 文档源文本、插件清单或翻译缓存影响。
- 缓存一致性：无缓存、快照、失效、刷新、单机或集群一致性影响。
- 数据权限：无列表、详情、导出、下载、聚合、批量信息、下拉候选、写入动作或租户/组织可见性边界影响。
- 开发工具跨平台：无 Makefile、脚本、CI、代码生成、服务启停、`linactl` 或跨平台入口变更。
- 插件：未修改 `apps/lina-plugins/<plugin-id>/` 文件、插件资源、插件生命周期、插件宿主服务或动态插件运行时边界。
- API：无 HTTP API、路由、DTO、OpenAPI 元数据、权限标签或响应结构影响。
- 后端 Go：无 Go 生产代码、测试代码、Controller、Middleware、Service、依赖注入或数据库访问路径影响。
- 运行时行为：本变更仅修改项目测试治理规则，不改变应用运行期行为。
- 测试策略：属于治理文档变更，使用 OpenSpec 严格校验、静态检索、`git diff --check` 和审查结论验证。

## 2. 规范验证

- [x] 2.1 运行 `openspec validate enhance-e2e-quality-review-rules --strict`。
- [x] 2.2 运行 Markdown/静态检索，确认新增规则未引入具体定位器、封装模式或实现方式硬约束。
- [x] 2.3 运行 `git diff --check`。

### 任务 2 验证记录

- `openspec validate enhance-e2e-quality-review-rules --strict`：通过。
- `rg -n "优先使用|必须使用.*getBy|必须使用.*data-testid|必须使用.*Page Object|必须.*封装选择器|不得.*硬编码文本|nth-child.*判定|定位器类型.*必须|getByTestId|getByRole|data-testid|nth-child|nth-of-type|Page Object" .agents/rules/testing.md openspec/changes/enhance-e2e-quality-review-rules/proposal.md openspec/changes/enhance-e2e-quality-review-rules/design.md openspec/changes/enhance-e2e-quality-review-rules/specs`：仅命中 `.agents/rules/testing.md` 中既有的“后端纯逻辑优先使用单元测试”，未命中新增 E2E 质量规则或 OpenSpec 产物中的具体定位器、封装模式或实现方式硬约束。
- `git diff --check`：通过。

## 3. 审查

- [x] 3.1 使用 `lina-review` 对本变更执行规则、文档和 OpenSpec 合规审查。
- [x] 3.2 根据审查结论修复阻塞问题，并记录剩余警告或无影响判断。

### Lina 审查记录

**变更：** `enhance-e2e-quality-review-rules`
**范围：** 全部变更
**审查文件数：** 5
**范围来源：** `git status --short`、`git ls-files --others --exclude-standard`、OpenSpec 当前变更上下文
**已读取规则文件：** `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/testing.md`

#### 发现的问题

- 未发现阻塞问题。

#### 规则域结论

- OpenSpec：通过。活跃变更 `enhance-e2e-quality-review-rules` 未归档，`proposal.md`、`design.md`、`specs/e2e-suite-organization/spec.md` 和 `tasks.md` 均存在，`openspec validate enhance-e2e-quality-review-rules --strict` 通过。
- 文档治理：通过。未新增目录级 `README`，无需中英文镜像；Markdown 使用中文上下文，未使用普通段落分隔线。
- 测试治理：通过。新增规则为 E2E 质量审查的结果级约束，未引入具体定位器、封装模式或实现方式硬约束；本变更为治理文档变更，不涉及运行时单元测试或 E2E 用例新增。
- `i18n`：无影响。未修改运行时用户可见文案、语言包、API 文档源文本、插件清单或翻译缓存。
- 缓存一致性：无影响。未修改缓存、快照、失效、刷新或集群一致性逻辑。
- 数据权限：无影响。未修改任何数据读取、写入、导出、聚合、下拉候选或可见性边界。
- 开发工具跨平台：无影响。未修改脚本、CI、Makefile、代码生成或 `linactl`。
- 插件：无影响。未修改插件目录、插件资源、插件生命周期或宿主插件契约。
- API、后端 Go、数据库、前端 UI：无影响。未修改运行时代码、HTTP API、数据库、前端页面或用户交互。

#### 验证证据

- `openspec validate enhance-e2e-quality-review-rules --strict`：通过。
- 静态检索具体实现硬约束：通过；仅命中既有“后端纯逻辑优先使用单元测试”，未命中新增 E2E 质量规则或 OpenSpec 产物中的具体定位器、封装模式或实现方式硬约束。
- `git diff --check`：通过。

#### 摘要

- 严重：0
- 警告：0
