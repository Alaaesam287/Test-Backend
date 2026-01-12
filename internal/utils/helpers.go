package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Password hashing
func HashPassword(password string) (string, error) {
	if len([]byte(password)) > 72 {
		return "", errors.New("password too long")
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// JWT generation
func GenerateJWT(userID int64, role string, storeID *int64, secret string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(duration).Unix(),
	}

	if storeID != nil {
		claims["store_id"] = *storeID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// JWT parsing
func ParseJWT(tokenStr, secret string) (*jwt.Token, jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, nil, errors.New("invalid token")
	}

	return token, claims, err
}

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Password policy check will be implemented later
func CheckPasswordPolicy(password string) error {
	return nil
}

func HashAttributes(attrs []models.VariantAttributeInput) string {
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].AttributeID < attrs[j].AttributeID
	})

	var parts []string
	for _, a := range attrs {
		parts = append(parts, fmt.Sprintf("%d:%s", a.AttributeID, a.Value))
	}

	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}


func InterfaceSlice[T any](s []T) []interface{} {
	out := make([]interface{}, len(s))
	for i := range s {
		out[i] = s[i]
	}
	return out
}

func NullStringToPtr(ns sql.NullString) *string {
    if ns.Valid {
        return &ns.String
    }
    return nil
}