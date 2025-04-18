package store

import (
	"context"

	"github.com/carsonalh/churchmanagerbackend/server/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ScheduleStore struct {
	pool *pgxpool.Pool
}

func CreateScheduleStore(pool *pgxpool.Pool) *ScheduleStore {
	return &ScheduleStore{
		pool: pool,
	}
}

func (store *ScheduleStore) Create(createDto *domain.ScheduleCreateDTO) (*domain.Schedule, error) {
	var count *uint
	var unit *domain.ScheduleRepeatUnit
	var day *domain.ScheduleDayOfWeek
	var n *int

	if createDto.RepeatInterval != nil {
		count = &createDto.RepeatInterval.Count
		unit = &createDto.RepeatInterval.Unit
	}

	if createDto.RepeatNthDayOfMonth != nil {
		day = &createDto.RepeatNthDayOfMonth.Day
		n = &createDto.RepeatNthDayOfMonth.N
	}

	row := domain.ScheduleRow{
		BeginDate:              createDto.BeginDate,
		EndDate:                createDto.EndDate,
		RepeatIntervalCount:    count,
		RepeatIntervalUnit:     unit,
		RepeatNthDayOfMonthDay: day,
		RepeatNthDayOfMonthN:   n,
	}

	err := store.pool.QueryRow(
		context.Background(),
		"INSERT INTO church_service_schedule (\n"+
			"begin_date, end_date, repeat_interval_count, repeat_interval_unit,\n"+
			"repeat_nth_day_of_month_day, repeat_nth_day_of_month_n)\n"+
			"VALUES ($1, $2, $3, $4, $5, $6)\n"+
			"RETURNING id;",
		row.BeginDate, row.EndDate,
		row.RepeatIntervalCount, row.RepeatIntervalUnit,
		row.RepeatNthDayOfMonthDay, row.RepeatNthDayOfMonthN,
	).Scan(&row.Id)
	if err != nil {
		return nil, err
	}

	schedule, err := row.ToSchedule()
	if err != nil {
		return nil, err
	}

	return schedule, nil
}
