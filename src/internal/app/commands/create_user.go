package commands

import (
	"context"
	"fmt"
	"iyaem/internal/domain/entities"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/domain/valueobjects"
	"log"
)

type CreateUserRequest struct {
	Email   string `json:"email"`
	Picture string `json:"picture"`
	Name    string `json:"name"`
	IdpId   string `json:"idp_id"`
}

type CreateUserCommand struct {
	userRepo repositories.UserRepository
}

func NewCreateUserCommand(
	userRepo repositories.UserRepository,
) *CreateUserCommand {
	return &CreateUserCommand{
		userRepo: userRepo,
	}
}

func (c *CreateUserCommand) Execute(ctx context.Context, r CreateUserRequest) (membershipId string, err error) {

	userId := valueobjects.GenerateUserId()

	identities := make([]valueobjects.Identity, 0)
	identity := valueobjects.NewIdentity(r.IdpId, userId)
	identities = append(identities, identity)

	user := entities.NewUser(
		userId,
		r.Name,
		r.Email,
		r.Picture,
		identities,
		make([]entities.Membership, 0),
	)

	err = c.userRepo.Insert(ctx, &user)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return "", fmt.Errorf("could not create user")
	}

	return userId.Value(), nil
}
