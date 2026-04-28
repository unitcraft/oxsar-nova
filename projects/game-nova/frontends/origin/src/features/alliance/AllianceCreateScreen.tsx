// S-010 Alliance create (план 72 Ф.3 Spring 2 ч.1).
//
// Pixel-perfect зеркало legacy `templates/standard/foundally.tpl`.
// Конфиг MIN/MAX_CHARS_ALLY_TAG/NAME в legacy = 3..5 / 3..64
// (см. openapi `/api/alliances` POST).

import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createAlliance } from '@/api/alliance';
import type { ApiError } from '@/api/client';
import { useTranslation } from '@/i18n/i18n';

export function AllianceCreateScreen() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  const navigate = useNavigate();
  const [tag, setTag] = useState('');
  const [name, setName] = useState('');
  const [errMsg, setErrMsg] = useState<string | null>(null);

  const create = useMutation({
    mutationFn: () => createAlliance({ tag, name }),
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['alliances'] });
      navigate('/alliance/me');
    },
    onError: (e) => {
      const err = e as ApiError;
      setErrMsg(err.message);
    },
  });

  const valid = tag.length >= 3 && tag.length <= 5 && name.length >= 3 && name.length <= 64;

  return (
    <form
      method="post"
      onSubmit={(ev) => {
        ev.preventDefault();
        if (!create.isPending && valid) create.mutate();
      }}
    >
      <table className="ntable">
        <thead>
          <tr>
            <th colSpan={2}>{t('alliance', 'foundAlliance')}</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <label htmlFor="tag">{t('alliance', 'allianceTag')}</label>
            </td>
            <td>
              <input
                type="text"
                name="tag"
                id="tag"
                maxLength={5}
                value={tag}
                onChange={(e) => setTag(e.target.value.toUpperCase())}
              />
            </td>
          </tr>
          <tr>
            <td>
              <label htmlFor="name">{t('alliance', 'allianceName')}</label>
            </td>
            <td>
              <input
                type="text"
                name="name"
                id="name"
                maxLength={64}
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </td>
          </tr>
          {errMsg && (
            <tr>
              <td colSpan={2} className="center">
                <span className="false">{errMsg}</span>
              </td>
            </tr>
          )}
          <tr>
            <td colSpan={2} className="center">
              <input
                type="submit"
                name="found"
                value={create.isPending ? '…' : t('alliance', 'createBtn')}
                className="button"
                disabled={!valid || create.isPending}
              />
            </td>
          </tr>
        </tbody>
      </table>
    </form>
  );
}
