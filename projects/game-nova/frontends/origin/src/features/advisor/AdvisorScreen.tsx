// Экран AI-советника / Автопилота для origin-фронта (план 06.1).
//
// В legacy oxsar2 советника не было — это новая фича origin-вселенной.
// Стилистика максимально близка к legacy: table.ntable заголовок,
// «true»/«false» классы для статусов, без эмодзи в кнопках.
//
// Две вкладки: AI (LLM-вопросы) и Автопилот (rule-based топ-3).
//
// Бэкенд тот же что у nova-фронта (план 06 §F.3 + 06.1):
//   POST /api/ai-advisor/ask, POST /api/ai-advisor/autopilot/advise, …

import { useEffect, useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import {
  askAdvisor,
  executeAutopilot,
  fetchAutopilotHistory,
  fetchAutopilotResult,
  startAutopilot,
  STRATEGIES,
  type AutopilotRecommendation,
  type AutopilotResult,
  type Strategy,
} from '@/api/advisor';

const STORAGE_KEY = 'autopilot:strategy';
const STORAGE_JOB = 'autopilot:lastJob';

const STRATEGY_LABEL: Record<Strategy, string> = {
  economy: 'Экономика',
  military: 'Военная',
  defense: 'Оборона',
  expansion: 'Экспансия',
  auto: 'Авто',
};

const CATEGORY_LABEL: Record<string, string> = {
  building: 'Строительство',
  research: 'Исследование',
  mission: 'Миссия',
  defense: 'Оборона',
  profession: 'Профессия',
  acs: 'Совместная атака',
};

const STATUS_LABEL: Record<string, string> = {
  pending: 'ожидает',
  ready: 'готово',
  executed: 'выполнено',
  expired: 'истекло',
};

function loadStrategy(): Strategy {
  if (typeof window === 'undefined') return 'auto';
  const v = window.localStorage.getItem(STORAGE_KEY);
  if (v && (STRATEGIES as string[]).includes(v)) return v as Strategy;
  return 'auto';
}

export function AdvisorScreen() {
  const [tab, setTab] = useState<'ai' | 'autopilot'>('autopilot');

  return (
    <table className="ntable">
      <thead>
        <tr>
          <th style={{ textAlign: 'center' }}>Советник</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td style={{ textAlign: 'center' }}>
            <button
              type="button"
              onClick={() => setTab('ai')}
              style={{ marginRight: 8, fontWeight: tab === 'ai' ? 'bold' : 'normal' }}
            >
              AI-ассистент
            </button>
            <button
              type="button"
              onClick={() => setTab('autopilot')}
              style={{ fontWeight: tab === 'autopilot' ? 'bold' : 'normal' }}
            >
              Автопилот
            </button>
          </td>
        </tr>
        <tr>
          <td>{tab === 'ai' ? <AITab /> : <AutopilotTab />}</td>
        </tr>
      </tbody>
    </table>
  );
}

function AITab() {
  const [model, setModel] = useState('claude-haiku-4-5-20251001');
  const [question, setQuestion] = useState('');
  const [answer, setAnswer] = useState('');
  const [error, setError] = useState('');

  const ask = useMutation({
    mutationFn: () => askAdvisor(model, question),
    onSuccess: (res) => {
      setAnswer(res.answer);
      setError('');
    },
    onError: (err) => setError(err instanceof Error ? err.message : String(err)),
  });

  return (
    <div style={{ padding: 8 }}>
      <p>
        Задайте вопрос об игре. Стоимость зависит от модели: Haiku — 5 кр,
        Sonnet — 20 кр, Opus — 80 кр (если не задан OLLAMA_URL).
      </p>
      <div style={{ marginBottom: 6 }}>
        Модель:{' '}
        <select value={model} onChange={(e) => setModel(e.target.value)}>
          <option value="claude-haiku-4-5-20251001">Haiku (5 кр)</option>
          <option value="claude-sonnet-4-6">Sonnet (20 кр)</option>
          <option value="claude-opus-4-7">Opus (80 кр)</option>
        </select>
      </div>
      <textarea
        value={question}
        onChange={(e) => setQuestion(e.target.value)}
        placeholder="Например: какое здание построить следующим?"
        rows={4}
        style={{ width: '90%' }}
      />
      <br />
      <button
        type="button"
        onClick={() => ask.mutate()}
        disabled={!question.trim() || ask.isPending}
      >
        {ask.isPending ? 'Думаю…' : 'Спросить'}
      </button>
      {answer && (
        <div
          style={{
            marginTop: 12,
            padding: 8,
            border: '1px solid #555',
            whiteSpace: 'pre-wrap',
          }}
        >
          {answer}
        </div>
      )}
      {error && (
        <div className="false" style={{ marginTop: 8 }}>
          {error}
        </div>
      )}
    </div>
  );
}

function AutopilotTab() {
  const qc = useQueryClient();
  const navigate = useNavigate();
  const [strategy, setStrategy] = useState<Strategy>(loadStrategy());
  const [activeJobID, setActiveJobID] = useState<string | null>(() => {
    if (typeof window === 'undefined') return null;
    return window.localStorage.getItem(STORAGE_JOB);
  });
  const [error, setError] = useState('');

  useEffect(() => {
    if (typeof window !== 'undefined') {
      window.localStorage.setItem(STORAGE_KEY, strategy);
    }
  }, [strategy]);

  const advise = useMutation({
    mutationFn: () => startAutopilot(strategy),
    onSuccess: (res) => {
      setActiveJobID(res.job_id);
      setError('');
      if (typeof window !== 'undefined') {
        window.localStorage.setItem(STORAGE_JOB, res.job_id);
      }
      void qc.invalidateQueries({ queryKey: ['me'] });
    },
    onError: (err) => setError(err instanceof Error ? err.message : String(err)),
  });

  const result = useQuery({
    queryKey: ['autopilot-result', activeJobID],
    queryFn: () => fetchAutopilotResult(activeJobID!),
    enabled: !!activeJobID,
    refetchInterval: (query) => {
      const data = query.state.data as AutopilotResult | undefined;
      if (!data) return 2000;
      return data.status === 'pending' ? 2000 : false;
    },
  });

  const execute = useMutation({
    mutationFn: (recID: string) => executeAutopilot(activeJobID!, recID),
    onSuccess: (_res, recID) => {
      const rec = result.data?.recs.find((r) => r.id === recID);
      void qc.invalidateQueries();
      void qc.invalidateQueries({ queryKey: ['autopilot-history'] });
      navigateForRec(navigate, rec);
      setActiveJobID(null);
      if (typeof window !== 'undefined') {
        window.localStorage.removeItem(STORAGE_JOB);
      }
    },
    onError: (err) => setError(err instanceof Error ? err.message : String(err)),
  });

  const history = useQuery({
    queryKey: ['autopilot-history'],
    queryFn: fetchAutopilotHistory,
    refetchInterval: 30000,
  });

  const data = result.data;

  return (
    <div style={{ padding: 8 }}>
      <p>
        Автопилот предложит топ-3 хода исходя из выбранной стратегии (2 кр
        за запрос). Результат появится через несколько секунд.
      </p>

      <div style={{ marginBottom: 8 }}>
        Стратегия:{' '}
        <select
          value={strategy}
          onChange={(e) => setStrategy(e.target.value as Strategy)}
        >
          {STRATEGIES.map((s) => (
            <option key={s} value={s}>
              {STRATEGY_LABEL[s]}
            </option>
          ))}
        </select>{' '}
        <button
          type="button"
          onClick={() => advise.mutate()}
          disabled={advise.isPending || data?.status === 'pending'}
        >
          {advise.isPending || data?.status === 'pending'
            ? 'Считаю…'
            : 'Запустить'}
        </button>
      </div>

      {error && (
        <div className="false" style={{ marginBottom: 8 }}>
          {error}
        </div>
      )}

      {data?.status === 'pending' && (
        <div style={{ fontStyle: 'italic' }}>Автопилот считает…</div>
      )}

      {data?.status === 'ready' && data.recs.length === 0 && (
        <div>
          Нет доступных рекомендаций. Попробуйте другую стратегию или дождитесь
          накопления ресурсов.
        </div>
      )}

      {data?.status === 'ready' && data.recs.length > 0 && (
        <table className="ntable" style={{ marginTop: 8 }}>
          <thead>
            <tr>
              <th>Рекомендация</th>
              <th style={{ width: 120 }}>Категория</th>
              <th style={{ width: 100 }}>Score</th>
              <th style={{ width: 120 }}>Действие</th>
            </tr>
          </thead>
          <tbody>
            {data.recs.map((rec) => (
              <tr key={rec.id}>
                <td>
                  <strong>{rec.description}</strong>
                  <br />
                  <span className="small">{rec.benefit}</span>
                </td>
                <td>{CATEGORY_LABEL[rec.category] ?? rec.category}</td>
                <td>{rec.score.toFixed(2)}</td>
                <td>
                  <button
                    type="button"
                    onClick={() => execute.mutate(rec.id)}
                    disabled={execute.isPending}
                  >
                    {execute.isPending ? '…' : 'Выполнить'}
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {(history.data?.items?.length ?? 0) > 0 && (
        <table className="ntable" style={{ marginTop: 12 }}>
          <thead>
            <tr>
              <th colSpan={4}>История запросов</th>
            </tr>
            <tr>
              <th>Время</th>
              <th>Стратегия</th>
              <th>Статус</th>
              <th>Списано</th>
            </tr>
          </thead>
          <tbody>
            {history.data!.items.slice(0, 10).map((it) => (
              <tr key={it.job_id}>
                <td>{new Date(it.created_at).toLocaleString('ru-RU')}</td>
                <td>{STRATEGY_LABEL[it.strategy as Strategy] ?? it.strategy}</td>
                <td>{STATUS_LABEL[it.status] ?? it.status}</td>
                <td>{it.credits} кр</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

function navigateForRec(
  navigate: (path: string) => void,
  rec: AutopilotRecommendation | undefined,
) {
  if (!rec) return;
  switch (rec.category) {
    case 'building':
      navigate('/constructions');
      break;
    case 'research':
      navigate('/research');
      break;
    case 'mission':
      navigate('/mission');
      break;
    case 'profession':
      navigate('/profession');
      break;
    default:
      // другие — остаёмся на советнике
      break;
  }
}
