package domain

import "time"

// WorkingHours represents a doctor's recurring weekly availability for a specific day.
type WorkingHours struct {
	ID        int64  `json:"id"`
	OrgID     int64  `json:"org_id"`
	DoctorID  int64  `json:"doctor_id"`
	DayOfWeek int    `json:"day_of_week"` // 0=Sunday..6=Saturday
	StartTime string `json:"start_time"`  // "09:00"
	EndTime   string `json:"end_time"`    // "17:00"
}

// ScheduleOverride represents a one-off change to a doctor's schedule (day off, modified hours).
type ScheduleOverride struct {
	ID        int64     `json:"id"`
	OrgID     int64     `json:"org_id"`
	DoctorID  int64     `json:"doctor_id"`
	Date      time.Time `json:"date"`
	StartTime *string   `json:"start_time"` // nil = day off
	EndTime   *string   `json:"end_time"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

// TimeSlot represents a bookable time window.
type TimeSlot struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
