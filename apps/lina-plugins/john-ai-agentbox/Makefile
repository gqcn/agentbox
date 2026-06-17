# LinaPro plugin Makefile.
# Shared code generation targets live in the repository root.

PLUGIN_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
REPO_ROOT := $(abspath $(PLUGIN_ROOT)/../../..)
PLUGIN_BUILD_STEP_1 := pnpm --dir "$(PLUGIN_ROOT)/frontend" run build

# build: Build the AgentBox frontend assets embedded by the source plugin
# build: 构建源码插件编译嵌入所需的 AgentBox 前端资产
.PHONY: build
build:
	@$(PLUGIN_BUILD_STEP_1)

include $(REPO_ROOT)/hack/makefiles/plugin.codegen.mk
