package domain

import (
	"fmt"
	"time"
)

type ScheduleCreateDTO struct {
	BeginDate           *time.Time                            `json:"beginDate"`
	EndDate             *time.Time                            `json:"endDate"`
	RepeatInterval      *ScheduleCreateDTORepeatInterval      `json:"repeatInterval"`
	RepeatNthDayOfMonth *ScheduleCreateDTORepeatNthDayOfMonth `json:"repeatNthDayOfMonth"`
}

type ScheduleCreateDTORepeatInterval struct {
	Count uint               `json:"count"`
	Unit  ScheduleRepeatUnit `json:"unit"`
}

type ScheduleCreateDTORepeatNthDayOfMonth struct {
	Day ScheduleDayOfWeek `json:"day"`
	N   int               `json:"n"`
}

func (dto *ScheduleCreateDTO) Validate() []error {
	errs := make([]error, 0)

	if dto.BeginDate == nil {
		errs = append(errs, fmt.Errorf("field beginDate cannot be null or missing"))
	}

	if (dto.RepeatInterval == nil) == (dto.RepeatNthDayOfMonth == nil) {
		if dto.RepeatInterval == nil {
			// both are nil
			errs = append(errs, fmt.Errorf("exactly one of repeatInterval and repeatNthDayOfMonth must be present"))
		} else {
			errs = append(errs, fmt.Errorf("exactly one of repeatInterval and repeatNthDayOfMonth must be absent (or null)"))
		}
	}

	if dto.RepeatInterval != nil {
		if dto.RepeatInterval.Count < 1 {
			errs = append(errs, fmt.Errorf("repeatInterval.count must be >= 1, got %d", dto.RepeatInterval.Count))
		}
	}

	if dto.RepeatNthDayOfMonth != nil {
		if dto.RepeatNthDayOfMonth.N == 0 {
			errs = append(errs, fmt.Errorf("repeatNthDayOfMonth.n cannot be "+
				"zero, it must either be positive, indicating the first, second, etc. "+
				"<day of week> of the month, or the first, second, etc. last <day of "+
				"week> of the month.\nE.g. n: -2 with day: \"Tuesday\" is the second "+
				"last day of the month"))
		}
	}

	return errs
}
