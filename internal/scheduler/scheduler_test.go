package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestSchedulerNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("expected scheduler, got nil")
	}
	if s.cron == nil {
		t.Error("expected cron instance, got nil")
	}
}

func TestSchedulerRegister(t *testing.T) {
	tests := []struct {
		name     string
		schedule string
		wantErr  bool
	}{
		{
			name:     "ValidSchedule",
			schedule: "0 * * * * *",
			wantErr:  false,
		},
		{
			name:     "InvalidSchedule",
			schedule: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New()
			err := s.Register(Job{
				Name:     "test",
				Schedule: tt.schedule,
				Func:     func(ctx context.Context) error { return nil },
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("Register() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(s.jobs) != 1 {
				t.Errorf("want 1 job, got %d", len(s.jobs))
			}
		})
	}
}

func TestSchedulerStartStop(t *testing.T) {
	s := New()

	executed := make(chan bool, 1)
	_ = s.Register(Job{
		Name:     "quick",
		Schedule: "* * * * * *",
		Func: func(ctx context.Context) error {
			select {
			case executed <- true:
			default:
			}
			return nil
		},
	})

	s.Start()
	defer s.Stop()

	select {
	case <-executed:
	case <-time.After(2 * time.Second):
		t.Error("job was not executed within timeout")
	}
}

func TestSchedulerJobError(t *testing.T) {
	s := New()

	_ = s.Register(Job{
		Name:     "failing",
		Schedule: "* * * * * *",
		Func: func(ctx context.Context) error {
			return errors.New("job error")
		},
	})

	s.Start()
	time.Sleep(1500 * time.Millisecond)
	s.Stop()
}
