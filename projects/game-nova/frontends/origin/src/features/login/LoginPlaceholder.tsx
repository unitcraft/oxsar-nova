// Login-placeholder origin-фронта (план 72 Ф.2 Spring 1).
//
// Реальный экран входа реализуется в Ф.3 Spring 2. На Spring 1 — только
// stub, чтобы 401-перехватчик в api/client.ts не приземлял пользователя
// на неизвестный URL.

export function LoginPlaceholder() {
  return (
    <table className="ntable">
      <tbody>
        <tr>
          <th>Вход</th>
        </tr>
        <tr>
          <td>
            Экран входа будет добавлен в Ф.3 Spring 2 плана 72. Сейчас
            origin-фронт ожидает токен из identity-service (план 36) —
            переходите через portal или nova-логин, токены
            переиспользуются между вселенными.
          </td>
        </tr>
      </tbody>
    </table>
  );
}
