// Router origin-фронта (план 72 Ф.2 Spring 1 + Ф.3 Spring 2).
//
// react-router v6 BrowserRouter.
//
// Spring 1 (Ф.2):
//   /                          → S-001 Main
//   /constructions/:planetId?  → S-002 Constructions
//   /research/:planetId?       → S-003 Research
//   /shipyard/:planetId?       → S-004 Shipyard
//   /galaxy/:g?/:s?            → S-005 Galaxy
//   /mission/:planetId?        → S-006 Mission
//   /empire                    → S-007 Empire
//
// Spring 2 ч.1 (Ф.3):
//   /alliance                  → S-008 Overview (lobby для не-членов)
//   /alliance/list             → S-009/S-019 Search
//   /alliance/create           → S-010 Create
//   /alliance/me               → S-008 Member-view
//   /alliance/members          → S-012 Member list
//   /alliance/manage           → S-018 Settings
//   /alliance/descriptions     → S-014 Descriptions (3 textarea)
//   /alliance/ranks            → S-013 Ranks management
//   /alliance/diplomacy        → S-015 Diplomacy
//   /alliance/audit            → S-016 Audit log
//   /alliance/transfer         → S-017 Transfer leadership
//   /alliance/:id              → S-011 External alliance page + apply
//
// Spring 2 ч.2 (Ф.3):
//   /resource-market           → S-020 Resource exchange
//   /market                    → S-015 Artefact market (legacy EXT_MODE)
//   /repair                    → S-048 Repair hangar
//   /battlestats               → S-017 Battle stats
//   /fleet-operations          → S-024 Fleet operations
//
// Spring 3 (Ф.4):
//   /artefacts                 → S-013 Artefacts (мой инвентарь)
//   /artefact/:id              → S-014 ArtefactInfo (каталог-описание)
//   /building/:type            → S-018 BuildingInfo
//   /unit/:type                → S-019 UnitInfo
//   /techtree                  → S-021 Techtree
//   /records                   → S-024 Records
//   /ranking                   → S-023 Ranking + публичная статистика (S-032)
//
//   /login                     → placeholder
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
import { AllianceOverviewScreen } from './features/alliance/AllianceOverviewScreen';
import { AllianceListScreen } from './features/alliance/AllianceListScreen';
import { AllianceCreateScreen } from './features/alliance/AllianceCreateScreen';
import { AllianceMyScreen } from './features/alliance/AllianceMyScreen';
import { AllianceMembersScreen } from './features/alliance/AllianceMembersScreen';
import { AllianceManageScreen } from './features/alliance/AllianceManageScreen';
import { AllianceDescriptionsScreen } from './features/alliance/AllianceDescriptionsScreen';
import { AllianceRanksScreen } from './features/alliance/AllianceRanksScreen';
import { AllianceDiplomacyScreen } from './features/alliance/AllianceDiplomacyScreen';
import { AllianceAuditLogScreen } from './features/alliance/AllianceAuditLogScreen';
import { AllianceTransferLeadershipScreen } from './features/alliance/AllianceTransferLeadershipScreen';
import { AlliancePageScreen } from './features/alliance/AlliancePageScreen';
import { ResourceMarketScreen } from './features/resource-market/ResourceMarketScreen';
import { MarketScreen } from './features/market/MarketScreen';
import { RepairScreen } from './features/repair/RepairScreen';
import { BattleStatsScreen } from './features/battlestats/BattleStatsScreen';
import { FleetOperationsScreen } from './features/fleet-operations/FleetOperationsScreen';
import { ArtefactsScreen } from './features/artefacts/ArtefactsScreen';
import { ArtefactInfoScreen } from './features/artefacts/ArtefactInfoScreen';
import { BuildingInfoScreen } from './features/info/BuildingInfoScreen';
import { UnitInfoScreen } from './features/info/UnitInfoScreen';
import { TechtreeScreen } from './features/techtree/TechtreeScreen';
import { RecordsScreen } from './features/records/RecordsScreen';
import { RankingScreen } from './features/ranking/RankingScreen';
// Spring 4 ч.1 (Ф.5) — communication / notes / search / settings
import { FriendsScreen } from './features/friends/FriendsScreen';
import {
  MessagesScreen,
  MessageComposeScreen,
} from './features/messages/MessagesScreen';
import {
  ChatGlobalScreen,
  ChatAllyScreen,
} from './features/chat/ChatScreen';
import { NotepadScreen } from './features/notepad/NotepadScreen';
import { SearchScreen } from './features/search/SearchScreen';
import { SettingsScreen } from './features/settings/SettingsScreen';

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

          {/* Spring 2 ч.1 — alliance (12 экранов, 11 React-компонентов) */}
          <Route path="/alliance" element={<AllianceOverviewScreen />} />
          <Route path="/alliance/list" element={<AllianceListScreen />} />
          <Route path="/alliance/create" element={<AllianceCreateScreen />} />
          <Route path="/alliance/me" element={<AllianceMyScreen />} />
          <Route path="/alliance/members" element={<AllianceMembersScreen />} />
          <Route path="/alliance/manage" element={<AllianceManageScreen />} />
          <Route
            path="/alliance/descriptions"
            element={<AllianceDescriptionsScreen />}
          />
          <Route path="/alliance/ranks" element={<AllianceRanksScreen />} />
          <Route
            path="/alliance/diplomacy"
            element={<AllianceDiplomacyScreen />}
          />
          <Route path="/alliance/audit" element={<AllianceAuditLogScreen />} />
          <Route
            path="/alliance/transfer"
            element={<AllianceTransferLeadershipScreen />}
          />
          <Route path="/alliance/:id" element={<AlliancePageScreen />} />

          {/* Spring 2 ч.2 — resource/market/repair/battlestats/fleet */}
          <Route path="/resource-market" element={<ResourceMarketScreen />} />
          <Route path="/market" element={<MarketScreen />} />
          <Route path="/repair" element={<RepairScreen />} />
          <Route path="/battlestats" element={<BattleStatsScreen />} />
          <Route
            path="/fleet-operations"
            element={<FleetOperationsScreen />}
          />

          {/* Spring 3 — artefacts / info / techtree / records / ranking */}
          <Route path="/artefacts" element={<ArtefactsScreen />} />
          <Route path="/artefact/:id" element={<ArtefactInfoScreen />} />
          <Route path="/building/:type" element={<BuildingInfoScreen />} />
          <Route path="/unit/:type" element={<UnitInfoScreen />} />
          <Route path="/techtree" element={<TechtreeScreen />} />
          <Route path="/records" element={<RecordsScreen />} />
          <Route path="/ranking" element={<RankingScreen />} />

          {/* Spring 4 ч.1 (Ф.5) — communication / notes / search / settings */}
          <Route path="/friends" element={<FriendsScreen />} />
          <Route path="/msg" element={<Navigate to="/msg/inbox" replace />} />
          <Route path="/msg/compose" element={<MessageComposeScreen />} />
          <Route path="/msg/:folder" element={<MessagesScreen />} />
          <Route path="/chat" element={<ChatGlobalScreen />} />
          <Route path="/chat/ally" element={<ChatAllyScreen />} />
          <Route path="/notepad" element={<NotepadScreen />} />
          <Route path="/search" element={<SearchScreen />} />
          <Route path="/settings" element={<SettingsScreen />} />

          <Route path="/login" element={<LoginPlaceholder />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </AppShell>
    </BrowserRouter>
  );
}
