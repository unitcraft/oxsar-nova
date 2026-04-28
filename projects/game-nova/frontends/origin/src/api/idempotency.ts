// Idempotency-Key генератор для origin-фронта (план 72 Ф.2 Spring 1).
//
// nova-API ожидает Idempotency-Key для всех мутаций (POST/PUT/PATCH/DELETE)
// — см. ТЗ §16.10, R9. crypto.randomUUID() есть везде в современных
// браузерах (Vite-target ES2022); для тестов в Node 20+ тоже доступен
// в globalThis.crypto.

export function newIdempotencyKey(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  // Fallback: timestamp + random — не криптографически стойкий, но для
  // anti-double-submit достаточно. В браузерах используется только если
  // crypto.randomUUID не поддерживается (теоретически старые версии).
  const ts = Date.now().toString(36);
  const rnd = Math.random().toString(36).slice(2, 10);
  return `org-${ts}-${rnd}`;
}
