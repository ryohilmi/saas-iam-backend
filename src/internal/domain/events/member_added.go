package events

import (
	"encoding/json"
	"time"
)

type MemberAdded struct {
	MembershipId string    `json:"membership_id"`
	UserId       string    `json:"user_org_id"`
	Level        string    `json:"level"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewMemberAdded(membershipId, userId, level string) MemberAdded {
	return MemberAdded{MembershipId: membershipId, UserId: userId, Level: level}
}

func (k MemberAdded) OccuredOn() time.Time {
	return k.Timestamp
}

func (k MemberAdded) JSON() ([]byte, error) {
	return json.Marshal(k)
}
