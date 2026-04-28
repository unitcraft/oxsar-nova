// S-043 UserAgreement — пользовательское соглашение (план 72 Ф.5
// Spring 4 ч.2).
//
// Юр-документ. Источник истины — портал (portal-frontend), который
// уже хостит /user-agreement и /privacy с актуальными текстами
// (план 50, маркировка 12+ и 149-ФЗ). Origin-фронт делает
// CROSS-LINK, не дубликат текста — это исключает риск рассинхрона
// при правках юр-документов и централизует юр-content.
//
// См. simplifications.md P72.S4.USER_AGREEMENT.
//
// Если VITE_PORTAL_BASE_URL не задана — fallback на относительный
// /user-agreement (404 в dev, корректно если portal обслуживает
// тот же домен в проде).

import { useTranslation } from '@/i18n/i18n';

const PORTAL_BASE =
  (import.meta.env['VITE_PORTAL_BASE_URL'] as string | undefined) ?? '';
const AGREEMENT_URL = `${PORTAL_BASE}/user-agreement`;

export function UserAgreementScreen() {
  const { t } = useTranslation();
  return (
    <table className="ntable">
      <thead>
        <tr>
          <th>{t('prefs', 'userAgreement') || 'Пользовательское соглашение'}</th>
        </tr>
      </thead>
      <tbody>
        <tr>
          <td style={{ textAlign: 'justify' }}>
            {t('prefs', 'userAgreementBody') ||
              'Полный текст пользовательского соглашения размещён на портале oxsar-nova.ru. Откройте ссылку ниже в новой вкладке.'}
          </td>
        </tr>
        <tr>
          <td className="center">
            <a
              href={AGREEMENT_URL}
              target="_blank"
              rel="noopener noreferrer"
              className="button"
              data-testid="user-agreement-link"
            >
              {t('prefs', 'userAgreementOpen') || 'Открыть соглашение'}
            </a>
          </td>
        </tr>
      </tbody>
    </table>
  );
}
