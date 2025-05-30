package models

import "errors"

type Event struct {
	ID          string
	EventName   string
	Date        string
	StartTime   string
	EndTime     string
	Location    string
	Description string
	Color       string
	Creator     User
	Invitees    []User
}

func NewEvent(id string, name string, date string, starttime string, endtime string, location string, description string, color string, creator User, invitees []User) *Event {
	return &Event{
		ID:          id,
		EventName:   name,
		Date:        date,
		StartTime:   starttime,
		EndTime:     endtime,
		Location:    location,
		Description: description,
		Color:       color,
		Creator:     creator,
		Invitees:    invitees,
	}
}

func (eventModel *Event) Update(newData Event) error {
	if newData.ID != eventModel.ID {
		return errors.New("Invalid ID when trying to update fields in Event")
	}
	if newData.EventName != "" {
		eventModel.EventName = newData.EventName
	}
	if newData.Date != "" {
		eventModel.Date = newData.Date
	}
	if newData.StartTime != "" {
		eventModel.StartTime = newData.StartTime
	}
	if newData.EndTime != "" {
		eventModel.EndTime = newData.EndTime
	}
	if newData.Location != "" {
		eventModel.Location = newData.Location
	}
	if newData.Description != "" {
		eventModel.Description = newData.Description
	}
	if newData.Color != "" {
		eventModel.Color = newData.Color
	}
	if len(newData.Invitees) > 0 {
		eventModel.Invitees = newData.Invitees
	}
	return nil
}
