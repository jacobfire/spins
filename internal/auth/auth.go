package auth

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = "supersecret"

// CreateToken to create JWT tokens with claims
func CreateToken(username string) (string, error) {
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,                         // Subject (user identifier)
		"iss": "auth-app",                       // Issuer
		"aud": "player",                         // Audience (user role)
		"exp": time.Now().Add(time.Hour).Unix(), // Expiration time
		"iat": time.Now().Unix(),                // Issued at
	})

	tokenString, err := claims.SignedString([]byte(secretKey))
	if err != nil {
		log.Printf("Token SignedString failed %+v\n", err)

		return "", err
	}

	return tokenString, nil
}

// AuthenticateMiddleware to verify JWT tokens
func AuthenticateMiddleware(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	authToken := strings.Split(tokenString, " ")
	if len(authToken) != 2 {
		fmt.Println("Token malformed")
		c.JSON(http.StatusUnauthorized, gin.H{
			"response": "token malformed",
		})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// Verify the token
	token, err := verifyToken(authToken[1])
	if err != nil {
		fmt.Printf("Token verification failed: %v\\n", err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"response": "token not correct",
		})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	cl, _ := token.Claims.(jwt.MapClaims)
	email, err := cl.GetSubject()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"response": "subject token not correct",
		})
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Set("current_user", email)

	// Continue with the next middleware or route handler
	c.Next()
}

// Function to verify JWT tokens
func verifyToken(tokenString string) (*jwt.Token, error) {
	// Parse the token with the secret key

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})

	// Check for verification errors
	if err != nil {
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Return the verified token
	return token, nil
}
