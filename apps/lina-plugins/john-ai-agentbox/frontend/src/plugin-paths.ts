// This file centralizes plugin portal, API, WebSocket, and public asset base
// paths so the frontend never hard-codes root-level production endpoints.

const pluginID = "john-ai-agentbox";

export const pluginApiBasePath = `/x/${pluginID}/api/v1`;
export const pluginApiSessionPath = `${pluginApiBasePath}/auth/session`;
export const pluginApiSessionsPath = `${pluginApiBasePath}/auth/sessions`;
export const pluginWebSocketBasePath = `${pluginApiBasePath}/ws`;

export function pluginWebSocketURL(path: string) {
  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  return `${protocol}://${window.location.host}${pluginWebSocketBasePath}${path}`;
}
