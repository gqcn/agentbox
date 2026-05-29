## ADDED Requirements

### Requirement: 归档聚合必须执行高价值摘要压缩

`lina-archive-consolidate` SHALL 在生成或更新聚合归档目录时，对输入归档内容执行高价值摘要压缩，而不是把原始`proposal.md`、`design.md`、`tasks.md`和`specs/`内容机械拼接到输出文档中。

#### Scenario: 聚合日期前缀原始归档

- **WHEN** 用户调用`lina-archive-consolidate`且未指定变更列表
- **THEN** 技能仅选择`openspec/changes/archive/`下目录名匹配`YYYY-MM-DD-`前缀的原始归档目录作为默认输入
- **AND** 技能按功能领域生成或更新非日期前缀聚合归档目录
- **AND** 技能对输入归档内容执行摘要压缩，保留背景、设计、规范、反馈、验证和审查的高价值语义

#### Scenario: 明确压缩既有聚合目录

- **WHEN** 用户明确指定非日期前缀归档目录并要求执行摘要压缩
- **THEN** 技能可以读取指定的非日期前缀归档目录
- **AND** 技能不得默认删除这些非日期前缀目录
- **AND** 技能必须在最终报告中说明这些目录是显式压缩输入而非默认日期归档输入

### Requirement: 聚合文档必须按信息类型分层承载历史语义

`lina-archive-consolidate` SHALL 将归档历史语义按维护目标分配到聚合归档的`proposal.md`、`design.md`、`tasks.md`和`specs/`，避免任一文件承担所有历史过程。

#### Scenario: 背景和影响进入 proposal

- **WHEN** 输入归档的`proposal.md`包含`Why`、背景、目标、影响或范围说明
- **THEN** 聚合后的`proposal.md`包含对应语义摘要
- **AND** 聚合后的`proposal.md`不得保留低价值的迭代来源元信息

#### Scenario: 设计决策进入 design

- **WHEN** 输入归档的`design.md`或`tasks.md`包含架构决策、方案演进、废弃方案或关键约束
- **THEN** 聚合后的`design.md`包含最终设计、演进动机和保留约束
- **AND** 聚合后的`design.md`以最终设计为准处理相互冲突的历史方案

#### Scenario: 最终契约进入 specs

- **WHEN** 输入归档包含`specs/<capability>/spec.md`
- **THEN** 聚合后的`specs/<capability>/spec.md`包含对应能力的最终需求和验收场景
- **AND** 同名能力的多份规范被语义合并而不是重复拼接

### Requirement: 聚合 tasks 必须保留维护证据摘要

`lina-archive-consolidate` SHALL 将聚合归档中的`tasks.md`写成维护摘要，保留未来排障和审查仍有价值的证据，并裁剪低价值执行流水。

#### Scenario: 保留反馈闭环

- **WHEN** 输入归档的`tasks.md`包含`FB-`编号、用户反馈、根因、修复说明或回归验证
- **THEN** 聚合后的`tasks.md`包含反馈闭环摘要
- **AND** 摘要至少保留反馈主题、根因或合理假设、最终修复方向和验证结论

#### Scenario: 保留治理影响

- **WHEN** 输入归档的`tasks.md`包含`i18n`、缓存一致性、数据权限、DI、开发工具跨平台、测试策略或审查结论
- **THEN** 聚合后的`tasks.md`包含对应治理影响摘要
- **AND** 摘要必须区分有影响、无影响和无法确认的情况

#### Scenario: 裁剪低价值流水

- **WHEN** 输入归档的`tasks.md`包含普通 checklist、重复验证命令、逐文件搬迁清单或已被`design.md`和`specs/`覆盖的执行流水
- **THEN** 聚合后的`tasks.md`可以合并或裁剪这些内容
- **AND** 裁剪不得移除反馈、根因、关键验证、审查结论或治理影响判断

### Requirement: 清理原始归档前必须通过语义覆盖门禁

`lina-archive-consolidate` SHALL 只有在确认聚合输出覆盖输入归档的高价值语义后，才允许清理本次参与聚合且符合日期前缀规则的原始归档目录。

#### Scenario: 覆盖门禁通过后清理

- **WHEN** 技能完成聚合归档写入
- **AND** 每个输入归档的背景、设计决策、增量规范、反馈闭环、验证证据和审查治理影响均已进入聚合输出或被明确判定为不存在
- **THEN** 技能可以清理本次输入集合中符合日期前缀规则的原始归档目录
- **AND** 技能在最终报告中列出已清理目录和语义覆盖验证结果

#### Scenario: 覆盖门禁失败时保留

- **WHEN** 技能无法确认某个输入归档的高价值语义已经进入聚合输出
- **THEN** 技能不得删除该输入归档目录
- **AND** 技能必须在最终报告中说明未清理目录和阻断原因

### Requirement: 聚合报告必须说明压缩结果

`lina-archive-consolidate` SHALL 在最终报告中说明摘要压缩结果，支持维护者审查本次聚合是否安全。

#### Scenario: 输出压缩报告

- **WHEN** 技能完成归档聚合或压缩
- **THEN** 最终报告包含输入归档数量、输出聚合目录、已清理目录、保留的高价值信息类别、裁剪的低价值信息类别、未压缩或未清理原因和验证结果
- **AND** 若无人值守流程中无法生成可信报告，技能必须失败而不是静默清理原始归档
