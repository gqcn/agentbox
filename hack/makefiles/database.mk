# LinaPro Database Commands
# LinaPro 数据库指令
# =====================

# Initialize the backend database with schema and required seed data.
# The backend dispatches by database.default.link and currently supports PostgreSQL.
# 初始化后端数据库表结构和系统必需的种子数据。
# 后端会按 database.default.link 自动分发，目前仅支持 PostgreSQL 方言。
## init: Initialize the database with DDL and seed data only
.PHONY: init
init:
	@$(LINACTL) init confirm=$(confirm) $(if $(rebuild),rebuild=$(rebuild),)

# Load optional mock data for local demos and development verification.
# Mock loading uses the same database.default.link dialect and requires init first.
# 加载用于本地演示和开发验证的可选 Mock 数据。
# Mock 加载使用同一个 database.default.link 方言，并要求先完成 init。
## mock: Load mock demo data after init
.PHONY: mock
mock:
	@$(LINACTL) mock confirm=$(confirm)

# Generate GoFrame controller scaffolding. Defaults to the host backend; use
# p=<plugin-id> for an official plugin or dir=<backend-dir> for an explicit target.
# 生成 GoFrame 控制器骨架。默认面向宿主后端；可通过 p=<plugin-id> 指定官方插件，
# 或通过 dir=<backend-dir> 指定显式目标。
## ctrl: Generate GoFrame controller scaffolding
.PHONY: ctrl
ctrl:
	@$(LINACTL) ctrl $(if $(p),p="$(p)",) $(if $(dir),dir="$(dir)",)

# Generate DAO/DO/Entity files. Defaults to the host backend; use p=<plugin-id>
# for an official plugin or dir=<backend-dir> for an explicit target.
# 生成 DAO/DO/Entity 文件。默认面向宿主后端；可通过 p=<plugin-id> 指定官方插件，
# 或通过 dir=<backend-dir> 指定显式目标。
## dao: Generate GoFrame DAO/DO/Entity files
.PHONY: dao
dao:
	@$(LINACTL) dao $(if $(p),p="$(p)",) $(if $(dir),dir="$(dir)",)
