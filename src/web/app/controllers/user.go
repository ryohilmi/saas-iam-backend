package controllers

import (
	"database/sql"
	"log"
	"net/http"

	"iyaem/platform/authenticator"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type UserController struct {
	auth *authenticator.Authenticator
	db   *sql.DB
}

func NewUserController(auth *authenticator.Authenticator, db *sql.DB) *UserController {
	return &UserController{auth, db}
}

func (c *UserController) DoesUserExist(ctx *gin.Context) {
	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	type Params struct {
		Email string `form:"email" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Email is required",
		})
		return
	}

	token := authorizationHeader[len("Bearer "):]
	_, err = DecodeJWT(token)
	if err != nil {
		log.Print(err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	var userExists bool
	row := c.db.QueryRow(`
		SELECT EXISTS(
		SELECT 1 FROM public.user 
		WHERE email=$1);`, params.Email)

	row.Scan(&userExists)

	if !userExists {
		log.Printf("Error: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "User does not exist",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User exists",
	})
}

func (c *UserController) AssignRole(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		RoleId         string `json:"role_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Insert user role
	_, err = tx.Exec("INSERT INTO user_role (user_org_id, role_id, tenant_id) VALUES ($1, $2, $3);", params.UserOrgId, params.RoleId, params.TenantId)
	if err != nil {
		log.Printf("Error 1305: %v", err)

		if err.Error() == "pq: duplicate key value violates unique constraint \"user_role_user_org_id_idx\"" {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "User already has this role",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User assigned role",
	})
}

func (c *UserController) RemoveRole(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		RoleId         string `json:"role_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Insert user role
	_, err = tx.Exec("DELETE FROM user_role WHERE user_org_id=$1 AND role_id=$2 AND tenant_id=$3", params.UserOrgId, params.RoleId, params.TenantId)
	if err != nil {
		log.Printf("Error 1305: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove role",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Role removed from the user",
	})
}
func (c *UserController) RemoveGroup(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
		TenantId       string `json:"tenant_id" binding:"required"`
		GroupId        string `json:"group_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove group",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove group",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Insert user role
	_, err = tx.Exec("DELETE FROM user_group WHERE user_org_id=$1 AND group_id=$2 AND tenant_id=$3", params.UserOrgId, params.GroupId, params.TenantId)
	if err != nil {
		log.Printf("Error 1305: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove group",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove group",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Group removed from the user",
	})
}

func (c *UserController) Promote(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Update user level
	_, err = tx.Exec("UPDATE user_organization SET level='manager' WHERE id=$1;", params.UserOrgId)
	if err != nil {
		log.Printf("Error 1305: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	// Increment the member count
	_, err = tx.Exec("UPDATE organization SET manager_count = manager_count + 1 WHERE id=$1;", params.OrganizationId)
	if err != nil {
		log.Printf("Error 3131: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User assigned role",
	})
}

func (c *UserController) Demote(ctx *gin.Context) {
	type Params struct {
		OrganizationId string `json:"organization_id" binding:"required"`
		UserOrgId      string `json:"user_org_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindBodyWith(&params, binding.JSON)
	if err != nil {
		log.Printf("Error 1301: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 1302: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to assign role",
		})
		return
	}

	// Check if user exists in organization
	row := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM user_organization WHERE id=$1 AND organization_id=$2);", params.UserOrgId, params.OrganizationId)
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		log.Printf("Error 1303: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	if !exists {
		log.Printf("Error 1304: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "User does not exist in organization",
		})
		return
	}

	// Update user level
	_, err = tx.Exec("UPDATE user_organization SET level='member' WHERE id=$1;", params.UserOrgId)
	if err != nil {
		log.Printf("Error 1305: %v", err)

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	// Increment the member count
	_, err = tx.Exec("UPDATE organization SET manager_count = manager_count - 1 WHERE id=$1;", params.OrganizationId)
	if err != nil {
		log.Printf("Error 3131: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 1306: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to promote user",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "User assigned role",
	})
}
