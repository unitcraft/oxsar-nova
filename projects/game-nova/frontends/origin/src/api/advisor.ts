// API-модуль AI-советника / Автопилота для origin-фронта (план 06.1).
//
// Endpoints (см. openapi.yaml):
//   POST /api/ai-advisor/ask                          — LLM-вопрос
//   GET  /api/ai-advisor/estimate?model=…             — оценка стоимости
//   POST /api/ai-advisor/autopilot/advise              — поставить задание
//   GET  /api/ai-advisor/autopilot/result/{jobID}      — polling результата
//   POST /api/ai-advisor/autopilot/execute             — выполнить рек
//   GET  /api/ai-advisor/autopilot/history             — последние 20

import { api } from './client';
import { newIdempotencyKey } from './idempotency';

export type Strategy =
  | 'economy'
  | 'military'
  | 'defense'
  | 'expansion'
  | 'auto';

export const STRATEGIES: Strategy[] = [
  'economy',
  'military',
  'defense',
  'expansion',
  'auto',
];

export type AdvisorAskResponse = {
  answer: string;
  credits_used: number;
};

export function askAdvisor(
  model: string,
  question: string,
): Promise<AdvisorAskResponse> {
  return api.post<AdvisorAskResponse>('/api/ai-advisor/ask', {
    model,
    question,
  });
}

export type AutopilotRecommendation = {
  id: string;
  category: string;
  action_type: string;
  planet_id?: string;
  unit_id?: number;
  params?: Record<string, unknown>;
  score: number;
  description: string;
  benefit: string;
};

export type AutopilotResult = {
  job_id: string;
  status: 'pending' | 'ready' | 'executed' | 'expired';
  strategy: Strategy;
  recs: AutopilotRecommendation[];
};

export function startAutopilot(strategy: Strategy): Promise<{ job_id: string }> {
  return api.post<{ job_id: string }>(
    '/api/ai-advisor/autopilot/advise',
    { strategy },
    { idempotencyKey: newIdempotencyKey() },
  );
}

export function fetchAutopilotResult(jobID: string): Promise<AutopilotResult> {
  return api.get<AutopilotResult>(
    `/api/ai-advisor/autopilot/result/${encodeURIComponent(jobID)}`,
  );
}

export function executeAutopilot(
  jobID: string,
  recID: string,
): Promise<{ event_id: string }> {
  return api.post<{ event_id: string }>(
    '/api/ai-advisor/autopilot/execute',
    { job_id: jobID, rec_id: recID },
    { idempotencyKey: newIdempotencyKey() },
  );
}

export type AutopilotHistoryItem = {
  job_id: string;
  strategy: string;
  status: string;
  created_at: string;
  credits: number;
  executed_description?: string;
};

export function fetchAutopilotHistory(): Promise<{
  items: AutopilotHistoryItem[];
}> {
  return api.get<{ items: AutopilotHistoryItem[] }>(
    '/api/ai-advisor/autopilot/history',
  );
}
