package domain

import (
	"time"
)

type ScheduleResponseDTO struct {
	Id                  uint64                                  `json:"id"`
	BeginDate           time.Time                               `json:"beginDate"`
	EndDate             *time.Time                              `json:"endDate"`
	RepeatInterval      *ScheduleResponseDTORepeatInterval      `json:"repeatInterval"`
	RepeatNthDayOfMonth *ScheduleResponseDTORepeatNthDayOfMonth `json:"repeatNthDayOfMonth"`
}

type ScheduleResponseDTORepeatInterval struct {
	Count uint               `json:"count"`
	Unit  ScheduleRepeatUnit `json:"unit"`
}

type ScheduleResponseDTORepeatNthDayOfMonth struct {
	Day ScheduleDayOfWeek `json:"day"`
	N   int               `json:"n"`
}
