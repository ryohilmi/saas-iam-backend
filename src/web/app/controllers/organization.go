package controllers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"iyaem/platform/authenticator"
)

type OrganizationController struct {
	auth *authenticator.Authenticator
	db   *sql.DB
}

func NewOrganizationController(auth *authenticator.Authenticator, db *sql.DB) *OrganizationController {
	return &OrganizationController{auth, db}
}

func (c *OrganizationController) GetAffiliatedOrganizations(ctx *gin.Context) {
	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		ctx.String(http.StatusUnauthorized, "Unauthorized")
	}

	token := authorizationHeader[len("Bearer "):]
	claims, err := DecodeJWT(token)
	if err != nil {
		log.Print(err)
		ctx.String(http.StatusUnauthorized, "Unauthorized")
	}

	rows, err := c.db.Query(`	SELECT organization_id, name FROM user_organization 
								LEFT JOIN organization ON user_organization.organization_id = organization.id
								WHERE user_id = $1;`, claims["sub"])
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to get affiliated organizations")
	}

	type organization struct {
		OrganizationId string `json:"organization_id"`
		Name           string `json:"name"`
	}
	var organizations []organization

	for rows.Next() {
		var organizationId string
		var name string
		err = rows.Scan(&organizationId, &name)
		if err != nil {
			log.Printf("Error: %v", err)
			ctx.String(http.StatusInternalServerError, "Failed to get affiliated organizations")
		}

		organizations = append(organizations, organization{organizationId, name})
	}

	ctx.JSON(http.StatusOK, organizations)

}

func (c *OrganizationController) CreateOrganization(ctx *gin.Context) {
	type OrgParams struct {
		Name       string `json:"name"`
		Identifier string `json:"identifier"`
	}

	jsonData, err := ctx.GetRawData()
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusBadRequest, "Invalid request body")
	}

	var params OrgParams

	err = json.Unmarshal(jsonData, &params)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusBadRequest, "Invalid request body")
	}

	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		ctx.String(http.StatusUnauthorized, "Unauthorized")
	}

	token := authorizationHeader[len("Bearer "):]
	claims, err := DecodeJWT(token)
	if err != nil {
		log.Print(err)
		ctx.String(http.StatusUnauthorized, "Unauthorized")
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to create organization")
	}

	var organizationId string

	err = tx.QueryRow("INSERT INTO organization (name, identifier) VALUES ($1, $2) RETURNING id;", params.Name, params.Identifier).Scan(&organizationId)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to create organization")
	}

	_, err = tx.Exec("INSERT INTO user_organization (organization_id, user_id, level) VALUES ($1, $2, 'owner');", organizationId, claims["sub"])
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to create organization")
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to create organization")
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"organization_id": organizationId,
	})
}
