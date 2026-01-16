package utils

import (
	"crypto/rand"
	"crypto/sha1" // NEW: For HIBP check
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io" // NEW: For HIBP API response
	"net/http" // NEW: For HIBP API
	"sort"
	"strings"
	"time"
	"regexp" // NEW: For password pattern checks

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

// ---------------------- NEW: Password Policy with MFA ----------------------

func CheckPasswordPolicy(password string, role string) (mfaRequired bool, err error) {
	switch role {
	case "customer":
		err = validateCustomerPassword(password)
		mfaRequired = false // Customers do NOT require MFA
	case "store_owner":
		err = validateBusinessOwnerPassword(password)
		mfaRequired = true // Business Owners require MFA
	default:
		err = errors.New("unknown user role")
		mfaRequired = false
	}
	return
}

// Customer password rules: min 8 chars, letters + numbers
func validateCustomerPassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !hasLetter(password) || !hasNumber(password) {
		return errors.New("password must include letters and numbers")
	}
	return nil
}

// Business Owner password rules: min 12 chars, letters + numbers + symbols, HIBP check
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
		return errors.New("failed to check password against breached database")
	}
	if pwned {
		return errors.New("password has been found in a data breach, please choose another one")
	}

	return nil
}

// Helper functions for password patterns
func hasLetter(s string) bool {
	return regexp.MustCompile(`[A-Za-z]`).MatchString(s)
}
func hasNumber(s string) bool {
	return regexp.MustCompile(`[0-9]`).MatchString(s)
}
func hasSymbol(s string) bool {
	return regexp.MustCompile(`[^A-Za-z0-9]`).MatchString(s)
}

// HIBP k-anonymity API check
func IsPasswordPwned(password string) (bool, error) {
	hash := sha1.Sum([]byte(password))
	hashHex := strings.ToUpper(hex.EncodeToString(hash[:]))

	prefix := hashHex[:5]
	suffix := hashHex[5:]

	resp, err := http.Get("https://api.pwnedpasswords.com/range/" + prefix)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
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

// ---------------------- END OF NEW CODE PASSWORD POLICY FUNCTIONS ----------------------

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
