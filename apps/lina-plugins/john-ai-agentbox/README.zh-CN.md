# John AI AgentBox

`john-ai-agentbox`是用于将外部 AgentBox 产品迁入 LinaPro 插件生命周期的源码插件。

当前范围：

| 区域 | 状态 |
|------|------|
| 插件身份 | `john-ai-agentbox` |
| 门户路由 | `/`和`/login` |
| 管理工作台 | 继续使用`/admin` |
| API 命名空间 | `/x/john-ai-agentbox/api/v1/...` |
| 公开资源 | 通过声明的`public_assets`发布`frontend/dist` |

构建：

```bash
make build
```

插件构建钩子会在宿主 Go 二进制嵌入插件资源前生成`frontend/dist`。`frontend/dist`是生成构建产物，继续由 Git 忽略。

插件首期按单语言治理，暂不启用插件`i18n`资源。
