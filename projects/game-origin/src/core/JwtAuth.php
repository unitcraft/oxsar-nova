<?php
/**
 * JWT authentication middleware for game-origin.
 *
 * Verifies RS256 JWT from auth-service (plan-36) and populates $_SESSION
 * so the rest of the game code can read $_SESSION["userid"] etc. as before.
 *
 * Auth-service exposes public keys at:
 *   GET {AUTH_JWKS_URL}  (e.g. https://auth.oxsar-nova.ru/.well-known/jwks.json)
 *
 * JWT claims used:
 *   sub              — global user UUID (stored as global_user_id in local DB)
 *   username
 *   active_universes — game-origin checks its UNIVERSE_ID is in this list
 *   roles
 *   exp
 */

if (!defined("RECIPE_ROOT_DIR")) { die("Hacking attempt detected."); }

class JwtAuth
{
    private static ?array $jwks = null;

    /**
     * Verify JWT from Authorization header or cookie, populate $_SESSION.
     * Returns true if authenticated, false if guest.
     */
    /**
     * Этап 1: Проверка JWT (без БД). Кладёт payload в $_SESSION['jwt_payload'].
     * Вызывается из global.inc.php до Core init.
     */
    public static function authenticate(): bool
    {
        // Если payload уже разобран в текущей сессии — пропускаем повторную проверку
        if (!empty($_SESSION['jwt_payload']) && !empty($_SESSION['userid'])) {
            return true;
        }

        $token = self::extractToken();
        if (!$token) {
            return false;
        }

        $payload = self::verifyToken($token);
        if (!$payload) {
            return false;
        }

        $globalUserId = $payload['sub'] ?? null;
        if (!$globalUserId) {
            return false;
        }

        $_SESSION['jwt_payload']    = $payload;
        $_SESSION['global_user_id'] = $globalUserId;
        return true;
    }

    /**
     * Этап 2: Lazy join с локальной БД. Вызывается после Core::setDatabase().
     */
    public static function resolveUser(): bool
    {
        if (!empty($_SESSION['userid'])) {
            return true;
        }
        $payload      = $_SESSION['jwt_payload']    ?? null;
        $globalUserId = $_SESSION['global_user_id'] ?? null;
        if (!$payload || !$globalUserId) {
            return false;
        }

        $localUser = self::lazyJoin($globalUserId, $payload);
        if (!$localUser) {
            return false;
        }

        $_SESSION['userid']     = $localUser['userid'];
        $_SESSION['username']   = $payload['username']           ?? $localUser['username'];
        $_SESSION['email']      = $localUser['email']            ?? '';
        $_SESSION['activation'] = $localUser['activation']       ?? 1;
        $_SESSION['is_admin']   = in_array('admin', $payload['roles'] ?? []);
        $_SESSION['skin_type']  = $localUser['templatepackage']  ?? 'standard';
        $_SESSION['curplanet']  = $localUser['curplanet']        ?? 0;
        $_SESSION['sid']        = '';
        return true;
    }

    private static function extractToken(): ?string
    {
        // 1. Authorization: Bearer <token>
        $header = $_SERVER['HTTP_AUTHORIZATION'] ?? '';
        if (preg_match('/^Bearer\s+(\S+)$/i', $header, $m)) {
            return $m[1];
        }
        // 2. Cookie oxsar-jwt (для браузерного флоу)
        return $_COOKIE['oxsar-jwt'] ?? null;
    }

    private static function verifyToken(string $token): ?array
    {
        $parts = explode('.', $token);
        if (count($parts) !== 3) {
            return null;
        }

        [$headerB64, $payloadB64, $sigB64] = $parts;

        $header  = json_decode(self::base64UrlDecode($headerB64), true);
        $payload = json_decode(self::base64UrlDecode($payloadB64), true);

        if (!$header || !$payload) {
            return null;
        }

        // Проверка срока действия
        if (!empty($payload['exp']) && $payload['exp'] < time()) {
            return null;
        }

        // Если AUTH_JWKS_URL не задан — пропускаем проверку подписи (dev-режим)
        $jwksUrl = defined('AUTH_JWKS_URL') ? AUTH_JWKS_URL : (getenv('AUTH_JWKS_URL') ?: '');
        if (!$jwksUrl) {
            return $payload;
        }

        // RS256 signature verification
        $kid = $header['kid'] ?? null;
        $key = self::getPublicKey($jwksUrl, $kid);
        if (!$key) {
            return null;
        }

        $data = $headerB64 . '.' . $payloadB64;
        $sig  = self::base64UrlDecode($sigB64);

        $ok = openssl_verify($data, $sig, $key, OPENSSL_ALGO_SHA256);
        return $ok === 1 ? $payload : null;
    }

    private static function getPublicKey(string $jwksUrl, ?string $kid)
    {
        if (self::$jwks === null) {
            $raw = @file_get_contents($jwksUrl);
            self::$jwks = $raw ? (json_decode($raw, true)['keys'] ?? []) : [];
        }

        foreach (self::$jwks as $jwk) {
            if ($kid && ($jwk['kid'] ?? null) !== $kid) {
                continue;
            }
            if (($jwk['kty'] ?? '') !== 'RSA') {
                continue;
            }
            return self::jwkToPublicKey($jwk);
        }
        return null;
    }

    private static function jwkToPublicKey(array $jwk)
    {
        $n = self::base64UrlDecode($jwk['n']);
        $e = self::base64UrlDecode($jwk['e']);
        // Сборка DER-структуры RSA public key
        $publicKey = openssl_pkey_get_public([
            'n' => $n,
            'e' => $e,
        ]);
        if ($publicKey) {
            return $publicKey;
        }
        // Fallback через pem если openssl_pkey_get_public не поддерживает массив
        return null;
    }

    /**
     * Lazy join: найти пользователя по global_user_id или создать новую запись.
     */
    private static function lazyJoin(string $globalUserId, array $payload): ?array
    {
        $db  = Core::getDB();
        $gid = $db->quote_db_value($globalUserId);

        $row = $db->queryRow(
            "SELECT * FROM `" . PREFIX . "user` WHERE global_user_id = $gid LIMIT 1"
        );

        if ($row) {
            return $row;
        }

        // Создаём нового пользователя (минимальный набор полей).
        // Пароль не нужен — аутентификация через JWT (auth-service plan-36).
        $username = $db->quote_db_value($payload['username'] ?? 'player_' . substr($globalUserId, 0, 8));
        $email    = $db->quote_db_value($payload['email']    ?? '');
        $now      = time();

        // Все NOT NULL без default из legacy-схемы na_user (16 полей).
        // activation — varbinary(32) (хеш email-токена), пустая строка = «активирован».
        // dm_points/regtime/last/etc — числовые, передаём 0/$now.
        $db->query(
            "INSERT INTO `" . PREFIX . "user`"
            . " (global_user_id, username, email, temp_email, languageid, timezone,"
            . "  templatepackage, theme, dm_points, activation, regtime, last,"
            . "  asteroid, umode, umodemin, planetorder, `delete`)"
            . " VALUES ($gid, $username, $email, '', 0, '',"
            . "         'standard', '', 0, '', $now, $now,"
            . "         0, 0, 0, 0, 0)"
        );

        $newId = $db->insert_id();
        return $db->queryRow("SELECT * FROM `" . PREFIX . "user` WHERE userid = $newId LIMIT 1");
    }

    private static function base64UrlDecode(string $input): string
    {
        $pad = strlen($input) % 4;
        if ($pad) {
            $input .= str_repeat('=', 4 - $pad);
        }
        return base64_decode(strtr($input, '-_', '+/'));
    }
}
