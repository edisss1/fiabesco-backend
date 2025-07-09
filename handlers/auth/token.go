package auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
)

type Claims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

func GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"id": userID})

	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", err
	}

	return t, nil

}

func VerifyToken(tokenStr string) (*jwt.Token, *Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return nil, nil, err
	}

	return token, claims, nil
}

func PrintDecodedJWT(token string) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		fmt.Println("Invalid token format")
		return
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		fmt.Println("Failed to decode payload:", err)
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		fmt.Println("Failed to unmarshal payload:", err)
		return
	}

	fmt.Println("Decoded JWT payload:")
	for k, v := range payload {
		fmt.Printf("  %s: %v\n", k, v)
	}
}
