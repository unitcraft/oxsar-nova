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

setcookie('oxsar-jwt', $token, time() + 86400 * 30, '/');

header('Location: /game.php?go=Main');
exit;
