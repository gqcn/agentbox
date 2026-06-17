// Package johnaiagentbox embeds the AgentBox plugin manifest and public assets,
// then registers the source plugin with the Lina host when imported.
package johnaiagentbox

import (
	"context"
	"embed"

	agentboxroutes "john-ai-agentbox/backend/routes"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/pluginhost"
)

const pluginID = "john-ai-agentbox"

// embeddedFiles contains only the plugin lifecycle resources and declared
// public frontend assets that the host is allowed to discover and serve.
//
//go:embed plugin.yaml backend/plugin.go manifest/config/*.yaml manifest/sql/*.sql manifest/sql/uninstall/*.sql frontend/dist/**/* frontend/dist/*
var embeddedFiles embed.FS

func init() {
	plugin := pluginhost.NewSourcePlugin(pluginID)
	plugin.Assets().UseEmbeddedFiles(embeddedFiles)
	if err := plugin.HTTP().RegisterRoutes(
		pluginhost.ExtensionPointHTTPRouteRegister,
		pluginhost.CallbackExecutionModeBlocking,
		func(ctx context.Context, registrar pluginhost.HTTPRegistrar) error {
			return agentboxroutes.Register(ctx, registrar, embeddedFiles)
		},
	); err != nil {
		panic(gerror.Wrapf(err, "register source plugin routes %s", pluginID))
	}
	if err := pluginhost.RegisterSourcePlugin(plugin); err != nil {
		panic(gerror.Wrapf(err, "register source plugin %s", pluginID))
	}
}
