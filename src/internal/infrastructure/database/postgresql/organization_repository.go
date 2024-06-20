package postgresql

import (
	"context"
	"database/sql"
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

		member := entities.NewMembership(
			id,
			userId,
			orgId,
			valueobjects.MembershipLevel(memberRecord.Level),
			make([]valueobjects.UserRole, 0),
		)

		members = append(members, member)
	}

	org = entities.NewOrganization(orgId, orgRecord.Name, orgRecord.Identifier, members, make([]entities.Tenant, 0))

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

		member := entities.NewMembership(
			id,
			userId,
			orgId,
			valueobjects.MembershipLevel(memberRecord.Level),
			make([]valueobjects.UserRole, 0),
		)

		members = append(members, member)
	}

	org = entities.NewOrganization(orgId, orgRecord.Name, orgRecord.Identifier, members, make([]entities.Tenant, 0))

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
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
