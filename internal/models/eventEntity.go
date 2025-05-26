package models

import "github.com/Azure/azure-sdk-for-go/sdk/data/aztables"

type EventEntity struct {
	aztables.Entity
	EventName  string
	Date       string
	StartTime  string
	EndTime    string
	CreatorID  string `json:"Creator"`
	InviteeIDs string `json:"Invitees"`
}
