package aiadvisor

import (
	"testing"

	"oxsar/game-nova/internal/config"
)

func TestKnownModels(t *testing.T) {
	t.Parallel()
	cases := []struct {
		model string
		cost  int
	}{
		{"claude-haiku-4-5-20251001", 5},
		{"claude-sonnet-4-6", 20},
		{"claude-opus-4-7", 80},
	}
	for _, tc := range cases {
		cost, ok := KnownModels[tc.model]
		if !ok {
			t.Errorf("model %s not in KnownModels", tc.model)
			continue
		}
		if cost != tc.cost {
			t.Errorf("model %s: expected cost %d, got %d", tc.model, tc.cost, cost)
		}
	}
}

func TestErrSentinels(t *testing.T) {
	t.Parallel()
	for _, err := range []error{ErrNotEnoughCredit, ErrRateLimitReached, ErrUnknownModel, ErrNoBackend} {
		if err == nil {
			t.Errorf("sentinel error must not be nil")
		}
	}
}

func TestBuildStaticGameKnowledge(t *testing.T) {
	t.Parallel()
	s := buildStaticGameKnowledge()
	if len(s) < 50 {
		t.Fatalf("static game knowledge too short: %d chars", len(s))
	}
}

func TestNewService_NoBackend(t *testing.T) {
	t.Parallel()
	// Ни APIKey ни OllamaURL не заданы → llm == nil → ErrNoBackend при Ask.
	svc := NewService(nil, config.AIAdvisorConfig{MaxPerDay: 20, MaxTokens: 1024})
	if svc.llm != nil {
		t.Fatal("expected nil llm when no backend configured")
	}
}
