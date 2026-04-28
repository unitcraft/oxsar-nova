// План 67 Ф.5 ч.2 — передача лидерства альянса (U-004, D-040).
//
// 2-step flow:
//   1) POST /api/alliances/{id}/transfer-leadership/code
//      Body: { new_owner_id }
//      → 202 { expires_at, ttl_seconds }
//      Код 8-символьный отправляется текущему owner'у системным
//      сообщением (folder=13). НЕ возвращается в HTTP-ответе.
//   2) POST /api/alliances/{id}/transfer-leadership
//      Body: { new_owner_id, code }
//      → 204
//      В одной транзакции: смена owner_id, переразметка rank,
//      audit-запись, удаление кода.
//
// Idempotency-Key обязателен на обоих шагах (R9): запрос кода — чтобы
// одна и та же отправка не дублировалась; confirm — чтобы повторный
// confirm с тем же ключом был no-op.
//
// Возможные коды ошибок (см. internal/alliance/transfer.go):
//   - 400 invalid_code / expired_code / not_a_member /
//          billing_unavailable (если identity недоступна)
//   - 403 not_owner
//   - 404 alliance_not_found / member_not_found

import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api, genIdempotencyKey } from '@/api/client';
import { Modal } from '@/ui/Modal';
import { useToast } from '@/ui/Toast';
import { useTranslation } from '@/i18n/i18n';

interface Member {
  user_id: string;
  username: string;
  rank: string;
}

type Step = 'pick' | 'code';

export function TransferLeadershipDialog({
  allianceID,
  members,
  currentOwnerID,
  onClose,
}: {
  allianceID: string;
  members: Member[];
  currentOwnerID: string;
  onClose: () => void;
}) {
  const { t } = useTranslation('alliance');
  const qc = useQueryClient();
  const toast = useToast();

  const [step, setStep] = useState<Step>('pick');
  const [newOwnerID, setNewOwnerID] = useState<string>('');
  const [code, setCode] = useState<string>('');
  // Idempotency-Key фиксируется на этапе запроса кода и переиспользуется
  // при confirm (commit-семантика: один и тот же transfer-intent).
  const [codeKey, setCodeKey] = useState<string>('');
  const [confirmKey, setConfirmKey] = useState<string>('');
  const [ttlSec, setTtlSec] = useState<number>(0);

  // Кандидаты — все члены кроме owner'а; owner кикается из выбора, чтобы
  // нельзя было «передать самому себе» (бэкенд тоже отбьёт, но
  // дисциплина UI — не показывать невалидный вариант).
  const candidates = members.filter((m) => m.user_id !== currentOwnerID);

  const requestCode = useMutation({
    mutationFn: () => {
      const key = genIdempotencyKey();
      setCodeKey(key);
      return api.post<{ expires_at: string; ttl_seconds: number }>(
        `/api/alliances/${allianceID}/transfer-leadership/code`,
        { new_owner_id: newOwnerID },
        { idempotencyKey: key },
      );
    },
    onSuccess: (data) => {
      setTtlSec(data.ttl_seconds);
      setStep('code');
      toast.show('info', t('transferLeadership.title'), t('transferLeadership.codeSent'));
    },
    onError: (e) => {
      const err = e as Error & { code?: string };
      toast.show('danger', t('transferLeadership.title'), translateError(err.code, err.message, t));
    },
  });

  const confirm = useMutation({
    mutationFn: () => {
      // Свежий confirm-key только для первого вызова; при retry
      // (mutation rerun) переиспользуем тот же — это и есть смысл R9.
      const key = confirmKey || genIdempotencyKey();
      if (!confirmKey) setConfirmKey(key);
      return api.post<void>(
        `/api/alliances/${allianceID}/transfer-leadership`,
        { new_owner_id: newOwnerID, code: code.trim().toUpperCase() },
        { idempotencyKey: key },
      );
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      toast.show('success', t('transferLeadership.title'), t('transferLeadership.success'));
      onClose();
    },
    onError: (e) => {
      const err = e as Error & { code?: string };
      toast.show('danger', t('transferLeadership.title'), translateError(err.code, err.message, t));
      // Сброс confirm-key, чтобы следующий attempt был свежим (например,
      // юзер ввёл новый код после expired_code).
      setConfirmKey('');
    },
  });

  const onBack = () => {
    setStep('pick');
    setCode('');
    setConfirmKey('');
  };

  return (
    <Modal title={t('transferLeadership.title')} onClose={onClose} maxWidth={520}>
      {step === 'pick' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <p style={{ margin: 0, fontSize: 14, color: 'var(--ox-fg-dim)' }}>
            {t('transferLeadership.pickIntro')}
          </p>

          {candidates.length === 0 ? (
            <div style={{ fontSize: 14, color: 'var(--ox-fg-dim)', fontStyle: 'italic' }}>
              {t('transferLeadership.noCandidates')}
            </div>
          ) : (
            <div
              role="radiogroup"
              aria-label={t('transferLeadership.pickIntro')}
              style={{
                display: 'flex',
                flexDirection: 'column',
                gap: 4,
                maxHeight: 280,
                overflowY: 'auto',
                border: '1px solid var(--ox-border)',
                borderRadius: 4,
                padding: 8,
              }}
            >
              {candidates.map((m) => (
                <label
                  key={m.user_id}
                  style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer', padding: '4px 6px' }}
                >
                  <input
                    type="radio"
                    name="new-owner"
                    value={m.user_id}
                    checked={newOwnerID === m.user_id}
                    onChange={() => setNewOwnerID(m.user_id)}
                  />
                  <span>{m.username}</span>
                  <span style={{ marginLeft: 'auto', fontSize: 13, color: 'var(--ox-fg-muted)' }}>
                    {m.rank}
                  </span>
                </label>
              ))}
            </div>
          )}

          <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
            <button type="button" className="btn-ghost btn-sm" onClick={onClose}>
              {t('cancelBtn')}
            </button>
            <button
              type="button"
              className="btn btn-sm"
              disabled={!newOwnerID || requestCode.isPending}
              onClick={() => requestCode.mutate()}
            >
              {requestCode.isPending ? '…' : t('transferLeadership.requestCodeBtn')}
            </button>
          </div>
        </div>
      )}

      {step === 'code' && (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
          <p style={{ margin: 0, fontSize: 14, color: 'var(--ox-fg-dim)' }}>
            {t('transferLeadership.codeIntro', { ttl: String(Math.floor(ttlSec / 60)) })}
          </p>

          <label style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <span style={{ fontSize: 13, color: 'var(--ox-fg-dim)' }}>
              {t('transferLeadership.codeLabel')}
            </span>
            <input
              value={code}
              onChange={(e) => setCode(e.target.value.toUpperCase())}
              maxLength={8}
              minLength={8}
              autoComplete="off"
              autoFocus
              style={{
                width: 200,
                fontFamily: 'var(--ox-mono)',
                fontSize: 18,
                letterSpacing: '0.2em',
              }}
            />
            <span style={{ fontSize: 12, color: 'var(--ox-fg-muted)' }}>
              {t('transferLeadership.codeHint')}
            </span>
          </label>

          <div style={{ display: 'flex', gap: 8, justifyContent: 'space-between' }}>
            <button type="button" className="btn-ghost btn-sm" onClick={onBack}>
              ← {t('transferLeadership.backBtn')}
            </button>
            <div style={{ display: 'flex', gap: 8 }}>
              <button type="button" className="btn-ghost btn-sm" onClick={onClose}>
                {t('cancelBtn')}
              </button>
              <button
                type="button"
                className="btn btn-sm"
                disabled={code.length !== 8 || confirm.isPending}
                onClick={() => confirm.mutate()}
              >
                {confirm.isPending ? '…' : t('confirmBtn')}
              </button>
            </div>
          </div>
        </div>
      )}
      {/* Сохраняем codeKey в state для UX-debug, чтобы повторный
          requestCode не отправлялся по той же мутации; используется в
          dev-режиме для логов. */}
      <span style={{ display: 'none' }} data-codekey={codeKey} />
    </Modal>
  );
}

// translateError — известные коды → i18n-ключи, для всего остального
// — оригинальный message (он уже locale-aware на бэке).
function translateError(
  code: string | undefined,
  message: string,
  t: (k: string) => string,
): string {
  if (!code) return message;
  switch (code) {
    case 'invalid_code': return t('transferLeadership.errInvalidCode');
    case 'expired_code': return t('transferLeadership.errExpiredCode');
    case 'not_a_member': return t('transferLeadership.errNotMember');
    case 'not_owner':    return t('transferLeadership.errNotOwner');
    case 'billing_unavailable':
    case 'identity_unavailable':
      return t('transferLeadership.errBillingUnavailable');
    case 'rate_limited': return t('transferLeadership.errRateLimited');
    case 'attempts_exhausted': return t('transferLeadership.errAttemptsExhausted');
    default: return message;
  }
}
