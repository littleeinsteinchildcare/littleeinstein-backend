package models

type Event struct {
	ID        string
	EventName string
	Date      string
	StartTime string
	EndTime   string
	Creator   User
	Invitees  []User
}

func NewEvent(id string, name string, date string, starttime string, endtime string, creator User, invitees []User) *Event {
	return &Event{
		ID:        id,
		EventName: name,
		Date:      date,
		StartTime: starttime,
		EndTime:   endtime,
		Creator:   creator,
		Invitees:  invitees,
	}
}
