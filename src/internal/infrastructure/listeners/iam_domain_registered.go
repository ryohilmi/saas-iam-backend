package listeners

import (
	"context"
	"encoding/json"
	"io"
	"iyaem/internal/domain/repositories"
	"iyaem/internal/providers"
	"log"
	"net/http"
	"os"
	"strings"
)

type IamDomainRegisteredHandlers struct {
	organizationRepo repositories.OrganizationRepository
}

func NewIamDomainRegisteredHandlers(
	organizationRepo repositories.OrganizationRepository,
) *IamDomainRegisteredHandlers {
	return &IamDomainRegisteredHandlers{
		organizationRepo: organizationRepo,
	}
}

func (l *IamDomainRegisteredHandlers) GetHandlers() []providers.Callback {
	return []providers.Callback{
		l.AddCallbackUrl,
	}
}

func (l *IamDomainRegisteredHandlers) AddCallbackUrl(ctx context.Context, payload map[string]interface{}) {
	apiToken := providers.GetTokenSingleton().Token

	url := "https://saasiam.us.auth0.com/api/v2/clients/" + os.Getenv("AUTH0_CLIENT_ID")

	req, _ := http.NewRequest("GET", url, nil)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	clientJson := make(map[string]interface{})
	err = json.Unmarshal(bodyBytes, &clientJson)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	IlogoutUrls := clientJson["allowed_logout_urls"].([]interface{})
	Icallbacks := clientJson["callbacks"].([]interface{})

	newLogoutUrl := payload["url"].(string)
	newCallbackUrl := payload["url"].(string) + "/callback"

	logoutUrls := make([]string, 0)
	callbacks := make([]string, 0)

	for _, url := range IlogoutUrls {
		logoutUrls = append(logoutUrls, url.(string))
	}

	for _, url := range Icallbacks {
		callbacks = append(callbacks, url.(string))
	}

	logoutUrls = append(logoutUrls, newLogoutUrl)
	callbacks = append(callbacks, newCallbackUrl)

	reqPayload := strings.NewReader(`{
		"allowed_logout_urls": ["` + strings.Join(logoutUrls, "\",\"") + `"],
		"callbacks": ["` + strings.Join(callbacks, "\",\"") + `"]
	}`)

	log.Printf("Payload: %v", reqPayload)

	req, _ = http.NewRequest("PATCH", url, reqPayload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+apiToken)

	_, err = http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	log.Printf("Added callback url %s", newCallbackUrl)
}
