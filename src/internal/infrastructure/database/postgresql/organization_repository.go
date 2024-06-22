package postgresql

import (
	"context"
	"database/sql"
	"encoding/json"
	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/events"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
	"log"
)

type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) repositories.OrganizationRepository {
	return &OrganizationRepository{
		db: db,
	}
}

func (r *OrganizationRepository) FindById(ctx context.Context, id string) (*entities.Organization, error) {

	// Retreive Organization Data
	var org entities.Organization
	var orgRecord struct {
		Id         string
		Name       string
		Identifier string
	}

	row := r.db.QueryRow(`	
		SELECT id, name, identifier FROM organization WHERE id=$1;`, id,
	)
	err := row.Scan(&orgRecord.Id, &orgRecord.Name, &orgRecord.Identifier)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	orgId, err := valueobjects.NewOrganizationId(orgRecord.Id)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	// Retreive Members Data
	var members []entities.Membership
	var memberRecord struct {
		Id             string
		UserId         string
		OrganizationId string
		Level          string
		Roles          json.RawMessage
		Groups         json.RawMessage
	}

	rows, err := r.db.Query(`
		select uo.id, uo.user_id, uo.organization_id, level, coalesce(json_agg(
			json_build_object('role_id', ur.role_id, 'tenant_id', ur.tenant_id)
		) filter (where ur.role_id notnull), '[]') roles, coalesce(json_agg(
			json_build_object('group_id', ug.group_id, 'tenant_id', ug.tenant_id)
		) filter (where ug.group_id notnull), '[]') groups
		from user_organization uo
		left join user_role ur on uo.id = ur.user_org_id
		left join user_group ug on uo.id = ug.user_org_id 
		where organization_id=$1
		group by uo.id, uo.organization_id, level`, orgId.Value(),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&memberRecord.Id, &memberRecord.UserId, &memberRecord.OrganizationId, &memberRecord.Level, &memberRecord.Roles, &memberRecord.Groups)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		log.Printf("Roles: %v", string(memberRecord.Roles))
		log.Printf("Groups: %v", string(memberRecord.Groups))

		id, err := valueobjects.NewMembershipId(memberRecord.Id)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		userId, err := valueobjects.NewUserId(memberRecord.UserId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		userRoles := make([]valueobjects.UserRole, 0)
		userGroups := make([]valueobjects.UserGroup, 0)

		var roleRecord []interface{}

		err = json.Unmarshal(memberRecord.Roles, &roleRecord)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		for _, role := range roleRecord {
			roleId, _ := valueobjects.NewRoleId(role.(map[string]interface{})["role_id"].(string))
			tenantId, _ := valueobjects.NewTenantId(role.(map[string]interface{})["tenant_id"].(string))

			userRole := valueobjects.NewUserRole(id, roleId, tenantId)
			userRoles = append(userRoles, userRole)
		}

		var groupRecord []interface{}

		err = json.Unmarshal(memberRecord.Groups, &groupRecord)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		for _, group := range groupRecord {
			groupId, _ := valueobjects.NewGroupId(group.(map[string]interface{})["group_id"].(string))
			tenantId, _ := valueobjects.NewTenantId(group.(map[string]interface{})["tenant_id"].(string))

			userGroup := valueobjects.NewUserGroup(id, groupId, tenantId)
			userGroups = append(userGroups, userGroup)
		}

		log.Printf("UserRoles: %v", userRoles)

		member := entities.NewMembership(
			id,
			userId,
			orgId,
			valueobjects.MembershipLevel(memberRecord.Level),
			userRoles,
			userGroups,
		)

		members = append(members, member)
	}

	// Retreive Tenants Data
	var tenants []entities.Tenant
	var tenantRecord struct {
		Id             string
		OrganizationId string
		ApplilcationId string
	}

	rows, err = r.db.Query(`
		SELECT id, org_id, app_id FROM tenant WHERE org_id=$1;`, orgId.Value(),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&tenantRecord.Id, &tenantRecord.OrganizationId, &tenantRecord.ApplilcationId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		id, err := valueobjects.NewTenantId(tenantRecord.Id)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		applicationId, err := valueobjects.NewApplicationId(tenantRecord.ApplilcationId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		tenant := entities.NewTenant(
			id,
			orgId,
			applicationId,
			make([]entities.Role, 0),
		)

		tenants = append(tenants, tenant)
	}

	org = entities.NewOrganization(orgId, orgRecord.Name, orgRecord.Identifier, members, tenants)

	log.Printf("Organization: %v", org)

	return &org, nil
}

func (r *OrganizationRepository) FindByIdentifier(ctx context.Context, identifier string) (*entities.Organization, error) {

	var org entities.Organization
	var orgRecord struct {
		Id         string
		Name       string
		Identifier string
	}

	row := r.db.QueryRow(`	
		SELECT id, name, identifier FROM organization WHERE identifier=$1;`, identifier,
	)
	err := row.Scan(&orgRecord.Id, &orgRecord.Name, &orgRecord.Identifier)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	orgId, err := valueobjects.NewOrganizationId(orgRecord.Id)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	var members []entities.Membership
	var memberRecord struct {
		Id             string
		UserId         string
		OrganizationId string
		Level          string
	}

	rows, err := r.db.Query(`
		SELECT id, user_id, organization_id, level FROM user_organization WHERE organization_id=$1;`, orgId.Value(),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&memberRecord.Id, &memberRecord.UserId, &memberRecord.OrganizationId, &memberRecord.Level)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		id, err := valueobjects.NewMembershipId(memberRecord.Id)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		userId, err := valueobjects.NewUserId(memberRecord.UserId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

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

			userRole := valueobjects.NewUserRole(id, roleId, tenantId)
			userRoles = append(userRoles, userRole)
		}

		log.Printf("UserRoles: %v", userRoles)

		member := entities.NewMembership(
			id,
			userId,
			orgId,
			valueobjects.MembershipLevel(memberRecord.Level),
			userRoles,
			make([]valueobjects.UserGroup, 0),
		)

		members = append(members, member)
	}

	// Retreive Tenants Data
	var tenants []entities.Tenant
	var tenantRecord struct {
		Id             string
		OrganizationId string
		ApplilcationId string
	}

	rows, err = r.db.Query(`
		SELECT id, org_id, app_id FROM tenant WHERE org_id=$1;`, orgId.Value(),
	)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(&tenantRecord.Id, &tenantRecord.OrganizationId, &tenantRecord.ApplilcationId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		id, err := valueobjects.NewTenantId(tenantRecord.Id)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		applicationId, err := valueobjects.NewApplicationId(tenantRecord.ApplilcationId)
		if err != nil {
			log.Printf("Error: %v", err)
			return nil, err
		}

		tenant := entities.NewTenant(
			id,
			orgId,
			applicationId,
			make([]entities.Role, 0),
		)

		tenants = append(tenants, tenant)
	}

	org = entities.NewOrganization(orgId, orgRecord.Name, orgRecord.Identifier, members, tenants)

	log.Printf("Organization: %v", org)

	return &org, nil
}

func (r *OrganizationRepository) Insert(ctx context.Context, org *entities.Organization) error {

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO organization (id, name, identifier) VALUES ($1, $2, $3);`,
		org.Id().Value(), org.Name(), org.Identifier(),
	)

	if err != nil {
		return err
	}

	for _, member := range org.Members() {
		_, err = tx.Exec(
			`INSERT INTO user_organization (id, organization_id, user_id, level) VALUES ($1, $2, $3, $4);`,
			member.Id().Value(), org.Id().Value(), member.UserId().Value(), member.Level(),
		)

		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (r *OrganizationRepository) Update(ctx context.Context, org *entities.Organization) error {

	tx, err := r.db.Begin()

	if err != nil {
		return err
	}

	managerCount := 0
	for _, member := range org.Members() {
		if member.Level() == "manager" {
			managerCount++
		}
	}

	_, err = tx.Exec(`
		UPDATE organization SET name=$1, identifier=$2, tenant_count=$4, member_count=$5, manager_count=$6 WHERE id=$3;`,
		org.Name(), org.Identifier(), org.Id().Value(), len(org.Tenants()), len(org.Members()), managerCount,
	)

	if err != nil {

		return err
	}

	for _, event := range org.Events() {
		switch e := event.(type) {
		case events.MemberAdded:
			_, err = tx.Exec(`
				INSERT INTO user_organization 
					(id, organization_id, user_id, level) 
				VALUES ($1, $2, $3, $4);`,
				e.MembershipId, org.Id().Value(), e.UserId, e.Level,
			)

			if err != nil {
				return err
			}
		case events.MemberPromoted:
			_, err = tx.Exec(`
				UPDATE user_organization SET level='manager' WHERE id=$1;`,
				e.MembershipId,
			)

			if err != nil {
				return err
			}
		case events.MemberDemoted:
			_, err = tx.Exec(`
				UPDATE user_organization SET level='member' WHERE id=$1;`,
				e.MembershipId,
			)

			if err != nil {
				return err
			}
		case events.RoleAddedToMember:
			_, err = tx.Exec(`
				INSERT INTO user_role (user_org_id, role_id, tenant_id) VALUES ($1, $2, $3);`,
				e.MembershipId, e.RoleId, e.TenantId,
			)

			if err != nil {
				return err
			}
		case events.RoleRemovedFromMember:
			_, err = tx.Exec(`
				DELETE FROM user_role WHERE user_org_id=$1 AND role_id=$2 AND tenant_id=$3;`,
				e.MembershipId, e.RoleId, e.TenantId,
			)

			log.Printf("Id: %v, RoleId: %v, TenantId: %v", e.MembershipId, e.RoleId, e.TenantId)

			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
