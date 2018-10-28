// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package model

import (
	"encoding/json"
	"io"
)

const (
	CMD_REMIND            = "remind"
	REMIND_BOTNAME        = "remindbot"
	REMIND_EXCEPTION_TEXT = "api.command_remind.exception"
	REMIND_HELP_TEXT      = "api.command_remind.help"
	DEFAULT_TIME          = "9:00AM"
)

type Reminders []Reminder
type Occurrences []Occurrence

type Reminder struct {
	Id        string `json:"id"`
	TeamId    string `json:"team_id"`
	UserId    string `json:"user_id"`
	Target    string `json:"target"`
	Message   string `json:"message"`
	When      string `json:"when"`
	Completed string `json:"completed"`
}

type Occurrence struct {
	Id         string `json:"id"`
	UserId     string `json:"user_id"`
	ReminderId string `json:"reminder_id"`
	Repeat     string `json:"repeat"`
	Occurrence string `json:"occurrence"`
	Snoozed    string `json:"snoozed"`
}

type ReminderRequest struct {
	TeamId      string      `json:"team_id"`
	UserId      string      `json:"user_id"`
	Payload     string      `json:"payload"`
	Reminder    Reminder    `json:"reminder"`
	Occurrences Occurrences `json:"occurrences"`
}

func (r *Reminder) ToJson() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func ReminderFromJson(data io.Reader) *Reminder {
	var r *Reminder
	json.NewDecoder(data).Decode(&r)
	return r
}

func (o *Occurrence) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func OccurrenceFromJson(data io.Reader) *Occurrence {
	var o *Occurrence
	json.NewDecoder(data).Decode(&o)
	return o
}

func (r *ReminderRequest) ToJson() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func ReminderRequestFromJson(data io.Reader) *ReminderRequest {
	var r *ReminderRequest
	json.NewDecoder(data).Decode(&r)
	return r
}