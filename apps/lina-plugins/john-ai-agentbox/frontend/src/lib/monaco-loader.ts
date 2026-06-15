// This module keeps Monaco configured for local Vite assets before the React
// wrapper initializes. Without this, @monaco-editor/react falls back to its
// default CDN-based AMD loader at runtime.

import type * as Monaco from "monaco-editor";

export async function loadMonacoEditor() {
  const [reactMonaco, monaco] = await Promise.all([
    import("@monaco-editor/react"),
    loadLocalMonaco(),
  ]);
  reactMonaco.loader.config({ monaco });
  return { default: reactMonaco.Editor };
}

export async function loadMonacoDiffEditor() {
  const [reactMonaco, monaco] = await Promise.all([
    import("@monaco-editor/react"),
    loadLocalMonaco(),
  ]);
  reactMonaco.loader.config({ monaco });
  return { default: reactMonaco.DiffEditor };
}

async function loadLocalMonaco() {
  const monaco = await import("monaco-editor/esm/vs/editor/editor.api");
  await import("./monaco-basic-languages");
  return monaco as typeof Monaco;
}
