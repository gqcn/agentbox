import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import { fileURLToPath, URL } from "node:url";

const pluginID = "john-ai-agentbox";
const pluginVersion = "0.1.0";
const publicBase = `/x-assets/${pluginID}/${pluginVersion}/`;
const manualChunkPackages = [
  {
    name: "vendor-markdown-diagrams",
    packages: [
      "@braintree/sanitize-url",
      "@iconify/utils",
      "@mermaid-js/parser",
      "@upsetjs/venn.js",
      "cytoscape",
      "cytoscape-cose-bilkent",
      "cytoscape-fcose",
      "d3",
      "d3-sankey",
      "dagre-d3-es",
      "dayjs",
      "dompurify",
      "es-toolkit",
      "katex",
      "khroma",
      "marked",
      "mermaid",
      "roughjs",
      "stylis",
      "ts-dedent",
      "uuid",
    ],
  },
  {
    name: "vendor-react",
    packages: [
      "react",
      "react-dom",
      "@base-ui/react",
      "lucide-react",
      "clsx",
      "tailwind-merge",
      "class-variance-authority",
    ],
  },
  {
    name: "vendor-monaco",
    packages: ["@monaco-editor/react"],
  },
  {
    name: "vendor-xterm",
    packages: ["@xterm/xterm", "@xterm/addon-fit"],
  },
  {
    name: "vendor-table",
    packages: ["@tanstack/react-table", "@tanstack/react-form", "zod"],
  },
  {
    name: "vendor-panels",
    packages: ["react-resizable-panels"],
  },
];

export default defineConfig({
  base: publicBase,
  publicDir: false,
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    port: 5180,
    strictPort: true,
    proxy: {
      "/x/john-ai-agentbox/api/v1": {
        target: "http://localhost:8000",
        ws: true,
      },
    },
  },
  build: {
    outDir: "public",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks,
      },
    },
  },
});

function manualChunks(id: string) {
  const normalizedId = id.split("\\").join("/");
  if (
    normalizedId.includes(
      "/node_modules/monaco-editor/esm/vs/basic-languages/",
    )
  ) {
    return "monaco-basic-languages";
  }
  if (normalizedId.includes("/node_modules/d3-")) {
    return "vendor-markdown-diagrams";
  }
  for (const chunk of manualChunkPackages) {
    if (
      chunk.packages.some((packageName) =>
        isNodePackage(normalizedId, packageName),
      )
    ) {
      return chunk.name;
    }
  }
}

function isNodePackage(id: string, packageName: string) {
  return (
    id.includes(`/node_modules/${packageName}/`) ||
    id.includes(`/node_modules/${packageName}.js`)
  );
}
