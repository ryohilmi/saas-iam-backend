package valueobjects

import (
	"errors"
)

type MembershipLevel string

func NewMembershipLevel(level string) (MembershipLevel, error) {

	if level != "owner" && level != "manager" && level != "member" {
		return "", errors.New("invalid_membership_level")
	}

	return MembershipLevel(level), nil
}
