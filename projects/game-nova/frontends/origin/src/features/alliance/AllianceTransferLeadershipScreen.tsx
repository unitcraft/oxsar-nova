// S-017 Alliance transfer-leadership (план 72 Ф.3 Spring 2 ч.1).
//
// 2-step flow plan 67 Ф.3 (U-004 / D-040):
//   1) POST /api/alliances/{id}/transfer-leadership/code
//      Body: { new_owner_id }
//      → 202 { expires_at, ttl_seconds }
//      Код 8-символьный отправляется текущему owner'у системным
//      сообщением (folder=13). НЕ возвращается в HTTP-ответе.
//   2) POST /api/alliances/{id}/transfer-leadership
//      Body: { new_owner_id, code }
//      → 204
//
// Idempotency-Key обязателен на обоих шагах (R9): один и тот же
// transfer-intent должен иметь один и тот же ключ при retry.

import { useState } from 'react';
import { Link, Navigate, useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { confirmTransfer, requestTransferCode } from '@/api/alliance';
import { newIdempotencyKey } from '@/api/idempotency';
import type { ApiError } from '@/api/client';
import { useAuthStore } from '@/stores/auth';
import { useTranslation } from '@/i18n/i18n';
import { useMyAlliance } from './common';

type Step = 'pick' | 'code';

export function AllianceTransferLeadershipScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const my = useMyAlliance();
  const userId = useAuthStore((s) => s.userId);

  const [step, setStep] = useState<Step>('pick');
  const [newOwnerID, setNewOwnerID] = useState<string>('');
  const [code, setCode] = useState<string>('');
  const [codeKey, setCodeKey] = useState<string>('');
  const [confirmKey, setConfirmKey] = useState<string>('');
  const [ttlSec, setTtlSec] = useState<number>(0);
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const allianceID = my.data?.alliance.id ?? '';

  const requestCode = useMutation({
    mutationFn: () => {
      const key = codeKey || newIdempotencyKey();
      if (!codeKey) setCodeKey(key);
      return requestTransferCode(allianceID, newOwnerID, key);
    },
    onSuccess: (data) => {
      setTtlSec(data.ttl_seconds);
      setStep('code');
      setErrMsg(null);
    },
    onError: (e) => setErrMsg(translateErr(e as ApiError, t)),
  });

  const confirm = useMutation({
    mutationFn: () => {
      const key = confirmKey || newIdempotencyKey();
      if (!confirmKey) setConfirmKey(key);
      return confirmTransfer(allianceID, newOwnerID, code.trim().toUpperCase(), key);
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      navigate('/alliance/me');
    },
    onError: (e) => {
      setErrMsg(translateErr(e as ApiError, t));
      setConfirmKey('');
    },
  });

  if (my.isLoading) return <div className="idiv">…</div>;
  if (!my.data) return <Navigate to="/alliance" replace />;

  const al = my.data.alliance;
  const isOwner = !!userId && userId === al.owner_id;
  if (!isOwner) {
    return (
      <div className="idiv">
        <span className="false">{t('alliance', 'rightManagement')}</span>
      </div>
    );
  }

  const candidates = my.data.members.filter((m) => m.user_id !== al.owner_id);

  return (
    <>
      <div className="idiv">
        <Link to="/alliance/me">← {al.tag}</Link>
      </div>

      <table className="ntable">
        <thead>
          <tr>
            <th>{t('alliance', 'referFounderStatus')}</th>
          </tr>
        </thead>
        <tbody>
          {step === 'pick' && (
            <>
              <tr>
                <td>
                  <select
                    value={newOwnerID}
                    onChange={(e) => setNewOwnerID(e.target.value)}
                  >
                    <option value="">—</option>
                    {candidates.map((m) => (
                      <option key={m.user_id} value={m.user_id}>
                        {m.username} ({m.rank_name || m.rank})
                      </option>
                    ))}
                  </select>
                </td>
              </tr>
              <tr>
                <td className="center">
                  <input
                    type="button"
                    className="button"
                    disabled={!newOwnerID || requestCode.isPending}
                    value={
                      requestCode.isPending
                        ? '…'
                        : t('alliance', 'transferLeadership.requestCodeBtn') ||
                          t('alliance', 'commit') ||
                          'OK'
                    }
                    onClick={() => requestCode.mutate()}
                  />
                </td>
              </tr>
            </>
          )}
          {step === 'code' && (
            <>
              <tr>
                <td>
                  {t('alliance', 'transferLeadership.codeIntro', {
                    ttl: String(Math.floor(ttlSec / 60)),
                  })}
                </td>
              </tr>
              <tr>
                <td>
                  <input
                    type="text"
                    value={code}
                    maxLength={8}
                    autoFocus
                    onChange={(e) => setCode(e.target.value.toUpperCase())}
                    style={{
                      letterSpacing: '0.2em',
                      fontFamily: 'monospace',
                      fontSize: '1.2em',
                      width: 200,
                    }}
                  />
                </td>
              </tr>
              <tr>
                <td className="center">
                  <input
                    type="button"
                    className="button"
                    value="←"
                    onClick={() => {
                      setStep('pick');
                      setCode('');
                      setConfirmKey('');
                      setErrMsg(null);
                    }}
                  />{' '}
                  <input
                    type="button"
                    className="button"
                    disabled={code.length !== 8 || confirm.isPending}
                    value={
                      confirm.isPending ? '…' : t('alliance', 'confirmBtn')
                    }
                    onClick={() => confirm.mutate()}
                  />
                </td>
              </tr>
            </>
          )}
          {errMsg && (
            <tr>
              <td className="center">
                <span className="false">{errMsg}</span>
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </>
  );
}

function translateErr(
  err: ApiError,
  t: (g: string, k: string) => string,
): string {
  switch (err.code) {
    case 'invalid_code':
      return t('alliance', 'transferLeadership.errInvalidCode');
    case 'expired_code':
      return t('alliance', 'transferLeadership.errExpiredCode');
    case 'not_a_member':
      return t('alliance', 'transferLeadership.errNotMember');
    case 'not_owner':
      return t('alliance', 'transferLeadership.errNotOwner');
    case 'rate_limited':
      return t('alliance', 'transferLeadership.errRateLimited');
    case 'attempts_exhausted':
      return t('alliance', 'transferLeadership.errAttemptsExhausted');
    default:
      return err.message;
  }
}
