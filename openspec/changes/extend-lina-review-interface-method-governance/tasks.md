## 1. 规范与审查规则

- [x] 1.1 更新 `AGENTS.md` 服务层接口规范，要求后端接口方法定义唯一且语义清晰。
- [x] 1.2 更新 `lina-review` 技能，新增后端接口方法定义审查规则。
- [x] 1.3 增加 `backend-conformance` 增量规范，覆盖重复方法、近义方法、歧义方法和兼容期说明。

## 2. 验证与审查

- [x] 2.1 运行 `openspec validate extend-lina-review-interface-method-governance --strict`。
- [x] 2.2 运行静态检索，确认 `AGENTS.md` 与 `lina-review` 均包含接口方法定义治理规则。
- [x] 2.3 记录 i18n、缓存一致性、数据权限、开发工具跨平台和测试影响评估，并执行 `lina-review` 审查。

## 3. 执行记录

- i18n 影响评估：本变更只调整治理规范和审查技能说明，不新增、修改或删除运行时用户可见文案、菜单、路由、API 文档源文本、插件清单或语言包资源。
- 缓存一致性影响评估：本变更不新增或修改生产缓存、缓存键、失效路径、跨实例协调或运行时状态。
- 数据权限影响评估：本变更不新增或修改数据操作接口、服务数据访问路径、插件 host service 数据访问或权限边界。
- 开发工具跨平台影响评估：本变更不新增或修改开发工具、脚本、CI 命令实现或默认测试入口。
- 测试策略：项目治理类反馈不新增单元测试或 E2E；使用 OpenSpec 严格校验、静态检索和审查结论作为验证证据。
- Review：已按 `lina-review` 口径完成审查。审查范围限定为 `AGENTS.md`、`.agents/skills/lina-review/SKILL.md` 和 `openspec/changes/extend-lina-review-interface-method-governance`；未发现阻塞问题。该变更不涉及 Go 生产代码、REST API、后端数据权限、运行时缓存、前端 UI 或 i18n 资源，Go 编译门禁和 E2E 不适用。
