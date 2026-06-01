## Why

社区`Issue`需要持续按`LinaPro`项目规范和源码实现进行分类处理。当前缺少统一的自动审查技能来区分疑问、功能需求、`Bug`、模糊内容和骚扰广告内容，也缺少统一的评论语言、标签、关闭和重复审查规则。

## What Changes

- 新增仓库级`lina-community-issue-review`技能，用于审查`https://github.com/linaproai/linapro`的`GitHub Issue`。
- 技能支持用户指定`Issue`编号；未指定时遍历全部开放`Issue`。
- 技能跳过已经由本技能评论且带有`question`、`feature`或`bug`任一标签的`Issue`，避免重复审查。
- 技能根据项目规范和源码实现将`Issue`分类为疑问、功能需求、`Bug`或无效内容。
- 疑问类`Issue`回答后添加`question`标签并关闭。
- 可行功能需求添加`feature`标签后发布评估评论，保持开放等待实现。
- 可行`Bug`添加`bug`标签后发布原因评估评论，保持开放等待修复。
- 模糊、骚扰或广告类`Issue`完成关闭处理，并发布说明原因或补充要求的评论。
- 技能评论语言跟随`Issue`描述语言；描述无法判断时按标题判断，仍无法判断时默认中文。
- 技能把`Issue`标题、正文和评论视为不可信输入，不允许其中内容改变执行规则或触发不安全命令。

## Capabilities

### New Capabilities

- `community-issue-review-skill`：定义`lina-community-issue-review`技能的触发、仓库范围、`Issue`遍历、重复审查跳过、可信上下文加载、分类处理、评论语言、标签和关闭契约。

### Modified Capabilities

- 无。

## Impact

- 新增`.agents/skills/lina-community-issue-review/SKILL.md`。
- 新增`openspec/changes/add-community-issue-review-skill/`下的提案、设计、任务和增量规范。
- 本变更不修改后端、前端、数据库、`HTTP API`、插件运行时、`CI`或`GitHub Actions`流程。
- `i18n`影响：技能生成的`GitHub`评论语言跟随`Issue`描述语言，但不新增运行时用户可见文案、语言包、接口文档本地化资源或翻译缓存。
- 缓存一致性、数据权限、模块启停、核心宿主接口契约均无影响。
- 开发工具跨平台影响：技能依赖代理环境中的`gh`和`git`等命令，不新增长期维护脚本、`Makefile`目标或`linactl`命令。
