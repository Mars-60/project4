package core

import (
	"context"
	"sync"
	"time"
)

type Job func(context.Context) error

type Scheduler struct {
	mu     sync.Mutex
	jobs   map[string]scheduledJob
	stopCh chan struct{}
}

type scheduledJob struct {
	name    string
	at      func(time.Time) time.Time
	job     Job
	holiday HolidayCalendar
}

type HolidayCalendar interface {
	IsHoliday(time.Time) bool
}

type StaticHolidayCalendar struct {
	holidays map[string]bool
}

func NewStaticHolidayCalendar(days []time.Time) StaticHolidayCalendar {
	holidays := make(map[string]bool, len(days))
	for _, day := range days {
		holidays[day.Format("2006-01-02")] = true
	}
	return StaticHolidayCalendar{holidays: holidays}
}

func (c StaticHolidayCalendar) IsHoliday(day time.Time) bool {
	return c.holidays[day.Format("2006-01-02")]
}

func NewScheduler() *Scheduler {
	return &Scheduler{jobs: make(map[string]scheduledJob), stopCh: make(chan struct{})}
}

func (s *Scheduler) RegisterMarketOpen(name string, hour int, minute int, calendar HolidayCalendar, job Job) {
	s.register(name, dailyAt(hour, minute), calendar, job)
}

func (s *Scheduler) RegisterMarketClose(name string, hour int, minute int, calendar HolidayCalendar, job Job) {
	s.register(name, dailyAt(hour, minute), calendar, job)
}

func (s *Scheduler) RegisterSquareOff(name string, hour int, minute int, calendar HolidayCalendar, job Job) {
	s.register(name, dailyAt(hour, minute), calendar, job)
}

func (s *Scheduler) register(name string, at func(time.Time) time.Time, calendar HolidayCalendar, job Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs[name] = scheduledJob{name: name, at: at, holiday: calendar, job: job}
}

func (s *Scheduler) Start(ctx context.Context) {
	s.mu.Lock()
	jobs := make([]scheduledJob, 0, len(s.jobs))
	for _, job := range s.jobs {
		jobs = append(jobs, job)
	}
	s.mu.Unlock()

	for _, job := range jobs {
		go s.run(ctx, job)
	}
}

func (s *Scheduler) Stop() {
	close(s.stopCh)
}

func (s *Scheduler) run(ctx context.Context, job scheduledJob) {
	for {
		now := time.Now()
		next := job.at(now)
		if !next.After(now) {
			next = job.at(now.Add(24 * time.Hour))
		}
		timer := time.NewTimer(time.Until(next))
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-s.stopCh:
			timer.Stop()
			return
		case <-timer.C:
			if job.holiday == nil || !job.holiday.IsHoliday(time.Now()) {
				_ = job.job(ctx)
			}
		}
	}
}

func dailyAt(hour int, minute int) func(time.Time) time.Time {
	return func(base time.Time) time.Time {
		return time.Date(base.Year(), base.Month(), base.Day(), hour, minute, 0, 0, base.Location())
	}
}
