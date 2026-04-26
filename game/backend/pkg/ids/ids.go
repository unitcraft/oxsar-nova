// Package ids выдаёт идентификаторы. Используем UUIDv7 — монотонные,
// удобны для индексов и партиционирования.
package ids

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
