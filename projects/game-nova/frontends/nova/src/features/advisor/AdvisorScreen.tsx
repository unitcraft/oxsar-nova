// AdvisorScreen — экран ИИ-советника (план 06 §F.3 + 06.1).
//
// Две вкладки:
//   - «AI» (LLM) — свободные вопросы к Claude/Ollama (план 06 F.3,
//     уже реализован на бэке: POST /api/ai-advisor/ask).
//   - «Автопилот» (rule-based) — топ-3 рекомендованных хода с
//     подтверждением (план 06.1).
//
// Автопилот:
//   - Кнопка «Запустить» → POST /api/ai-advisor/autopilot/advise
//     с {strategy} → {job_id}. Списывает 2 кредита.
//   - Polling GET /api/ai-advisor/autopilot/result/{job_id} раз в 2 сек,
//     пока status='ready'. Затем останавливаемся, показываем 3 карточки.
//   - На каждой карточке кнопка «Выполнить» → POST /execute → backend
//     создаёт игровое событие; тост, навигация в зависимости от категории.
//   - История последних N запросов (на сервере не храним отдельный
//     эндпоинт, но Result даёт состояние любого jobID — для истории
//     нужен HISTORY-эндпоинт; пока история показывается по последнему
//     активному job в localStorage).

import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { api } from '../../api/client';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

type Strategy = 'economy' | 'military' | 'defense' | 'expansion' | 'auto';

const STRATEGIES: Strategy[] = ['economy', 'military', 'defense', 'expansion', 'auto'];

type Recommendation = {
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

type AutopilotResult = {
  job_id: string;
  status: 'pending' | 'ready' | 'executed' | 'expired';
  strategy: Strategy;
  recs: Recommendation[];
};

const STORAGE_KEY = 'autopilot:strategy';
const STORAGE_JOB = 'autopilot:lastJob';

function loadStrategy(): Strategy {
  if (typeof window === 'undefined') return 'auto';
  const v = window.localStorage.getItem(STORAGE_KEY);
  if (v && (STRATEGIES as string[]).includes(v)) return v as Strategy;
  return 'auto';
}

export function AdvisorScreen() {
  const { t } = useTranslation('advisor');
  const [tab, setTab] = useState<'ai' | 'autopilot'>('autopilot');

  return (
    <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 12 }}>
      <h2 style={{ margin: 0 }}>{t('title')}</h2>

      <div className="ox-tabs" role="tablist" style={{ display: 'flex', gap: 4 }}>
        <button
          className="ox-btn"
          role="tab"
          aria-pressed={tab === 'ai'}
          onClick={() => setTab('ai')}
        >
          {t('tabAI')}
        </button>
        <button
          className="ox-btn"
          role="tab"
          aria-pressed={tab === 'autopilot'}
          onClick={() => setTab('autopilot')}
        >
          {t('tabAutopilot')}
        </button>
      </div>

      {tab === 'ai' && <AITab />}
      {tab === 'autopilot' && <AutopilotTab />}
    </div>
  );
}

function AITab() {
  const { t } = useTranslation('advisor');
  const toast = useToast();
  const [question, setQuestion] = useState('');
  const [answer, setAnswer] = useState('');
  const [model, setModel] = useState('claude-haiku-4-5-20251001');

  const ask = useMutation({
    mutationFn: () =>
      api.post<{ answer: string; credits_used: number }>('/api/ai-advisor/ask', { model, question }),
    onSuccess: (res) => {
      setAnswer(res.answer);
      if (res.credits_used > 0) {
        toast.show('info', t('aiAskOK'), `−${res.credits_used} cr`);
      }
    },
    onError: (err) => {
      toast.show('danger', t('aiAskFail'), err instanceof Error ? err.message : '');
    },
  });

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>{t('aiHint')}</div>

      <select
        value={model}
        onChange={(e) => setModel(e.target.value)}
        className="ox-input"
      >
        <option value="claude-haiku-4-5-20251001">Haiku (5 cr)</option>
        <option value="claude-sonnet-4-6">Sonnet (20 cr)</option>
        <option value="claude-opus-4-7">Opus (80 cr)</option>
      </select>

      <textarea
        value={question}
        onChange={(e) => setQuestion(e.target.value)}
        placeholder={t('aiQuestionPlaceholder')}
        className="ox-input"
        rows={4}
      />

      <button
        className="ox-btn ox-btn-primary"
        onClick={() => ask.mutate()}
        disabled={!question.trim() || ask.isPending}
      >
        {ask.isPending ? t('aiThinking') : t('aiAsk')}
      </button>

      {answer && (
        <div className="ox-panel" style={{ padding: 12, whiteSpace: 'pre-wrap' }}>
          {answer}
        </div>
      )}
    </div>
  );
}

function AutopilotTab() {
  const { t } = useTranslation('advisor');
  const toast = useToast();
  const qc = useQueryClient();
  const [strategy, setStrategy] = useState<Strategy>(loadStrategy());
  const [activeJobID, setActiveJobID] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null;
    return window.localStorage.getItem(STORAGE_JOB);
  });

  useEffect(() => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, strategy);
    }
  }, [strategy]);

  const advise = useMutation({
    mutationFn: () =>
      api.post<{ job_id: string }>('/api/ai-advisor/autopilot/advise', { strategy }),
    onSuccess: (res) => {
      setActiveJobID(res.job_id);
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(STORAGE_JOB, res.job_id);
      }
      void qc.invalidateQueries({ queryKey: ['me'] }); // балaнс кредитов
    },
    onError: (err) => {
      toast.show('danger', t('autoFail'), err instanceof Error ? err.message : '');
    },
  });

  const result = useQuery({
    queryKey: ['autopilot-result', activeJobID],
    queryFn: () => api.get<AutopilotResult>(`/api/ai-advisor/autopilot/result/${activeJobID!}`),
    enabled: !!activeJobID,
    refetchInterval: (query) => {
      const data = query.state.data as AutopilotResult | undefined;
      if (!data) return 2000;
      // Останавливаем polling когда status != pending.
      return data.status === 'pending' ? 2000 : false;
    },
  });

  const execute = useMutation({
    mutationFn: (recID: string) =>
      api.post<{ event_id: string }>('/api/ai-advisor/autopilot/execute', {
        job_id: activeJobID,
        rec_id: recID,
      }),
    onSuccess: (res, recID) => {
      const rec = result.data?.recs.find((r) => r.id === recID);
      toast.show('success', t('autoExecuted'), rec?.description ?? '');
      void qc.invalidateQueries(); // полностью обновляем — событие могло затронуть всё
      // Навигация на целевой экран — по категории.
      navigateForRec(rec);
      // Сбросим job (он stale после executed).
      setActiveJobID(null);
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem(STORAGE_JOB);
      }
    },
    onError: (err) => {
      toast.show('danger', t('autoExecuteFail'), err instanceof Error ? err.message : '');
    },
  });

  const data = result.data;

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
      <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>{t('autoHint')}</div>

      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <label htmlFor="ap-strategy" style={{ fontSize: 13 }}>{t('autoStrategy')}:</label>
        <select
          id="ap-strategy"
          className="ox-input"
          value={strategy}
          onChange={(e) => setStrategy(e.target.value as Strategy)}
          style={{ width: 'auto' }}
        >
          {STRATEGIES.map((s) => (
            <option key={s} value={s}>{t(`strategy${capitalize(s)}`)}</option>
          ))}
        </select>
        <button
          className="ox-btn ox-btn-primary"
          onClick={() => advise.mutate()}
          disabled={advise.isPending || (data?.status === 'pending')}
        >
          {advise.isPending || data?.status === 'pending' ? t('autoComputing') : t('autoStart')}
        </button>
      </div>

      {data?.status === 'pending' && (
        <div className="ox-panel" style={{ padding: 12, fontStyle: 'italic' }}>
          {t('autoComputing')}…
        </div>
      )}

      {data?.status === 'ready' && data.recs.length === 0 && (
        <div className="ox-panel" style={{ padding: 12 }}>
          {t('autoEmpty')}
        </div>
      )}

      {data?.status === 'ready' && data.recs.length > 0 && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {data.recs.map((rec) => (
            <div key={rec.id} className="ox-panel" style={{ padding: 12 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 'bold', marginBottom: 4 }}>{rec.description}</div>
                  <div style={{ fontSize: 13, color: 'var(--ox-fg-muted)' }}>{rec.benefit}</div>
                  <div style={{ fontSize: 11, color: 'var(--ox-fg-muted)', marginTop: 4 }}>
                    {t('autoCategory')}: {t(`category${capitalize(rec.category)}`)} • score {rec.score.toFixed(2)}
                  </div>
                </div>
                <button
                  className="ox-btn ox-btn-primary"
                  onClick={() => execute.mutate(rec.id)}
                  disabled={execute.isPending}
                >
                  {execute.isPending ? '…' : t('autoExecute')}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function capitalize(s: string): string {
  return s.length === 0 ? s : s[0]!.toUpperCase() + s.slice(1);
}

// navigateForRec — после Execute переключаем хеш-роут на нужный экран.
// Нав-логика выбрана так, чтобы игрок мог посмотреть результат своего
// выбора в контексте.
function navigateForRec(rec: Recommendation | undefined) {
  if (!rec) return;
  switch (rec.category) {
    case 'building':
      window.location.hash = '#buildings';
      break;
    case 'research':
      window.location.hash = '#research';
      break;
    case 'mission':
      window.location.hash = '#fleet';
      break;
    case 'profession':
      window.location.hash = '#profession';
      break;
    default:
      // другие категории — остаёмся на advisor.
  }
}
