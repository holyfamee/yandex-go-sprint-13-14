package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type SignInRequest struct {
	Password string `json:"password"`
}

type SignInResponse struct {
	Token string `json:"token,omitempty"`
	Error string `json:"error,omitempty"`
}

func getPasswordHash(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// generateJWT генерирует JWT токен с хэшем пароля в claims
func generateJWT(password string) (string, error) {
	claims := jwt.MapClaims{
		"password_hash": getPasswordHash(password),
		"exp":           time.Now().Add(8 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(password))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// validateJWT проверяет валидность JWT токена
func validateJWT(tokenString string, password string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(password), nil
	})

	if err != nil {
		return false
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if passwordHash, ok := claims["password_hash"].(string); ok {
			return passwordHash == getPasswordHash(password)
		}
	}

	return false
}

// signInHandler обрабатывает POST-запросы для аутентификации
func signInHandler(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Ошибка чтения тела запроса: %v", err)
		writeJSON(w, SignInResponse{Error: "Ошибка чтения запроса"}, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("Ошибка десериализации JSON: %v", err)
		writeJSON(w, SignInResponse{Error: "Ошибка десериализации JSON"}, http.StatusBadRequest)
		return
	}

	envPassword := os.Getenv("TODO_PASSWORD")
	if envPassword == "" {
		writeJSON(w, SignInResponse{Error: "Аутентификация не настроена"}, http.StatusInternalServerError)
		return
	}

	if req.Password != envPassword {
		writeJSON(w, SignInResponse{Error: "Неверный пароль"}, http.StatusUnauthorized)
		return
	}

	token, err := generateJWT(req.Password)
	if err != nil {
		log.Printf("Ошибка генерации JWT токена: %v", err)
		writeJSON(w, SignInResponse{Error: "Ошибка генерации токена"}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, SignInResponse{Token: token}, http.StatusOK)
}

// auth middleware для проверки аутентификации
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) > 0 {
			var jwtToken string

			cookie, err := r.Cookie("token")
			if err == nil {
				jwtToken = cookie.Value
			}

			valid := validateJWT(jwtToken, pass)

			if !valid {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
		}

		next(w, r)
	})
}
