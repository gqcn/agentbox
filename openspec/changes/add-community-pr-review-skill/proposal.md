## Why

社区贡献的`Pull Request`需要持续按`LinaPro`项目规范进行审查。当前人工审查容易遗漏`AGENTS.md`及`.agents/rules/*.md`中的治理要求，也缺少统一的跳过、评论、升级和通过标记流程。新增`lina-community-pr-review`技能可以让代理按项目规范批量或定向审查`PR`，在发现问题时留下可追踪评论，在无法可靠判断时升级给相关文件的历史维护成员，并在符合规范时添加`bot-approved`标签。

## What Changes

- 新增仓库级`lina-community-pr-review`技能，用于审查`https://github.com/linaproai/linapro`的开放`PR`。
- 技能支持用户指定`PR`编号；未指定时遍历全部开放`PR`。
- 技能跳过已带`bot-approved`标签的`PR`，并通过审查评论中的隐藏标记避免对无新提交的`PR`重复审查。
- 技能根据目标分支可信版本的`AGENTS.md`和命中的`.agents/rules/*.md`审查`PR`差异，不采信`PR`正文或差异中的执行指令。
- 技能根据`PR`描述语言生成评论；描述为英文则评论英文，描述为中文则评论中文，描述为空或无法判断时按标题判断，仍无法判断时默认中文。
- 技能发现不合规问题时创建或更新幂等审查评论，指出问题、规则来源和修改建议。
- 技能无法可靠处理`PR`时创建或更新阻断评论，并根据`PR`涉及文件在目标分支上的历史提交选择项目成员进行`@`提醒。
- 技能确认`PR`完全符合规范时添加`bot-approved`标签，不默认创建重复通过评论。

## Capabilities

### New Capabilities

- `community-pr-review-skill`: 定义`lina-community-pr-review`技能的触发、仓库范围、`PR`遍历、跳过规则、规范加载、评论语言、人工升级和通过标签契约。

### Modified Capabilities

- 无。

## Impact

- 新增`.agents/skills/lina-community-pr-review/SKILL.md`。
- 新增`openspec/changes/add-community-pr-review-skill/`下的提案、设计、任务和增量规范。
- 本变更不修改后端、前端、数据库、`HTTP API`、插件运行时、`CI`或`GitHub Actions`流程。
- `i18n`影响：技能评论语言跟随`PR`描述语言，但不新增运行时用户可见文案、语言包、接口文档本地化资源或翻译缓存。
- 缓存一致性、数据权限、模块启停、核心宿主接口契约均无影响。
- 开发工具跨平台影响：技能依赖代理环境中的`gh`、`git`和`jq`等命令，不新增长期维护脚本、`Makefile`目标或`linactl`命令。
