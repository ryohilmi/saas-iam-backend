package controller

import (
	"database/sql"
	"encoding/json"
	"io"
	"iyaem/internal/app/commands"
	"iyaem/internal/app/queries"
	"iyaem/internal/providers"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct {
	db *sql.DB

	createOrganizationCommand *commands.CreateOrganizationCommand
	addUserCommand            *commands.AddOrganizationUserCommand
	createUserCommand         *commands.CreateUserCommand
	addRoleCommand            *commands.AddRoleToMemberCommand

	organizationQuery queries.OrganizationQuery
}

func NewOrganizationController(
	db *sql.DB,
	createOrganizationCommand *commands.CreateOrganizationCommand,
	addUser *commands.AddOrganizationUserCommand,
	createUser *commands.CreateUserCommand,
	addRoleCommand *commands.AddRoleToMemberCommand,
	organizationQuery queries.OrganizationQuery,
) *OrganizationController {
	return &OrganizationController{
		db,
		createOrganizationCommand,
		addUser,
		createUser,
		addRoleCommand,
		organizationQuery,
	}
}

func (c *OrganizationController) Statistics(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `form:"organization_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusBadRequest, "Organization ID is required")
	}

	type Statistics struct {
		MemberCount  int `json:"member_count"`
		TenantCount  int `json:"tenant_count"`
		ManagerCount int `json:"manager_count"`
	}

	var statistics Statistics

	err = c.db.QueryRow(`
		SELECT member_count, manager_count, tenant_count 
		FROM organization 
		WHERE id=$1;`, params.OrganizationId).Scan(&statistics.MemberCount, &statistics.ManagerCount, &statistics.TenantCount)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get statistics",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": statistics,
	})
}

func (c *OrganizationController) FindById(ctx *gin.Context) {
	organizationId := ctx.Param("id")

	organization, err := c.organizationQuery.FindById(ctx, organizationId)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, organization)
}

func (c *OrganizationController) GetAffiliatedOrganizations(ctx *gin.Context) {
	token := GetBearerToken(ctx)

	claims, err := DecodeJWT(token)
	if err != nil {
		log.Print(err)
		ctx.Error(err)
		return
	}

	organizations, err := c.organizationQuery.AllAffilatedOrganizations(ctx, claims["sub"].(string))
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, organizations)
}

func (c *OrganizationController) GetUsers(ctx *gin.Context) {
	var params struct {
		OrganizationId string `form:"organization_id" binding:"required"`
	}
	if err := ctx.ShouldBind(&params); err != nil {
		ctx.Error(err)
		return
	}

	users, err := c.organizationQuery.UsersInOrganization(ctx, params.OrganizationId)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *OrganizationController) GetRecentUsers(ctx *gin.Context) {
	var params struct {
		OrganizationId string `form:"organization_id" binding:"required"`
	}
	if err := ctx.ShouldBind(&params); err != nil {
		ctx.Error(err)
		return
	}

	users, err := c.organizationQuery.RecentUsersInOrganization(ctx, params.OrganizationId)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *OrganizationController) CreateOrganization(ctx *gin.Context) {
	var params struct {
		Name       string `json:"name" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
	}

	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	token := GetBearerToken(ctx)
	claims, err := DecodeJWT(token)
	if err != nil {
		log.Print(err)
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	req := commands.CreateOrganizationRequest{
		Name:       params.Name,
		Identifier: params.Identifier,
		UserId:     claims["sub"].(string),
	}
	orgId, err := c.createOrganizationCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "success",
		"data":    orgId,
	})
}

func (c *OrganizationController) AddUser(ctx *gin.Context) {
	var params struct {
		Email          string `json:"email" binding:"required"`
		OrganizationId string `json:"organization_id" binding:"required"`
	}

	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		log.Printf("Error 0101: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	req := commands.AddOrganizationUserRequest{
		Email:          params.Email,
		OrganizationId: params.OrganizationId,
	}
	membershipId, err := c.addUserCommand.Execute(ctx, req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "success",
		"data":    membershipId,
	})
}

func (c *OrganizationController) CreateUser(ctx *gin.Context) {
	var params struct {
		Email          string `json:"email" binding:"required"`
		Name           string `json:"name" binding:"required"`
		Password       string `json:"password" binding:"required"`
		OrganizationId string `json:"organization_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&params); err != nil {
		log.Printf("Error 0101: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	apiToken := providers.GetTokenSingleton().Token

	url := "https://saasiam.us.auth0.com/api/v2/users"

	payload := strings.NewReader("{\"email\":\"" + params.Email + "\",\"nickname\":\"" + params.Email + "\",\"name\":\"" + params.Name + "\",\"password\":\"" + params.Password + "\",\"connection\":\"Username-Password-Authentication\"}")

	log.Printf("Payload: %v", payload)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user",
		})
		return
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Printf("Body: %v", bodyString)

	if res.StatusCode != 201 {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create user",
		})
		return
	}

	defer res.Body.Close()

	type User struct {
		IdpId   string `json:"user_id"`
		Picture string `json:"picture"`
		Email   string `json:"email"`
		Name    string `json:"name"`
	}

	var user User
	err = json.Unmarshal(bodyBytes, &user)
	if err != nil {
		log.Printf("Error 0104: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to insert user",
		})
		return
	}

	createReq := commands.CreateUserRequest{
		Email:   user.Email,
		Name:    user.Name,
		Picture: user.Picture,
		IdpId:   user.IdpId,
	}
	_, err = c.createUserCommand.Execute(ctx, createReq)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	addReq := commands.AddOrganizationUserRequest{
		Email:          user.Email,
		OrganizationId: params.OrganizationId,
	}
	membershipId, err := c.addUserCommand.Execute(ctx, addReq)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "success",
		"data":    membershipId,
	})
}

func (c *OrganizationController) RemoveUser(ctx *gin.Context) {
	type OrgParams struct {
		UserOrgId      string `json:"user_org_id" binding:"required"`
		OrganizationId string `json:"organization_id" binding:"required"`
	}

	var params OrgParams

	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		log.Printf("Error 0101: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   true,
			"message": "Invalid request body",
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 0102: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error":   true,
			"message": "Failed to remove user from organization",
		})
		return
	}

	var level string
	row := c.db.QueryRow(`
		SELECT level FROM user_organization
		WHERE id=$1;`, params.UserOrgId)

	row.Scan(&level)

	_, err = tx.Exec("DELETE FROM user_role WHERE user_org_id=$1;", params.UserOrgId)
	if err != nil {
		log.Printf("Error 0105: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to remove user from organization",
		})
		return
	}

	_, err = tx.Exec("DELETE FROM user_group WHERE user_org_id=$1;", params.UserOrgId)
	if err != nil {
		log.Printf("Error 0105: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to remove user from organization",
		})
		return
	}

	_, err = tx.Exec("DELETE FROM user_organization WHERE id=$1;", params.UserOrgId)
	if err != nil {
		log.Printf("Error 0105: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to remove user from organization",
		})
		return
	}

	// decrement the member count
	_, err = tx.Exec("UPDATE organization SET member_count = member_count - 1 WHERE id=$1;", params.OrganizationId)
	if err != nil {
		log.Printf("Error 3131: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to add user to organization",
		})
		return
	}

	// decrement the manager count
	if level == "manager" {
		_, err = tx.Exec("UPDATE organization SET manager_count = manager_count - 1 WHERE id=$1;", params.OrganizationId)
		if err != nil {
			log.Printf("Error 3131: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":   true,
				"message": "Failed to add user to organization",
			})
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 0106: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":   true,
			"message": "Failed to add user to organization",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User removed successfully",
	})
}
