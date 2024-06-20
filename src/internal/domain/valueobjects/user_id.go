package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type UserId struct {
	id string
}

func NewUserId(id string) (UserId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return UserId{}, errors.New("invalid_user_id")
	}

	return UserId{id}, nil
}

func (d UserId) Value() string {
	return d.id
}

func (d UserId) Equals(other UserId) bool {
	return strings.EqualFold(d.id, other.id)
}
