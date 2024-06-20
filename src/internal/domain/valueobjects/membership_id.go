package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type MembershipId struct {
	id string
}

func NewMembershipId(id string) (MembershipId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return MembershipId{}, errors.New("invalid_user_id")
	}

	return MembershipId{id}, nil
}

func GenerateMembershipId() MembershipId {
	return MembershipId{uuid.NewString()}
}

func (d MembershipId) Value() string {
	return d.id
}

func (d MembershipId) Equals(other UserId) bool {
	return strings.EqualFold(d.id, other.id)
}
