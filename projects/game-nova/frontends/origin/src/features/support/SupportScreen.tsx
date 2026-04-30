// S-045 Support — обращение в техподдержку (план 72 Ф.5 Spring 4 ч.2).
//
// Pixel-perfect зеркало legacy `templates/standard/support.tpl`:
//   <table class="ntable">
//     <thead><tr><th>Регламент работы технической поддержки</th></tr></thead>
//     <tbody>...нумерованные параграфы регламента...
//
// Legacy показывает только статичный регламент (без формы). В origin-
// фронте мы добавляем интерактивную форму обращения (legacy в реале
// направлял юзеров на email — это устарело).
//
// КРОСС-СЕРВИСНЫЙ ENDPOINT: portal-backend /api/reports (план 56).
// game-nova /api/reports НЕ существует, не нужно подключать. Логика в
// api/support.ts.

import { useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { submitSupport, type SupportReason } from '@/api/support';
import { useTranslation } from '@/i18n/i18n';
import {
  buildSupportComment,
  EMPTY_SUPPORT_FIELDS as EMPTY_FIELDS,
  type SupportFields,
} from './comment';

const REASONS: Array<{ value: SupportReason; labelRu: string }> = [
  { value: 'cheat', labelRu: 'Чит / эксплойт' },
  { value: 'spam', labelRu: 'Спам / реклама' },
  { value: 'profanity', labelRu: 'Мат / оскорбления' },
  { value: 'extremism', labelRu: 'Экстремизм / разжигание' },
  { value: 'drugs', labelRu: 'Наркотики' },
  { value: 'impersonation', labelRu: 'Выдача за другое лицо' },
  { value: 'other', labelRu: 'Другое (тех. поддержка / вопрос)' },
];

export function SupportScreen() {
  const { t } = useTranslation();
  const [reason, setReason] = useState<SupportReason>('cheat');
  const [fields, setFields] = useState<SupportFields>(EMPTY_FIELDS);
  const [sent, setSent] = useState(false);

  const mut = useMutation({
    mutationFn: () =>
      submitSupport({
        reason,
        comment: buildSupportComment(fields),
      }),
    onSuccess: () => {
      setSent(true);
      setFields(EMPTY_FIELDS);
    },
  });

  function update<K extends keyof SupportFields>(k: K, v: SupportFields[K]) {
    setFields((p) => ({ ...p, [k]: v }));
  }

  function submit(e: React.FormEvent) {
    e.preventDefault();
    if (!fields.description.trim()) return;
    mut.mutate();
  }

  if (sent) {
    return (
      <table className="ntable">
        <thead>
          <tr>
            <th>Техническая поддержка</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td className="center">
              <span className="true">
                Спасибо. Заявка принята и будет рассмотрена в течение 3 рабочих дней.
              </span>
            </td>
          </tr>
          <tr>
            <td className="center">
              <button
                type="button"
                className="button"
                onClick={() => setSent(false)}
              >
                Закрыть
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    );
  }

  return (
    <form onSubmit={submit} data-testid="support-form">
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>Техническая поддержка</th>
          </tr>
        </thead>
        <tfoot>
          <tr>
            <td colSpan={2} className="center">
              <input
                type="submit"
                className="button"
                value={mut.isPending ? '…' : 'Отправить'}
                disabled={mut.isPending || !fields.description.trim()}
              />
              {mut.isError && (
                <div>
                  <span className="false">
                    {(mut.error as Error)?.message ?? 'error'}
                  </span>
                </div>
              )}
            </td>
          </tr>
        </tfoot>
        <tbody>
          <tr>
            <td colSpan={2} className="false2">
              Перед отправкой убедитесь что в заявке указан логин, вселенная,
              страница где возникла ошибка, и пошаговое описание для воспроизведения.
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="reason">Тема</label>
            </td>
            <td>
              <select
                id="reason"
                value={reason}
                onChange={(e) => setReason(e.target.value as SupportReason)}
              >
                {REASONS.map((r) => (
                  <option key={r.value} value={r.value}>
                    {r.labelRu}
                  </option>
                ))}
              </select>
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="login">Логин</label>
            </td>
            <td>
              <input
                id="login"
                type="text"
                maxLength={50}
                value={fields.login}
                onChange={(e) => update('login', e.target.value)}
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="universe">Вселенная</label>
            </td>
            <td>
              <input
                id="universe"
                type="text"
                maxLength={50}
                value={fields.universe}
                onChange={(e) => update('universe', e.target.value)}
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="page">Страница</label>
            </td>
            <td>
              <input
                id="page"
                type="text"
                maxLength={120}
                value={fields.page}
                onChange={(e) => update('page', e.target.value)}
                placeholder="например, /shipyard или /galaxy/4/120"
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="browser">Браузер</label>
            </td>
            <td>
              <input
                id="browser"
                type="text"
                maxLength={120}
                value={fields.browser}
                onChange={(e) => update('browser', e.target.value)}
                placeholder="Firefox 123, Chrome 130, …"
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="description">Описание</label>
            </td>
            <td>
              <textarea
                id="description"
                rows={4}
                maxLength={2000}
                value={fields.description}
                onChange={(e) => update('description', e.target.value)}
                required
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="steps">Шаги воспроизведения</label>
            </td>
            <td>
              <textarea
                id="steps"
                rows={3}
                maxLength={2000}
                value={fields.steps}
                onChange={(e) => update('steps', e.target.value)}
              />
            </td>
          </tr>
        </tbody>
      </table>
    </form>
  );
}
