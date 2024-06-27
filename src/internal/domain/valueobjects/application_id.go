package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type ApplicationId struct {
	id string
}

func NewApplicationId(id string) (ApplicationId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return ApplicationId{}, errors.New("invalid_application_id")
	}

	return ApplicationId{id}, nil
}

func GenerateApplicationId() ApplicationId {
	return ApplicationId{uuid.NewString()}
}

func (d ApplicationId) Value() string {
	return d.id
}

func (d ApplicationId) Equals(other ApplicationId) bool {
	return strings.EqualFold(d.id, other.id)
}
