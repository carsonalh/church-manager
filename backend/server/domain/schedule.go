package domain

import (
	"fmt"
	"time"
)

type Schedule struct {
	id                  uint64
	beginDate           time.Time
	endDate             *time.Time
	repeatInterval      *scheduleRepeatInterval
	repeatNthDayOfMonth *scheduleRepeatNthDayOfMonth
}

type scheduleRepeatInterval struct {
	count uint
	unit  ScheduleRepeatUnit
}

type scheduleRepeatNthDayOfMonth struct {
	day ScheduleDayOfWeek
	n   int
}

type ScheduleDayOfWeek string

const (
	DayMonday    ScheduleDayOfWeek = "Monday"
	DayTuesday   ScheduleDayOfWeek = "Tuesday"
	DayWednesday ScheduleDayOfWeek = "Wednesday"
	DayThursday  ScheduleDayOfWeek = "Thursday"
	DayFriday    ScheduleDayOfWeek = "Friday"
	DaySaturday  ScheduleDayOfWeek = "Saturday"
	DaySunday    ScheduleDayOfWeek = "Sunday"
)

type ScheduleRepeatUnit string

const (
	RepeatUnitDay   ScheduleRepeatUnit = "Day"
	RepeatUnitWeek  ScheduleRepeatUnit = "Week"
	RepeatUnitMonth ScheduleRepeatUnit = "Month"
	RepeatUnitYear  ScheduleRepeatUnit = "Year"
)

func (schedule *Schedule) ToResponseDTO() *ScheduleResponseDTO {
	var repeatInterval *ScheduleResponseDTORepeatInterval
	var repeatNthDayOfMonth *ScheduleResponseDTORepeatNthDayOfMonth

	if schedule.repeatInterval != nil {
		repeatInterval = &ScheduleResponseDTORepeatInterval{
			Count: schedule.repeatInterval.count,
			Unit:  schedule.repeatInterval.unit,
		}
	}

	if schedule.repeatNthDayOfMonth != nil {
		repeatNthDayOfMonth = &ScheduleResponseDTORepeatNthDayOfMonth{
			Day: schedule.repeatNthDayOfMonth.day,
			N:   schedule.repeatNthDayOfMonth.n,
		}
	}

	return &ScheduleResponseDTO{
		Id:                  schedule.id,
		BeginDate:           schedule.beginDate,
		EndDate:             schedule.endDate,
		RepeatInterval:      repeatInterval,
		RepeatNthDayOfMonth: repeatNthDayOfMonth,
	}
}

type ScheduleRow struct {
	Id                     *uint64
	BeginDate              *time.Time
	EndDate                *time.Time
	RepeatIntervalCount    *uint
	RepeatIntervalUnit     *ScheduleRepeatUnit
	RepeatNthDayOfMonthDay *ScheduleDayOfWeek
	RepeatNthDayOfMonthN   *int
}

func (row *ScheduleRow) ToSchedule() (*Schedule, error) {
	if row.Id == nil {
		return nil, fmt.Errorf("id cannot be nil")
	}

	if row.BeginDate == nil {
		return nil, fmt.Errorf("BeginDate cannot be nil")
	}

	if (row.RepeatIntervalCount == nil) != (row.RepeatIntervalUnit == nil) {
		return nil, fmt.Errorf("repeat interval count and unit must either both or both not be defined")
	}

	if (row.RepeatNthDayOfMonthDay == nil) != (row.RepeatNthDayOfMonthN == nil) {
		return nil, fmt.Errorf("repeat nth day of month day and N must either both or both not be defined")
	}

	if (row.RepeatIntervalCount == nil) == (row.RepeatNthDayOfMonthDay == nil) {
		return nil, fmt.Errorf("one of the repeat interval fields or the repeat nth day of the month field must be nil and the other non-nil")
	}

	var repeatInterval *scheduleRepeatInterval
	var repeatNthDayOfMonth *scheduleRepeatNthDayOfMonth

	if row.RepeatIntervalCount != nil {
		repeatInterval = &scheduleRepeatInterval{
			count: *row.RepeatIntervalCount,
			unit:  *row.RepeatIntervalUnit,
		}
	}

	if row.RepeatNthDayOfMonthDay != nil {
		repeatNthDayOfMonth = &scheduleRepeatNthDayOfMonth{
			day: *row.RepeatNthDayOfMonthDay,
			n:   *row.RepeatNthDayOfMonthN,
		}
	}

	schedule := &Schedule{
		id:                  *row.Id,
		beginDate:           *row.BeginDate,
		endDate:             row.EndDate,
		repeatInterval:      repeatInterval,
		repeatNthDayOfMonth: repeatNthDayOfMonth,
	}

	return schedule, nil
}
