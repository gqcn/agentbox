const filenameLanguages: Record<string, string> = {
  '.dockerignore': 'ini',
  '.env': 'ini',
  '.gitattributes': 'ini',
  '.gitignore': 'ini',
  containerfile: 'dockerfile',
  dockerfile: 'dockerfile',
  makefile: 'shell',
};

const extensionLanguages: Record<string, string> = {
  '.abap': 'abap',
  '.aes': 'aes',
  '.azcli': 'azcli',
  '.bash': 'shell',
  '.bat': 'bat',
  '.bicep': 'bicep',
  '.c': 'c',
  '.cc': 'cpp',
  '.cjs': 'javascript',
  '.clj': 'clojure',
  '.cljs': 'clojure',
  '.cljc': 'clojure',
  '.cmd': 'bat',
  '.coffee': 'coffeescript',
  '.cpp': 'cpp',
  '.cs': 'csharp',
  '.cshtml': 'razor',
  '.css': 'css',
  '.cts': 'typescript',
  '.cxx': 'cpp',
  '.dart': 'dart',
  '.dockerfile': 'dockerfile',
  '.edn': 'clojure',
  '.ex': 'elixir',
  '.exs': 'elixir',
  '.fs': 'fsharp',
  '.fsx': 'fsharp',
  '.gemspec': 'ruby',
  '.go': 'go',
  '.graphql': 'graphql',
  '.gql': 'graphql',
  '.h': 'c',
  '.handlebars': 'handlebars',
  '.hbs': 'handlebars',
  '.hcl': 'hcl',
  '.hh': 'cpp',
  '.hpp': 'cpp',
  '.htm': 'html',
  '.html': 'html',
  '.hxx': 'cpp',
  '.ini': 'ini',
  '.java': 'java',
  '.jl': 'julia',
  '.js': 'javascript',
  '.json': 'json',
  '.jsonc': 'json',
  '.jsx': 'javascript',
  '.kt': 'kotlin',
  '.kts': 'kotlin',
  '.less': 'less',
  '.liquid': 'liquid',
  '.lua': 'lua',
  '.m': 'objective-c',
  '.markdown': 'markdown',
  '.md': 'markdown',
  '.mdx': 'mdx',
  '.mjs': 'javascript',
  '.ml': 'fsharp',
  '.mli': 'fsharp',
  '.mts': 'typescript',
  '.pas': 'pascal',
  '.php': 'php',
  '.pl': 'perl',
  '.pm': 'perl',
  '.properties': 'ini',
  '.proto': 'proto',
  '.ps1': 'powershell',
  '.psd1': 'powershell',
  '.psm1': 'powershell',
  '.py': 'python',
  '.r': 'r',
  '.rb': 'ruby',
  '.redis': 'redis',
  '.rmd': 'markdown',
  '.rs': 'rust',
  '.rst': 'restructuredtext',
  '.scala': 'scala',
  '.sbt': 'scala',
  '.scss': 'scss',
  '.sh': 'shell',
  '.sol': 'sol',
  '.sql': 'sql',
  '.swift': 'swift',
  '.tf': 'hcl',
  '.tfvars': 'hcl',
  '.toml': 'ini',
  '.ts': 'typescript',
  '.tsx': 'typescript',
  '.twig': 'twig',
  '.vb': 'vb',
  '.vue': 'html',
  '.wgsl': 'wgsl',
  '.xml': 'xml',
  '.yaml': 'yaml',
  '.yml': 'yaml',
  '.zsh': 'shell',
};

export function languageFromPath(pathValue: string) {
  const name = baseName(pathValue).toLowerCase();
  if (name.endsWith('.html.liquid')) return 'liquid';
  if (name.startsWith('.env.')) return 'ini';
  if (name === 'go.mod' || name === 'go.sum') return 'go';
  if (name === 'package.json' || name === 'tsconfig.json' || name === 'jsconfig.json') return 'json';
  if (filenameLanguages[name]) return filenameLanguages[name];
  return extensionLanguages[extensionName(name)] ?? 'plaintext';
}

function baseName(pathValue: string) {
  const normalized = pathValue.replace(/\\/g, '/').replace(/\/+$/, '');
  const index = normalized.lastIndexOf('/');
  return index >= 0 ? normalized.slice(index + 1) : normalized;
}

function extensionName(name: string) {
  const index = name.lastIndexOf('.');
  return index > 0 ? name.slice(index) : '';
}
