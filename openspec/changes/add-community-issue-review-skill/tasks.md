## 1. 技能实现

- [x] 1.1 新建`.agents/skills/lina-community-issue-review/`目录和`SKILL.md`，声明触发场景、默认仓库、依赖工具和不可信输入边界。
- [x] 1.2 在`SKILL.md`中定义`Issue`范围解析、开放`Issue`遍历、指定`Issue`审查和重复审查跳过规则。
- [x] 1.3 在`SKILL.md`中定义可信项目规范与源码加载流程，以及疑问、功能需求、`Bug`、模糊骚扰广告内容的分类规则。
- [x] 1.4 在`SKILL.md`中定义评论语言、评论模板、`question`、`feature`、`bug`标签和关闭`Issue`动作。

## 2. 质量验证

- [x] 2.1 运行`openspec validate add-community-issue-review-skill --strict`并记录结果。
- [x] 2.2 执行技能结构和`frontmatter`静态检查，确认新技能可被发现。
- [x] 2.3 执行静态检索检查，确认技能覆盖默认仓库、指定`Issue`、全部开放`Issue`、重复审查跳过、语言跟随、三类标签、无效关闭和不可信输入边界。
- [x] 2.4 检查本次变更未修改`.github/workflows/`、后端、前端、数据库或插件运行时代码。

## 3. 治理与门禁

- [x] 3.1 记录影响分析：`i18n`仅影响技能生成评论语言，缓存一致性无影响，数据权限无影响，模块启停无影响，核心宿主接口契约无影响，开发工具跨平台无新增长期脚本。
- [x] 3.2 完成实现后调用`lina-review`进行规范和实现审查。

## 4. 执行记录

- 技能实现：已新增`.agents/skills/lina-community-issue-review/SKILL.md`。技能覆盖默认仓库`linaproai/linapro`、指定`Issue`审查、全部开放`Issue`遍历、评论和标签双条件跳过、可信项目规范与源码加载、不可信`Issue`输入边界、评论语言跟随`Issue`描述、疑问类回答并关闭、可行功能需求打`feature`标签、可行`Bug`打`bug`标签、模糊骚扰广告类关闭和最终报告要求。
- 验证命令：`openspec validate add-community-issue-review-skill --strict`通过；`git diff --check`通过；`quick_validate.py`因本地`Python`环境缺少`PyYAML`无法运行，已用`Ruby YAML`等价检查`SKILL.md`的`frontmatter`，确认`name`为`lina-community-issue-review`、`description`非空、无尖括号且长度为`375`；静态检索确认技能覆盖默认仓库、指定`Issue`、全部开放`Issue`、重复审查跳过、语言跟随、三类标签、无效关闭和不可信输入边界。
- 变更范围检查：`git diff --name-only -- .github apps manifest hack Makefile make.cmd`无输出，确认未修改`.github/workflows/`、后端、前端、数据库、插件运行时代码、构建入口或长期脚本。
- 影响分析：`i18n`仅影响技能生成`GitHub`评论时的语言选择，不修改运行时语言包、接口文档翻译或翻译缓存；缓存一致性无影响；数据权限无影响；模块启停无影响；核心宿主接口契约无影响；开发工具跨平台无新增长期脚本、`Makefile`目标或`linactl`命令；未新增运行期依赖，`DI`来源无影响。
- `lina-review`结果：审查范围为`.agents/skills/lina-community-issue-review/SKILL.md`和`openspec/changes/add-community-issue-review-skill/`下的`OpenSpec`文件。已按`git status --short --untracked-files=all`展开未跟踪目录，并读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`和`.agents/instructions/markdown-format.instructions.md`。审查中发现成功评论模板会声明标签或关闭状态，若先发布评论再执行`GitHub`状态变更，权限失败时可能留下不准确评论；已修正技能和增量规范，要求先完成标签、保持开放或关闭等状态变更，再发布最终成功评论。修正后重新运行`openspec validate add-community-issue-review-skill --strict`、`git diff --check`和`Ruby YAML`检查均通过。未发现阻塞问题；剩余风险：本次未对线上`Issue`执行真实评论、标签或关闭操作，验证以静态治理检查为主。
