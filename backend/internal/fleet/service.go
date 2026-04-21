package fleet

import (
	"context"
	"errors"
	"fmt"
)

// Service — публичный фасад пакета. TODO (M3): полноценная реализация
// всех миссий. Сейчас — каркас с Dispatch-валидацией и регистрацией
// миссий.
type Service struct {
	missions map[int]Mission
}

func NewService() *Service { return &Service{missions: map[int]Mission{}} }

// Register добавляет реализацию миссии. Ошибка — если уже зарегистрирована.
func (s *Service) Register(m Mission) error {
	if _, ok := s.missions[m.Kind()]; ok {
		return fmt.Errorf("fleet: mission %d already registered", m.Kind())
	}
	s.missions[m.Kind()] = m
	return nil
}

// Send валидирует и запускает миссию. TODO: списать ресурсы, записать
// флот в БД, создать event прибытия. Требуется реализация galaxy-
// пакета (расчёт дистанции / времени).
func (s *Service) Send(ctx context.Context, in Dispatch) (Fleet, error) {
	m, ok := s.missions[in.Mission]
	if !ok {
		return Fleet{}, errors.New("fleet: unknown mission")
	}
	if err := m.Validate(ctx, in); err != nil {
		return Fleet{}, err
	}
	return Fleet{}, errors.New("fleet.Send: not implemented yet (M3)")
}
