## 1. 规则与影响确认

- [x] 1.1 实施前重新读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`，确认本变更仅影响 OpenSpec 文档、技能治理说明和月度归档提示词。
- [x] 1.2 在任务记录中明确无运行时代码、HTTP API、数据库、缓存、数据权限、前端 UI、用户可见运行时文案和插件目录变更；`i18n`影响限于中文治理提示词和中文 OpenSpec 文档。
- [x] 1.3 记录开发工具跨平台影响：本变更不新增脚本、命令或平台专属执行入口；月度 CI 仍通过既有工具运行时调用共享提示词。

实施记录：

- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`。
- 运行时影响：无 Go、HTTP API、数据库、缓存、数据权限、前端 UI、用户可见运行时文案、插件目录或运行期依赖变更。
- `i18n`影响：不修改运行时语言包、`manifest/i18n`、`apidoc i18n JSON`或用户界面文案；仅修改中文 OpenSpec 文档、中文技能说明和中文月度归档提示词。
- 开发工具跨平台影响：不新增脚本、命令、`linactl`入口或平台专属执行依赖；月度 CI 仍通过既有工具运行时读取共享提示词。
- 测试策略：本变更为治理文档和技能说明变更，不涉及可执行业务行为；验证采用 OpenSpec 严格校验、静态检索、空白格式检查和审查结论。

## 2. 技能说明改造

- [x] 2.1 更新`.agents/skills/lina-archive-consolidate/SKILL.md`的技能描述，明确该技能负责归档聚合和高价值摘要压缩。
- [x] 2.2 在输入模式中增加显式压缩既有非日期前缀聚合目录的规则，并保持默认只处理日期前缀原始归档目录。
- [x] 2.3 增加信息分层规则：`proposal.md`承载背景和影响，`design.md`承载设计决策和演进，`specs/`承载最终能力契约，`tasks.md`承载反馈、根因、验证、审查和治理影响摘要。
- [x] 2.4 增加`tasks.md`高价值抽取规则，覆盖`FB-*`、根因、修复说明、验证证据、审查结论、`i18n`、缓存一致性、数据权限、DI、跨平台和测试策略。
- [x] 2.5 增加低价值流水裁剪规则，允许合并或删除普通 checklist、重复命令、逐文件搬迁清单和已被设计或规范覆盖的执行流水。
- [x] 2.6 增加语义覆盖门禁和失败优先规则：无法确认高价值信息已迁移时，不得清理原始归档目录。
- [x] 2.7 更新最终报告模板，包含保留信息类别、裁剪信息类别、未压缩或未清理原因和验证结果。

## 3. 月度归档提示词对齐

- [x] 3.1 更新`.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md`，要求无人值守流程使用增强后的`lina-archive-consolidate`摘要压缩和语义覆盖门禁。
- [x] 3.2 明确月度 CI 中无法确认语义覆盖时必须失败，不得静默清理原始归档或继续创建归档 PR。
- [x] 3.3 确认提示词变更不引入平台专属命令或新的执行依赖。

## 4. 验证与审查

- [x] 4.1 运行`openspec validate improve-archive-consolidation-compaction --strict`并记录结果。
- [x] 4.2 静态检索确认`lina-archive-consolidate`技能说明包含关键规则：`FB-`、根因、验证、审查、语义覆盖、压缩报告和非日期目录保护。
- [x] 4.3 静态检索确认月度归档聚合提示词引用摘要压缩、语义覆盖和失败优先规则。
- [x] 4.4 运行`git diff --check`，确认 Markdown 和提示词变更无空白格式问题。
- [x] 4.5 执行`lina-review`审查本变更，重点检查 OpenSpec 规范、技能治理、文档规则、测试策略、`i18n`影响判断和开发工具跨平台影响记录。

验证记录：

- `openspec validate improve-archive-consolidation-compaction --strict`：通过，输出`Change 'improve-archive-consolidation-compaction' is valid`。
- `rg -n "FB-|根因|验证|审查|语义覆盖|压缩报告|非日期|摘要压缩|数据权限|DI|跨平台|测试策略" .agents/skills/lina-archive-consolidate/SKILL.md`：通过，命中技能描述、摘要压缩边界、显式非日期目录规则、`tasks.md`维护摘要写法、清理门禁、报告模板和硬性规则。
- `rg -n "摘要压缩|语义覆盖|失败|不得静默清理|OpenSpec 校验|保留的高价值|裁剪的低价值" .github/prompts/monthly-openspec-archive-consolidate.zh-CN.md`：通过，命中无人值守摘要压缩、语义覆盖、失败优先和报告要求。
- `git diff --check -- .agents/skills/lina-archive-consolidate/SKILL.md .github/prompts/monthly-openspec-archive-consolidate.zh-CN.md openspec/changes/improve-archive-consolidation-compaction`：通过，无输出。
- `lina-review`审查结论：未发现阻塞问题。审查范围包含`.agents/skills/lina-archive-consolidate/SKILL.md`、`.github/prompts/monthly-openspec-archive-consolidate.zh-CN.md`和`openspec/changes/improve-archive-consolidation-compaction/`下全部新建文件。已按`AGENTS.md`读取并检查`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/testing.md`和`.agents/rules/i18n.md`。审查中发现并修复了技能正文使用`---`作为普通段落分隔线的问题，复查确认正文不再存在普通分隔线，仅保留 YAML frontmatter。无运行时代码、API、数据库、缓存、数据权限、前端 UI、插件目录、运行时文案或新平台命令影响；测试策略采用治理验证，剩余风险为语义摘要质量依赖后续真实归档执行时的覆盖门禁和报告审查。
