// API coverage: проверяет, что каждый эндпоинт из api/openapi.yaml вызывается
// из frontend/src/. Если эндпоинт висит без UI — падаем с кодом выхода 1,
// если он не в whitelist.
//
// Запуск: `npm run api:coverage` из frontend/.
//
// Stack: tsx + node:fs + yaml. Без сборки, без deps для браузера.

import { readFileSync, readdirSync, statSync, existsSync } from 'node:fs';
import { join, relative, resolve } from 'node:path';
import { parse as parseYaml } from 'yaml';

const ROOT = resolve(__dirname, '..');
const REPO = resolve(ROOT, '..');
const OPENAPI = resolve(REPO, 'api/openapi.yaml');
const SRC_DIR = resolve(ROOT, 'src');
const WHITELIST_FILE = resolve(ROOT, 'scripts/api-coverage.whitelist.txt');

const HTTP_METHODS = new Set(['get', 'post', 'put', 'patch', 'delete']);

interface Endpoint {
  method: string;
  path: string; // как в openapi, с {param}
}

function loadEndpoints(): Endpoint[] {
  const raw = readFileSync(OPENAPI, 'utf-8');
  const doc = parseYaml(raw) as { paths: Record<string, Record<string, unknown>> };
  const result: Endpoint[] = [];
  for (const [pathKey, ops] of Object.entries(doc.paths ?? {})) {
    if (!pathKey.startsWith('/api/')) continue; // служебные (healthz) пропускаем
    for (const method of Object.keys(ops)) {
      if (HTTP_METHODS.has(method.toLowerCase())) {
        result.push({ method: method.toUpperCase(), path: pathKey });
      }
    }
  }
  return result;
}

// Параметрическую часть OpenAPI-пути ({id}) превращаем в regex-фрагмент,
// который ищем среди литералов в TS. Сами вызовы могут быть
// `api.get(`/api/planets/${id}/buildings`)` → заменяем ${…} / любые
// алфанумы на универсальный маркер.
function pathToMatcher(path: string): RegExp {
  // /api/planets/{id}/buildings/queue/{taskId}
  const pattern = path
    .replace(/\{[^}]+\}/g, '[^\'"`\\s]+?') // {id} → non-ws
    .replace(/\//g, '\\/');
  return new RegExp(pattern);
}

function listTsFiles(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    const st = statSync(full);
    if (st.isDirectory()) {
      if (entry === 'node_modules' || entry === 'schema.d.ts') continue;
      listTsFiles(full, out);
    } else if (entry.endsWith('.ts') || entry.endsWith('.tsx')) {
      out.push(full);
    }
  }
  return out;
}

function readAllSources(): string {
  const files = listTsFiles(SRC_DIR);
  return files.map((f) => readFileSync(f, 'utf-8')).join('\n---FILE---\n');
}

function loadWhitelist(): Set<string> {
  if (!existsSync(WHITELIST_FILE)) return new Set();
  return new Set(
    readFileSync(WHITELIST_FILE, 'utf-8')
      .split('\n')
      .map((l) => l.trim())
      .filter((l) => l && !l.startsWith('#')),
  );
}

function endpointKey(e: Endpoint): string {
  return `${e.method} ${e.path}`;
}

function main(): void {
  const endpoints = loadEndpoints();
  const src = readAllSources();
  const whitelist = loadWhitelist();

  const missing: Endpoint[] = [];
  for (const ep of endpoints) {
    if (whitelist.has(endpointKey(ep))) continue;
    const re = pathToMatcher(ep.path);
    if (!re.test(src)) {
      missing.push(ep);
    }
  }

  console.log(`api-coverage: checked ${endpoints.length} endpoints`);
  console.log(`  whitelisted: ${whitelist.size}`);
  console.log(`  missing UI:  ${missing.length}`);

  if (missing.length > 0) {
    console.log('\nEndpoints without UI calls:');
    for (const ep of missing) {
      console.log(`  ${endpointKey(ep)}`);
    }
    console.log(
      `\nAdd a UI caller, or whitelist the endpoint in ${relative(REPO, WHITELIST_FILE)}`,
    );
    process.exit(1);
  }
  console.log('\nOK — all endpoints have UI coverage.');
}

main();
