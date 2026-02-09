package utils

import (
	"crypto/rand"
	// #nosec G505
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Password Hashing

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

func GenerateJWT(
	userID int64,
	role string,
	storeID *int64,
	secret string,
	duration time.Duration,
) (string, error) {

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

	return token, claims, nil
}


func GenerateRefreshToken() (string, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// Password Policy 

func CheckPasswordPolicy(password string, role string) (bool, error) {
	switch role {
	case "customer":
		return false, validateCustomerPassword(password)

	case "store_owner":
		return true, validateBusinessOwnerPassword(password)

	default:
		return false, errors.New("unknown user role")
	}
}

// Customer rules
func validateCustomerPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !hasLetter(password) || !hasNumber(password) {
		return errors.New("password must include letters and numbers")
	}
	return nil
}

// Business Owner rules
func validateBusinessOwnerPassword(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}
	if !hasLetter(password) || !hasNumber(password) || !hasSymbol(password) {
		return errors.New("password must include letters, numbers, and symbols")
	}

	// HIBP check 
	pwned, err := IsPasswordPwned(password)
	if err != nil {
		// TODO: log the error and skip HIBP check
	} else if pwned {
		return errors.New("password has been found in a data breach, please choose another one")
	}

	return nil
}


func hasLetter(s string) bool {
	return regexp.MustCompile(`[A-Za-z]`).MatchString(s)
}

func hasNumber(s string) bool {
	return regexp.MustCompile(`[0-9]`).MatchString(s)
}

func hasSymbol(s string) bool {
	return regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(s)
}

// HIBP Check 

func IsPasswordPwned(password string) (bool, error) {
	// #nosec G401
	hash := sha1.Sum([]byte(password))
	hashHex := strings.ToUpper(hex.EncodeToString(hash[:]))

	prefix := hashHex[:5]
	suffix := hashHex[5:]

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest(
		http.MethodGet,
		"https://api.pwnedpasswords.com/range/"+prefix,
		nil,
	)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", "password-checker/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HIBP API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		if parts[0] == suffix {
			return true, nil
		}
	}

	return false, nil
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
