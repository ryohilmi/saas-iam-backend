package valueobjects

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type OrganizationId struct {
	id string
}

func NewOrganizationId(id string) (OrganizationId, error) {
	_, err := uuid.Parse(id)
	if err != nil {
		return OrganizationId{}, errors.New("invalid_organization_id")
	}

	return OrganizationId{id}, nil
}

func GenerateOrganizationId() OrganizationId {
	return OrganizationId{uuid.NewString()}
}

func (d OrganizationId) Value() string {
	return d.id
}

func (d OrganizationId) Equals(other OrganizationId) bool {
	return strings.EqualFold(d.id, other.id)
}
