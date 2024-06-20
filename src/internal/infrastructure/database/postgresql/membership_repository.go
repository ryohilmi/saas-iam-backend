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
		SELECT id, user_id, organization_id, level FROM user_organization WHERE id=$1;`, id,
	)
	err := row.Scan(&memberRecord.Id, &memberRecord.UserId, &memberRecord.OrganizationId, &memberRecord.Level)
	if err != nil {
		log.Printf("Error: %v", err)
		return nil, err
	}

	memId, _ := valueobjects.NewMembershipId(memberRecord.Id)
	userId, _ := valueobjects.NewUserId(memberRecord.UserId)
	orgId, _ := valueobjects.NewOrganizationId(memberRecord.OrganizationId)

	membership = entities.NewMembership(
		memId,
		userId,
		orgId,
		valueobjects.MembershipLevel(memberRecord.Level),
		make([]entities.Role, 0),
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

	membership = entities.NewMembership(
		memId,
		userId,
		orgId,
		valueobjects.MembershipLevel(memberRecord.Level),
		make([]entities.Role, 0),
	)

	return &membership, nil
}
