// Корневой компонент origin-фронта.
//
// План 72 Ф.1 — содержал Bootstrap-заглушку.
// План 72 Ф.2 Spring 1 — переключён на AppRouter с 7 главными
// игровыми экранами (Main, Constructions, Research, Shipyard, Galaxy,
// Mission, Empire) и заглушкой /login.

import { AppRouter } from './router';

export function App() {
  return <AppRouter />;
}
