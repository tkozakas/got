package scheduler

import (
	"context"
	"log/slog"

	"github.com/robfig/cron/v3"
)

type Job struct {
	Name     string
	Schedule string
	Func     func(ctx context.Context) error
}

type Scheduler struct {
	cron *cron.Cron
	jobs []Job
}

func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(cron.WithSeconds()),
	}
}

func (s *Scheduler) Register(job Job) error {
	_, err := s.cron.AddFunc(job.Schedule, func() {
		ctx := context.Background()
		if err := job.Func(ctx); err != nil {
			slog.Error("Job failed", "name", job.Name, "error", err)
			return
		}
		slog.Info("Job completed", "name", job.Name)
	})
	if err != nil {
		return err
	}
	s.jobs = append(s.jobs, job)
	return nil
}

func (s *Scheduler) Start() {
	slog.Info("Scheduler started", "jobs", len(s.jobs))
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}
