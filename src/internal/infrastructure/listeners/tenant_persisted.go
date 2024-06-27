package listeners

import (
	"context"
	"iyaem/internal/app/commands"
	"iyaem/internal/providers"
	"log"
)

type TenantPersistedHandlers struct {
	addTenantCommand *commands.AddTenantCommand
}

func NewTenantPersistedHandlers(
	addTenantCommand *commands.AddTenantCommand,
) *TenantPersistedHandlers {
	return &TenantPersistedHandlers{
		addTenantCommand: addTenantCommand,
	}
}

func (l *TenantPersistedHandlers) GetHandlers() []providers.Callback {
	return []providers.Callback{
		l.AddCallbackUrl,
	}
}

func (l *TenantPersistedHandlers) AddCallbackUrl(ctx context.Context, payload map[string]interface{}) {

	req := commands.AddTenantRequest{
		OrganizationId: payload["organization_id"].(string),
		TenantId:       payload["tenant_id"].(string),
		ApplicationId:  payload["application_id"].(string),
	}

	_, err := l.addTenantCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

}
