## MODIFIED Requirements

### Requirement: `linactl`必须内嵌 GoFrame 代码生成入口

系统 SHALL 由`linactl`直接承载宿主和插件后端的 GoFrame controller 与 DAO/DO/Entity 代码生成入口。`linactl ctrl`和`linactl dao`不得要求开发者在本机预先安装`gf`，也不得在默认开发路径中下载、安装或调用`PATH`中的外部`gf`可执行文件。`linactl ctrl`和`linactl dao`默认以`apps/lina-core`作为生成目标；当调用方显式传入插件 ID 或后端目录时，命令 MUST 在对应插件`backend/`目录中执行同一套内嵌 GoFrame 生成流程。

#### Scenario: controller 生成不依赖外部 `gf`

- **WHEN** 开发者运行`linactl ctrl`、根目录`make ctrl`、插件目录`make ctrl`或带插件目标的根目录`make ctrl`
- **THEN** 命令通过`linactl`内嵌的 GoFrame CLI module 执行`gen ctrl`
- **AND** 命令不得调用`gf`、`gf -v`、`gf install`或 GitHub release 下载地址
- **AND** 开发者不需要在`PATH`中提供`gf`可执行文件

#### Scenario: DAO 生成不依赖外部 `gf`

- **WHEN** 开发者运行`linactl dao`、根目录`make dao`、插件目录`make dao`或带插件目标的根目录`make dao`
- **THEN** 命令通过`linactl`内嵌的 GoFrame CLI module 执行`gen dao`
- **AND** 命令不得调用`gf`、`gf -v`、`gf install`或 GitHub release 下载地址
- **AND** 开发者不需要在`PATH`中提供`gf`可执行文件

### Requirement: GoFrame 代码生成必须保持目标工作目录语义

系统 SHALL 在执行内嵌 GoFrame CLI 生成命令时使用明确的目标后端目录作为工作目录，使`hack/config.yaml`、`api/`、`internal/`和`go.mod`解析结果与目标项目一致。未指定目标时，默认目标 MUST 为仓库根目录下的`apps/lina-core`；指定插件 ID 时，目标 MUST 为仓库根目录下的`apps/lina-plugins/<plugin-id>/backend`；指定目录时，目标 MUST 为该目录。任何目标目录都 MUST 包含`hack/config.yaml`，否则命令必须拒绝执行并返回清晰错误。

#### Scenario: controller 生成默认使用宿主工作目录

- **WHEN** `linactl ctrl`或根目录`make ctrl`未指定插件或目录目标
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** `api/`和`internal/controller`路径按宿主目录解析

#### Scenario: DAO 生成默认使用宿主工作目录

- **WHEN** `linactl dao`或根目录`make dao`未指定插件或目录目标
- **THEN** GoFrame CLI 在仓库根目录下的`apps/lina-core`目录中执行
- **AND** `gfcli.gen.dao`配置从宿主`hack/config.yaml`解析
- **AND** `internal/dao`、`internal/model/do`和`internal/model/entity`路径按宿主目录解析

#### Scenario: controller 生成使用插件后端工作目录

- **WHEN** 开发者在插件根目录运行`make ctrl`或在根目录运行带插件 ID 的`make ctrl`
- **THEN** GoFrame CLI 在对应`apps/lina-plugins/<plugin-id>/backend`目录中执行
- **AND** `api/`和`internal/controller`路径按插件后端目录解析

#### Scenario: DAO 生成使用插件后端工作目录

- **WHEN** 开发者在插件根目录运行`make dao`或在根目录运行带插件 ID 的`make dao`
- **THEN** GoFrame CLI 在对应`apps/lina-plugins/<plugin-id>/backend`目录中执行
- **AND** `gfcli.gen.dao`配置从插件`backend/hack/config.yaml`解析
- **AND** `internal/dao`、`internal/model/do`和`internal/model/entity`路径按插件后端目录解析

#### Scenario: 缺少 GoFrame 配置的目标目录被拒绝

- **WHEN** 开发者将代码生成目标指向没有`hack/config.yaml`的目录
- **THEN** `linactl`拒绝执行 GoFrame 生成命令
- **AND** 错误消息说明目标目录缺少`hack/config.yaml`

### Requirement: 公开开发命令必须保持稳定

系统 SHALL 保持`make image`、`make image.build`、`make wasm`、`make i18n.check`、`make ctrl`、`make dao`、`linactl image`、`linactl image.build`、`linactl wasm`、`linactl i18n.check`、`linactl ctrl`和`linactl dao`的公开入口稳定。工具实现迁移不得要求开发者改用新的命令名称。宿主和插件目录的本地`Makefile` MUST 仅作为薄转发层调用仓库统一工具链，不得重新承载外部`gf`安装、代码生成业务逻辑或与根`Makefile`重复的构建治理逻辑。

#### Scenario: Make 入口继续调用`linactl`

- **WHEN** 开发者运行`make image`、`make image.build`、`make wasm`、`make i18n.check`、`make ctrl`或`make dao`
- **THEN** 根`Makefile`转发到对应`linactl`命令
- **AND** 开发者不需要直接调用旧独立工具目录或外部`gf`

#### Scenario: 宿主本地 Makefile 只保留薄转发层

- **WHEN** 开发者查看`apps/lina-core/Makefile`
- **THEN** 该文件只保留指向仓库统一工具链的宿主相关薄入口
- **AND** 不包含安装外部`gf`、更新外部`gf`或重复维护根构建治理逻辑的目标

#### Scenario: 插件根目录提供一致代码生成入口

- **WHEN** 开发者查看任一官方插件根目录`apps/lina-plugins/<plugin-id>/Makefile`
- **THEN** 该文件提供`ctrl`和`dao`目标
- **AND** 两个目标转发到仓库统一`linactl`并以该插件`backend/`作为生成目标
- **AND** 代码生成目标逻辑集中维护在根目录`hack/makefiles/plugin.codegen.mk`
- **AND** 插件根目录`Makefile`不得硬编码`apps/lina-plugins/<plugin-id>/backend`
