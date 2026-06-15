# John AI AgentBox

`john-ai-agentbox` is a LinaPro source plugin for migrating the external AgentBox product into the host plugin lifecycle.

Current scope:

| Area | Status |
|------|--------|
| Plugin identity | `john-ai-agentbox` |
| Portal routes | `/` and `/login` |
| Admin workspace | Remains under `/admin` |
| API namespace | `/x/john-ai-agentbox/api/v1/...` |
| Public assets | `frontend/public` through declared `public_assets` |

The plugin is single-language for the initial migration and does not enable plugin `i18n` resources yet.
