// Package ids выдаёт идентификаторы. Используем UUIDv7 — монотонные,
// удобны для индексов и партиционирования.
package ids

// DUPLICATE: этот файл скопирован между Go-модулями oxsar/game-nova,
// oxsar/identity, oxsar/portal и oxsar/billing. При любом изменении
// синхронизируйте КОПИИ:
//   - projects/game-nova/backend/pkg/ids/ids.go
//   - projects/identity/backend/pkg/ids/ids.go
//   - projects/portal/backend/pkg/ids/ids.go
//   - projects/billing/backend/pkg/ids/ids.go
// Причина дубля: каждый домен — отдельный go.mod, без shared-модуля.

import "github.com/google/uuid"

// New возвращает новый UUIDv7 как строку. Паникует только на сбое
// генератора ОС, что означает, что приложение и так не сможет работать.
func New() string {
	v, err := uuid.NewV7()
	if err != nil {
		panic("uuid v7: " + err.Error())
	}
	return v.String()
}
