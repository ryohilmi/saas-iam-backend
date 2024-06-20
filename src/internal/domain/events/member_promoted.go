package events

import (
	"encoding/json"
	"time"
)

type MemberPromoted struct {
	MembershipId string    `json:"membership_id"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewMemberPromoted(membershipId string) MemberPromoted {
	return MemberPromoted{MembershipId: membershipId}
}

func (k MemberPromoted) OccuredOn() time.Time {
	return k.Timestamp
}

func (k MemberPromoted) JSON() ([]byte, error) {
	return json.Marshal(k)
}
