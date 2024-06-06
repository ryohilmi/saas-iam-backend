package auth_token

import (
	"io"
	"log"
	"net/http"
	"strings"
)

type TokenSingleton struct {
	Token string
}

var singleInstance *TokenSingleton

func GetTokenSingleton() *TokenSingleton {
	if singleInstance == nil {
		if singleInstance == nil {
			log.Println("Creating token instance now.")

			url := "https://saasiam.us.auth0.com/oauth/token"

			payload := strings.NewReader("{\"client_id\":\"ZJbRnQ3QbVrR8d9zQLXTDB1FQtXTXuaz\",\"client_secret\":\"KQfrH_R_YHuaH33KzksbAk5oxyHQOv_vP7LlRj9ige6qsIwVm5Quqvsdz9MAW4aT\",\"audience\":\"https://saasiam.us.auth0.com/api/v2/\",\"grant_type\":\"client_credentials\"}")

			req, _ := http.NewRequest("POST", url, payload)

			req.Header.Add("content-type", "application/json")

			res, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("Error: %v", err)
			}

			defer res.Body.Close()
			body, _ := io.ReadAll(res.Body)

			singleInstance = &TokenSingleton{
				Token: string(body),
			}
		} else {
			log.Println("Token instance already created.")
		}
	} else {
		log.Println("Token instance already created.")
	}

	return singleInstance
}
