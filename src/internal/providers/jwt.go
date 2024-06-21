package providers

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

func decodeJWT(token string) (*IamToken, error) {
	claims := &IamToken{}
	var err error
	var tokenInstance *jwt.Token

	if os.Getenv("APP_DEBUG") == "true" {
		tokenInstance, _, err = new(jwt.Parser).ParseUnverified(token, claims)
	} else {
		tokenInstance, err = new(jwt.Parser).ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
			}
			return os.Getenv("JWT_SECRET"), nil
		})
		if !tokenInstance.Valid {
			return nil, fmt.Errorf("invalid token: %")
		}
	}

	if err != nil {
		return nil, err
	}

	return claims, nil
}

type IamToken struct {
	Email      string `json:"email,omitempty"`
	ExpiresAt  int64  `json:"exp,omitempty"`
	IssuedAt   int64  `json:"iat,omitempty"`
	Name       string `json:"name,omitempty"`
	PictureUrl string `json:"picture,omitempty"`
	UserId     string `json:"sub,omitempty"`
}

// Valid Validates time based claims "exp, iat, nbf".
// There is no accounting for clock skew.
// As well, if any of the above claims are not in the token, it will still
// be considered a valid claim.
func (c IamToken) Valid() error {
	now := time.Now().Unix()

	if !c.verifyExp(c.ExpiresAt, now, false) {
		delta := time.Unix(now, 0).Sub(time.Unix(c.ExpiresAt, 0))
		return fmt.Errorf("token is expired by %v", delta)
	}

	if !c.verifyIat(c.IssuedAt, now, false) {
		return fmt.Errorf("token used before issued")
	}

	return nil
}

func (c IamToken) verifyExp(exp int64, now int64, required bool) bool {
	if exp == 0 {
		return !required
	}
	return now <= exp
}

func (c IamToken) verifyIat(iat int64, now int64, required bool) bool {
	if iat == 0 {
		return !required
	}
	return now >= iat
}
