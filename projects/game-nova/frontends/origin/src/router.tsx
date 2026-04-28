// Router origin-фронта (план 72 Ф.2 Spring 1).
//
// react-router v6 BrowserRouter. План 72 Ф.2 фиксирует пути:
//   /                          → S-001 Main
//   /constructions/:planetId?  → S-002 Constructions
//   /research/:planetId?       → S-003 Research
//   /shipyard/:planetId?       → S-004 Shipyard
//   /galaxy/:g?/:s?            → S-005 Galaxy
//   /mission/:planetId?        → S-006 Mission
//   /empire                    → S-007 Empire
//   /login                     → placeholder (Ф.3 Spring 2)
//   *                          → redirect на /
//
// AppShell оборачивает все маршруты — 3-frame layout остаётся
// единым на всё SPA, как в legacy-PHP.

import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AppShell } from './layout/AppShell';
import { MainScreen } from './features/main/MainScreen';
import { ConstructionsScreen } from './features/constructions/ConstructionsScreen';
import { ResearchScreen } from './features/research/ResearchScreen';
import { ShipyardScreen } from './features/shipyard/ShipyardScreen';
import { GalaxyScreen } from './features/galaxy/GalaxyScreen';
import { MissionScreen } from './features/mission/MissionScreen';
import { EmpireScreen } from './features/empire/EmpireScreen';
import { LoginPlaceholder } from './features/login/LoginPlaceholder';

export function AppRouter() {
  return (
    <BrowserRouter>
      <AppShell>
        <Routes>
          <Route path="/" element={<MainScreen />} />
          <Route path="/constructions" element={<ConstructionsScreen />} />
          <Route
            path="/constructions/:planetId"
            element={<ConstructionsScreen />}
          />
          <Route path="/research" element={<ResearchScreen />} />
          <Route path="/research/:planetId" element={<ResearchScreen />} />
          <Route path="/shipyard" element={<ShipyardScreen />} />
          <Route path="/shipyard/:planetId" element={<ShipyardScreen />} />
          <Route path="/galaxy" element={<GalaxyScreen />} />
          <Route path="/galaxy/:galaxy/:system" element={<GalaxyScreen />} />
          <Route path="/mission" element={<MissionScreen />} />
          <Route path="/mission/:planetId" element={<MissionScreen />} />
          <Route path="/empire" element={<EmpireScreen />} />
          <Route path="/login" element={<LoginPlaceholder />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AppShell>
    </BrowserRouter>
  );
}
