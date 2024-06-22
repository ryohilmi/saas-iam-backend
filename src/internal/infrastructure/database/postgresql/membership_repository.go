package postgresql

import (
	"context"
	"database/sql"
	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
	"log"
)

type MembershipRepository struct {
	db *sql.DB
}

func NewMembershipRepository(db *sql.DB) repositories.MembershipRepository {
	return &MembershipRepository{
		db: db,
	}
}

func (r *MembershipRepository) FindById(ctx context.Context, id string) (*entities.Membership, error) {

	var membership entities.Membership
	var memberRecord struct {
		Id             string
		UserId         string
		OrganizationId string
		Level          string
	}

	row := r.db.QueryRow(`	
		SELECT uo.id, user_id, organization_id, level 
		FROM user_organization uo
		LEFT JOIN public.user u ON uo.user_id = u.id
		WHERE uo.id=$1`, id,
	)
	err := row.Scan(&memberRecord.Id, &memberRecord.UserId, &memberRecord.OrganizationId, &memberRecord.Level)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	memId, _ := valueobjects.NewMembershipId(memberRecord.Id)
	userId, _ := valueobjects.NewUserId(memberRecord.UserId)
	orgId, _ := valueobjects.NewOrganizationId(memberRecord.OrganizationId)

	userRoles := make([]valueobjects.UserRole, 0)

	var roleRecord struct {
		MembershipId string
		RoleId       string
		TenantId     string
	}

	rows, err := r.db.Query(`
		SELECT user_org_id, role_id, tenant_id
		FROM user_role
		WHERE user_org_id=$1`, memberRecord.Id,
	)

	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&roleRecord.MembershipId, &roleRecord.RoleId, &roleRecord.TenantId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		roleId, _ := valueobjects.NewRoleId(roleRecord.RoleId)
		tenantId, _ := valueobjects.NewTenantId(roleRecord.TenantId)

		userRole := valueobjects.NewUserRole(memId, roleId, tenantId)
		userRoles = append(userRoles, userRole)
	}

	membership = entities.NewMembership(
		memId,
		userId,
		orgId,
		valueobjects.MembershipLevel(memberRecord.Level),
		userRoles,
		make([]valueobjects.UserGroup, 0),
	)

	return &membership, nil
}

func (r *MembershipRepository) FindByEmail(ctx context.Context, email string) (*entities.Membership, error) {

	var membership entities.Membership
	var memberRecord struct {
		Id             string
		UserId         string
		OrganizationId string
		Level          string
	}

	row := r.db.QueryRow(`	
		SELECT id, user_id, organization_id, level 
		FROM user_organization uo
		LEFT JOIN public.user u ON uo.user_id = u.id
		WHERE email=$1`, email,
	)
	err := row.Scan(&memberRecord.Id, &memberRecord.UserId, &memberRecord.OrganizationId, &memberRecord.Level)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	memId, _ := valueobjects.NewMembershipId(memberRecord.Id)
	userId, _ := valueobjects.NewUserId(memberRecord.UserId)
	orgId, _ := valueobjects.NewOrganizationId(memberRecord.OrganizationId)

	userRoles := make([]valueobjects.UserRole, 0)

	var roleRecord struct {
		MembershipId string
		RoleId       string
		TenantId     string
	}

	rows, err := r.db.Query(`
		SELECT user_org_id, role_id, tenant_id
		FROM user_role
		WHERE user_org_id=$1`, memberRecord.Id,
	)

	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&roleRecord.MembershipId, &roleRecord.RoleId, &roleRecord.TenantId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		roleId, _ := valueobjects.NewRoleId(roleRecord.RoleId)
		tenantId, _ := valueobjects.NewTenantId(roleRecord.TenantId)

		userRole := valueobjects.NewUserRole(memId, roleId, tenantId)
		userRoles = append(userRoles, userRole)
	}

	membership = entities.NewMembership(
		memId,
		userId,
		orgId,
		valueobjects.MembershipLevel(memberRecord.Level),
		userRoles,
		make([]valueobjects.UserGroup, 0),
	)

	return &membership, nil
}
