## 1. 技能实现

- [x] 1.1 新建`.agents/skills/lina-community-pr-review/`目录和`SKILL.md`，声明触发场景、默认仓库和依赖工具。
- [x] 1.2 在`SKILL.md`中定义`PR`范围解析、开放`PR`遍历、指定`PR`审查和`bot-approved`跳过规则。
- [x] 1.3 在`SKILL.md`中定义基于隐藏评论标记的幂等策略，确保无新提交的`PR`不会重复审查。
- [x] 1.4 在`SKILL.md`中定义可信规范读取、规则域判定和不可信`PR`输入边界。
- [x] 1.5 在`SKILL.md`中定义问题评论、阻断评论、通过标签和评论语言跟随`PR`描述的规则。
- [x] 1.6 在`SKILL.md`中定义无法处理`PR`时根据文件历史选择项目成员并`@`提醒的升级流程。

## 2. 质量验证

- [x] 2.1 运行`openspec validate add-community-pr-review-skill --strict`并记录结果。
- [x] 2.2 执行文件存在性和`frontmatter`静态检查，确认新技能可被发现。
- [x] 2.3 执行静态检索检查，确认技能覆盖默认仓库、指定`PR`、全部开放`PR`、`bot-approved`跳过、隐藏标记、语言跟随、人工升级和历史维护成员选择。
- [x] 2.4 检查本次变更未修改`.github/workflows/`、后端、前端、数据库或插件运行时代码。

## 3. 治理与门禁

- [x] 3.1 记录影响分析：`i18n`仅影响技能生成评论语言，缓存一致性无影响，数据权限无影响，模块启停无影响，核心宿主接口契约无影响，开发工具跨平台无新增长期脚本。
- [x] 3.2 完成实现后调用`lina-review`进行规范和实现审查。

## 4. 执行记录

- 技能实现：已新增`.agents/skills/lina-community-pr-review/SKILL.md`，覆盖默认仓库`linaproai/linapro`、指定`PR`审查、全部开放`PR`遍历、`bot-approved`跳过、隐藏评论标记、可信规范读取、不可信`PR`输入边界、评论语言跟随`PR`描述、问题评论、阻断评论、通过标签和历史维护成员`@`升级流程。
- 验证命令：`openspec validate add-community-pr-review-skill --strict`通过；`git diff --check`通过；`Ruby YAML`检查确认`SKILL.md`的`frontmatter`包含有效`name`和`description`；静态检索确认技能覆盖默认仓库、`bot-approved`、`headRefOid`、`Comment Language`、`Reviewer Escalation`、`collaborators`、`AGENTS.md`、`.agents/rules`和隐藏评论标记。
- 变更范围检查：`git diff --name-only -- .github apps manifest hack Makefile make.cmd`无输出，确认未修改`.github/workflows/`、后端、前端、数据库、插件运行时代码、构建入口或长期脚本。
- 影响分析：`i18n`仅影响技能生成`GitHub`评论时的语言选择，不修改运行时语言包、接口文档翻译或翻译缓存；缓存一致性无影响；数据权限无影响；模块启停无影响；核心宿主接口契约无影响；开发工具跨平台无新增长期脚本、`Makefile`目标或`linactl`命令；未新增运行期依赖，`DI`来源无影响。
- `lina-review`结果：审查范围为`.agents/skills/lina-community-pr-review/SKILL.md`和`openspec/changes/add-community-pr-review-skill/`下的`OpenSpec`文件；已按`git status --short`和`git ls-files --others --exclude-standard`展开未跟踪目录。审查已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`和`.agents/instructions/markdown-format.instructions.md`。审查中发现成员升级流程只查询近期提交可能遗漏较早维护者，已修正为分页查询文件提交历史并重新验证；修正后未发现阻塞问题。剩余风险：该技能为`GitHub`侧效应工作流，本次未实际对线上`PR`发表评论或添加标签，验证以静态治理检查为主。

## Feedback

- [x] **FB-1**: 技能说明需要改为中文描述

### FB-1 执行记录

- 根因：首版`.agents/skills/lina-community-pr-review/SKILL.md`沿用了通用技能写法，`frontmatter.description`、`compatibility`、章节标题和执行说明主要使用英文；这与当前仓库`lina-*`技能中文描述习惯以及用户明确要求“技能使用中文描述”不一致。
- 修复：已将`SKILL.md`的`frontmatter.description`、`compatibility`、章节标题和执行说明整体改为中文；英文`PR`问题评论模板和阻断评论模板保留英文内容，因为技能仍需在`PR`描述为英文时发布英文评论。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`；同时按`skill-creator`读取技能创建规范。
- 验证命令：`openspec validate add-community-pr-review-skill --strict`通过；`git diff --check`通过；`Ruby YAML`检查确认`SKILL.md`的`frontmatter`包含中文`description`和中文`compatibility`；静态检索确认旧英文章节标题与英文元描述已清理。
- 影响分析：`i18n`仅涉及技能文档语言调整，不修改运行时用户可见文案、语言包、接口文档源文本或翻译缓存；缓存一致性无影响；数据权限无影响；模块启停无影响；核心宿主接口契约无影响；开发工具跨平台无新增长期脚本、`Makefile`目标或`linactl`命令；未新增运行期依赖，`DI`来源无影响。
- `lina-review`结果：审查范围为`.agents/skills/lina-community-pr-review/SKILL.md`和`openspec/changes/add-community-pr-review-skill/tasks.md`。已按`git status --short`和`git ls-files --others --exclude-standard`展开未跟踪目录，并读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`和`.agents/instructions/markdown-format.instructions.md`。验证证据覆盖当前工作区：`openspec validate add-community-pr-review-skill --strict`通过，尾随空白静态检查通过，`Ruby YAML`中文`frontmatter`检查通过，`tasks.md`无未完成任务。未发现阻塞问题；剩余风险仍为未对线上`PR`执行真实评论或标签操作。
