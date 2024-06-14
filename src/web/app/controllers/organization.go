package controllers

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"iyaem/platform/authenticator"
	auth_token "iyaem/platform/token"
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
	var organizations []organization = make([]organization, 0)

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

func (c *OrganizationController) GetUsersInOrganization(ctx *gin.Context) {
	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	type Params struct {
		OrganizationId string `form:"organization_id" binding:"required"`
	}

	var params Params

	err := ctx.ShouldBindQuery(&params)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusBadRequest, "Organization ID is required")
		return
	}

	organization_id := params.OrganizationId

	token := authorizationHeader[len("Bearer "):]
	claims, err := DecodeJWT(token)
	if err != nil {
		log.Print(err)
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	var level string
	row := c.db.QueryRow(`
		SELECT level FROM user_organization 
		WHERE user_id=$1 AND organization_id=$2;`, claims["sub"], organization_id)

	row.Scan(&level)

	if level != "owner" && level != "manager" {
		ctx.String(http.StatusUnauthorized, "Unauthorized, only owner or manager can view users in organization")
		return
	}

	rows, err := c.db.Query(`
		SELECT uo.id, uo.user_id, u."picture", u."name", u."email", uo."level", uo.created_at as joined_at FROM user_organization uo 
		LEFT JOIN public."user" u ON u.id = uo.user_id 
		WHERE uo.organization_id=$1;`, organization_id)

	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to get users")
		return
	}

	type User struct {
		UserOrgId string `json:"user_org_id"`
		UserId    string `json:"user_id"`
		Picture   string `json:"picture"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Level     string `json:"level"`
		JoinedAt  string `json:"joined_at"`
	}
	var users []User = make([]User, 0)

	for rows.Next() {
		var u User

		err = rows.Scan(&u.UserOrgId, &u.UserId, &u.Picture, &u.Name, &u.Email, &u.Level, &u.JoinedAt)
		if err != nil {
			log.Printf("Error: %v", err)
			ctx.String(http.StatusInternalServerError, "Failed to get users")
			return
		}

		users = append(users, u)
	}

	ctx.JSON(http.StatusOK, users)
}

func (c *OrganizationController) CreateOrganization(ctx *gin.Context) {
	type OrgParams struct {
		Name       string `json:"name" binding:"required"`
		Identifier string `json:"identifier" binding:"required"`
	}
	var params OrgParams

	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusBadRequest, "Invalid request body")
		return
	}

	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	token := authorizationHeader[len("Bearer "):]
	claims, err := DecodeJWT(token)
	if err != nil {
		log.Print(err)
		ctx.String(http.StatusUnauthorized, "Unauthorized")
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to create organization")
		return
	}

	var organizationId string

	err = tx.QueryRow("INSERT INTO organization (name, identifier) VALUES ($1, $2) RETURNING id;", params.Name, params.Identifier).Scan(&organizationId)
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to create organization")
		return
	}

	_, err = tx.Exec("INSERT INTO user_organization (organization_id, user_id, level) VALUES ($1, $2, 'owner');", organizationId, claims["sub"])
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to create organization")
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error: %v", err)
		ctx.String(http.StatusInternalServerError, "Failed to create organization")
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"organization_id": organizationId,
	})
}

func (c *OrganizationController) AddUser(ctx *gin.Context) {
	type OrgParams struct {
		Email          string `json:"email" binding:"required"`
		OrganizationId string `json:"organization_id" binding:"required"`
	}

	var params OrgParams

	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		log.Printf("Error 0101: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	token := authorizationHeader[len("Bearer "):]
	claims, err := DecodeJWT(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 0102: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	var level string

	err = tx.QueryRow("SELECT level FROM user_organization uo left join public.user u on uo.user_id = u.id  WHERE u.email=$1 and uo.organization_id=$2;", claims["email"], params.OrganizationId).Scan(&level)
	if err != nil {
		log.Printf("Error 0103: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	log.Printf("Level: %v", level)

	if level != "owner" && level != "manager" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized, only owner or manager can add user to organization",
		})
		return
	}

	var userId string

	err = tx.QueryRow("SELECT id FROM public.user u WHERE u.email=$1", params.Email).Scan(&userId)
	if err != nil {
		log.Printf("Error 0104: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user, user doesn't exist",
		})
		return
	}

	_, err = tx.Exec("INSERT INTO user_organization (organization_id, user_id, level) VALUES ($1, $2, 'member');", params.OrganizationId, userId)
	if err != nil {
		log.Printf("Error 0105: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 0106: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"organization_id": params.OrganizationId,
		"user_id":         userId,
	})
}

func (c *OrganizationController) CreateUser(ctx *gin.Context) {
	type OrgParams struct {
		Email          string `json:"email" binding:"required"`
		Name           string `json:"name" binding:"required"`
		Password       string `json:"password" binding:"required"`
		OrganizationId string `json:"organization_id" binding:"required"`
	}

	var params OrgParams

	err := ctx.ShouldBindJSON(&params)
	if err != nil {
		log.Printf("Error 0101: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request body",
		})
		return
	}

	authorizationHeader := ctx.Request.Header.Get("Authorization")

	if authorizationHeader == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	token := authorizationHeader[len("Bearer "):]
	claims, err := DecodeJWT(token)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized",
		})
		return
	}

	tx, err := c.db.Begin()
	if err != nil {
		log.Printf("Error 0102: %v", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Failed to create user",
		})
		return
	}

	var level string

	err = tx.QueryRow("SELECT level FROM user_organization uo left join public.user u on uo.user_id = u.id  WHERE u.email=$1 and uo.organization_id=$2;", claims["email"], params.OrganizationId).Scan(&level)
	if err != nil {
		log.Printf("Error 0103: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	log.Printf("Level: %v", level)

	if level != "owner" && level != "manager" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized, only owner or manager can add user to organization",
		})
		return
	}

	apiToken := auth_token.GetTokenSingleton().Token

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

	var userId string

	err = tx.QueryRow("INSERT INTO public.user (picture, email, name) VALUES ($1, $2, $3) RETURNING id;", user.Picture, user.Email, user.Name).Scan(&userId)
	if err != nil {
		log.Printf("Error 0105: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	_, err = tx.Exec("INSERT INTO user_organization (organization_id, user_id, level) VALUES ($1, $2, 'member');", params.OrganizationId, userId)
	if err != nil {
		log.Printf("Error 0106: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Error 0107: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to add user to organization",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
	})
}
