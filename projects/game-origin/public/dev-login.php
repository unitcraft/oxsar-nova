<?php
/**
 * DEV-only: ставит JWT-cookie и редиректит на главную.
 * НЕ для продакшена. В prod аутентификация идёт через auth-service portal/login.
 *
 * `dev-user-001` привязан к легаси-юзеру `test` (userid=1) через
 * `na_user.global_user_id` — содержит 9 планет, 4 луны, ~36M очков.
 *
 * Чтобы войти под другим легаси-юзером — выставить `?as_uid=N` (нужен
 * соответствующий global_user_id в БД).
 */

// JWT с alg=none (валидно когда AUTH_JWKS_URL пустой — dev-режим)
$payload = [
    'sub'              => 'dev-user-001',  // → na_user.global_user_id
    'username'         => 'dev',
    'email'            => 'dev@oxsar-nova.local',
    'active_universes' => ['dm'],
    'roles'            => ['admin'],
    'exp'              => time() + 86400 * 30,
];

function b64url(string $s): string {
    return rtrim(strtr(base64_encode($s), '+/', '-_'), '=');
}

$token = b64url('{"typ":"JWT","alg":"none"}')
       . '.' . b64url(json_encode($payload, JSON_UNESCAPED_UNICODE))
       . '.';

// План 37.7.2: CSRF защита через SameSite=Strict cookie + httponly.
// Браузер не отправит cookie на cross-site POST → CSRF блокируется.
// secure только если HTTPS (для local dev http не ставим — иначе cookie
// не работает).
setcookie('oxsar-jwt', $token, [
    'expires'  => time() + 86400 * 30,
    'path'     => '/',
    'samesite' => 'Strict',
    'httponly' => true,
    'secure'   => !empty($_SERVER['HTTPS']),
]);

header('Location: /game.php?go=Main');
exit;
