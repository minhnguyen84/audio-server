package metadata

import "time"

type EventType string
type StatusType string

const (
	UPDATE EventType = "update"
	CREATE EventType = "create"

	CREATED  StatusType = "created"
	ONGOING  StatusType = "on_going"
	FINISHED StatusType = "finished"
)

type Event struct {
	Type       EventType
	SessionId  string
	Parameters interface{}
	Trace      interface{}
}

type CreateParameters struct {
	StartTime         time.Time
	CallerPhoneNumber *string //ani : peut-etre null
}

type UpdateParameters struct {
	Status            StatusType
	StopTime          *time.Time
	CallerPhoneNumber *string //ani : peut-etre null
	Language          *string
}
