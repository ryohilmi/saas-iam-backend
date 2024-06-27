package events

import (
	"encoding/json"
	"time"
)

type TenantAdded struct {
	TenantId       string    `json:"tenant_id"`
	OrganizationId string    `json:"organization_id"`
	ApplicationId  string    `json:"application_id"`
	Timestamp      time.Time `json:"timestamp"`
}

func NewTenantAdded(tenantId, orgId, appId string) TenantAdded {
	return TenantAdded{TenantId: tenantId, OrganizationId: orgId, ApplicationId: appId}
}

func (k TenantAdded) OccuredOn() time.Time {
	return k.Timestamp
}

func (k TenantAdded) JSON() ([]byte, error) {
	return json.Marshal(k)
}
