import { useMemo, useState } from 'react';
import { nameOf } from '@/api/catalog';
import { useTranslation } from '@/i18n/i18n';

type NodeKind = 'building' | 'research' | 'ship' | 'defense';

interface Requirement {
  kind: string;
  key: string;
  level: number;
  have: number;
  met: boolean;
}
interface TechNode {
  key: string;
  kind: NodeKind;
  id: number;
  current_level: number;
  unlocked: boolean;
  requirements: Requirement[];
}

// Расчёт layered layout: глубина = max(depth of reqs) + 1.
interface Positioned {
  node: TechNode;
  depth: number;
  row: number;
  x: number;
  y: number;
}

const NODE_W = 180;
const NODE_H = 52;
const GAP_X = 60;
const GAP_Y = 16;

export function TechtreeGraph({ nodes, kind }: { nodes: TechNode[]; kind: NodeKind }) {
  const { t } = useTranslation('techtreeUi');
  const { t: ti } = useTranslation('info');
  const [hover, setHover] = useState<string | null>(null);

  // Фильтрация по kind.
  const filtered = useMemo(() => nodes.filter((n) => n.kind === kind), [nodes, kind]);

  const positioned = useMemo(() => computeLayout(filtered, kind, ti), [filtered, kind, ti]);
  const byKey = useMemo(() => {
    const m = new Map<string, Positioned>();
    for (const p of positioned) m.set(p.node.key, p);
    return m;
  }, [positioned]);

  const width = Math.max(...positioned.map((p) => p.x + NODE_W), 400) + 40;
  const height = Math.max(...positioned.map((p) => p.y + NODE_H), 200) + 40;

  return (
    <div style={{ overflow: 'auto', border: '1px solid var(--ox-border)', borderRadius: 6, background: 'var(--ox-bg-panel)' }}>
      <svg width={width} height={height} style={{ display: 'block' }}>
        {/* Рёбра */}
        <g>
          {positioned.flatMap((p) =>
            p.node.requirements
              .filter((req) => req.kind === kind || (kind === 'research' && req.kind === 'research'))
              .map((req) => {
                const src = byKey.get(req.key);
                if (!src) return null;
                const color = p.node.unlocked ? 'var(--ox-success)' : req.met ? 'var(--ox-fg-muted)' : 'var(--ox-danger)';
                return (
                  <line
                    key={`${p.node.key}-${req.key}`}
                    x1={src.x + NODE_W}
                    y1={src.y + NODE_H / 2}
                    x2={p.x}
                    y2={p.y + NODE_H / 2}
                    stroke={color}
                    strokeWidth={hover === p.node.key || hover === req.key ? 2 : 1}
                    strokeOpacity={hover && hover !== p.node.key && hover !== req.key ? 0.2 : 0.6}
                  />
                );
              })
          )}
        </g>
        {/* Узлы */}
        {positioned.map((p) => (
          <g
            key={p.node.key}
            transform={`translate(${p.x}, ${p.y})`}
            onMouseEnter={() => setHover(p.node.key)}
            onMouseLeave={() => setHover(null)}
            style={{ cursor: 'default' }}
          >
            <rect
              width={NODE_W}
              height={NODE_H}
              rx={4}
              fill={p.node.current_level > 0 ? 'rgba(99,217,255,0.12)' : 'var(--ox-bg-panel)'}
              stroke={p.node.unlocked ? 'var(--ox-success)' : 'var(--ox-fg-muted)'}
              strokeWidth={hover === p.node.key ? 2 : 1}
              opacity={p.node.unlocked ? 1 : 0.7}
            />
            <text x={10} y={20} fill="var(--ox-fg)" fontSize={12} fontWeight={600}>
              {truncate(nameOf(p.node.id, ti) || p.node.key, 22)}
            </text>
            <text x={10} y={40} fill="var(--ox-fg-muted)" fontSize={10} fontFamily="var(--ox-mono)">
              {p.node.current_level > 0
                ? `${t('levelAbbr')} ${p.node.current_level}`
                : p.node.unlocked ? t('available') : t('locked')}
            </text>
          </g>
        ))}
      </svg>
    </div>
  );
}

function truncate(s: string, max: number): string {
  if (s.length <= max) return s;
  return s.slice(0, max - 1) + '…';
}

// Layered-layout: depth считается только по требованиям своего kind
// (иначе граф из исследований тянет здания — получается пересечение).
function computeLayout(nodes: TechNode[], kind: NodeKind, ti: (key: string) => string): Positioned[] {
  const depthByKey = new Map<string, number>();
  const nodeByKey = new Map<string, TechNode>();
  for (const n of nodes) nodeByKey.set(n.key, n);

  function depth(key: string, stack: Set<string>): number {
    if (depthByKey.has(key)) return depthByKey.get(key)!;
    if (stack.has(key)) return 0; // cycle guard
    stack.add(key);
    const n = nodeByKey.get(key);
    if (!n) { stack.delete(key); return 0; }
    let d = 0;
    for (const req of n.requirements) {
      if (req.kind === kind && nodeByKey.has(req.key)) {
        d = Math.max(d, depth(req.key, stack) + 1);
      }
    }
    stack.delete(key);
    depthByKey.set(key, d);
    return d;
  }

  for (const n of nodes) depth(n.key, new Set());

  // Группируем по depth.
  const byDepth = new Map<number, TechNode[]>();
  for (const n of nodes) {
    const d = depthByKey.get(n.key) ?? 0;
    if (!byDepth.has(d)) byDepth.set(d, []);
    byDepth.get(d)!.push(n);
  }
  // Сортируем внутри глубины по имени для стабильности.
  for (const arr of byDepth.values()) {
    arr.sort((a, b) => (nameOf(a.id, ti) || a.key).localeCompare(nameOf(b.id, ti) || b.key));
  }

  const out: Positioned[] = [];
  for (const [d, arr] of Array.from(byDepth.entries()).sort((a, b) => a[0] - b[0])) {
    arr.forEach((n, row) => {
      out.push({
        node: n,
        depth: d,
        row,
        x: 20 + d * (NODE_W + GAP_X),
        y: 20 + row * (NODE_H + GAP_Y),
      });
    });
  }
  return out;
}
