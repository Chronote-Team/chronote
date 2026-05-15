package app

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	postcardaiapp "chronote-refactor/internal/modules/postcardai/app"
	"chronote-refactor/internal/modules/postcardai/domain"
	platformconfig "chronote-refactor/internal/platform/config"
)

type fakeAnalysisProcessor struct {
	outcomes []*postcardaiapp.WorkerOutcome
	errs     []error
	calls    int
}

func (p *fakeAnalysisProcessor) RunNextAnalysisJob(ctx context.Context, workerID string) (*postcardaiapp.WorkerOutcome, error) {
	index := p.calls
	p.calls++
	if index < len(p.errs) && p.errs[index] != nil {
		return nil, p.errs[index]
	}
	if index < len(p.outcomes) {
		return p.outcomes[index], nil
	}
	return &postcardaiapp.WorkerOutcome{}, nil
}

func TestRunAnalysisWorkerRunOnceProcessesSingleJob(t *testing.T) {
	processor := &fakeAnalysisProcessor{
		outcomes: []*postcardaiapp.WorkerOutcome{{JobID: 10, Status: domain.StatusSucceeded}},
	}
	var logs bytes.Buffer
	err := RunAnalysisWorker(context.Background(), processor, WorkerOptions{
		WorkerID: "test-worker",
		RunOnce:  true,
		Logger:   log.New(&logs, "", 0),
		sleep: func(context.Context, time.Duration) error {
			t.Fatal("sleep should not be called after work")
			return nil
		},
	})
	if err != nil {
		t.Fatalf("RunAnalysisWorker returned error: %v", err)
	}
	if processor.calls != 1 {
		t.Fatalf("calls = %d, want 1", processor.calls)
	}
	rendered := logs.String()
	if !strings.Contains(rendered, "job_id 10") || !strings.Contains(rendered, "status succeeded") {
		t.Fatalf("worker log missing safe outcome fields: %q", rendered)
	}
}

func TestRunAnalysisWorkerSleepsWhenNoJob(t *testing.T) {
	processor := &fakeAnalysisProcessor{outcomes: []*postcardaiapp.WorkerOutcome{{}}}
	ctx, cancel := context.WithCancel(context.Background())
	sleepCalls := 0
	err := RunAnalysisWorker(ctx, processor, WorkerOptions{
		RunOnce: false,
		Logger:  log.New(&bytes.Buffer{}, "", 0),
		sleep: func(context.Context, time.Duration) error {
			sleepCalls++
			cancel()
			return context.Canceled
		},
	})
	if err != nil {
		t.Fatalf("RunAnalysisWorker returned error: %v", err)
	}
	if sleepCalls != 1 {
		t.Fatalf("sleepCalls = %d, want 1", sleepCalls)
	}
}

func TestRunAnalysisWorkerReturnsRunOnceError(t *testing.T) {
	processor := &fakeAnalysisProcessor{errs: []error{errors.New("database unavailable")}}
	err := RunAnalysisWorker(context.Background(), processor, WorkerOptions{
		RunOnce: true,
		Logger:  log.New(&bytes.Buffer{}, "", 0),
	})
	if err == nil {
		t.Fatal("expected run-once error")
	}
}

func TestWorkerOptionsFromConfigUsesDefaultsAndEnvValues(t *testing.T) {
	cfg := &platformconfig.Config{}
	opts := WorkerOptionsFromConfig(cfg)
	if opts.WorkerID != defaultWorkerID {
		t.Fatalf("WorkerID = %q, want %q", opts.WorkerID, defaultWorkerID)
	}
	if opts.IdleSleep != defaultIdleSleep {
		t.Fatalf("IdleSleep = %s, want %s", opts.IdleSleep, defaultIdleSleep)
	}
	if opts.ErrorSleep != defaultErrorSleep {
		t.Fatalf("ErrorSleep = %s, want %s", opts.ErrorSleep, defaultErrorSleep)
	}

	cfg.AIWorker.ID = "worker-2"
	cfg.AIWorker.IdleSleep = "25ms"
	cfg.AIWorker.ErrorSleep = "50ms"
	cfg.AIWorker.RunOnce = true
	opts = WorkerOptionsFromConfig(cfg)
	if opts.WorkerID != "worker-2" || opts.IdleSleep != 25*time.Millisecond || opts.ErrorSleep != 50*time.Millisecond || !opts.RunOnce {
		t.Fatalf("unexpected parsed worker options: %#v", opts)
	}
}

func TestNormalizeAIEndpointTypeSupportsChatCompletionsAliases(t *testing.T) {
	for _, input := range []string{"chat", "chat-completions", "chat_completions", " Chat_Completions "} {
		if got := normalizeAIEndpointType(input); got != "chat_completions" {
			t.Fatalf("normalizeAIEndpointType(%q) = %q, want chat_completions", input, got)
		}
	}
	if got := normalizeAIEndpointType("responses"); got != "responses" {
		t.Fatalf("normalizeAIEndpointType(responses) = %q, want responses", got)
	}
}
