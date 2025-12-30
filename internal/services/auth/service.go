package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Secure-Website-Builder/Backend/internal/models"
	"github.com/Secure-Website-Builder/Backend/internal/types"
	"github.com/Secure-Website-Builder/Backend/internal/utils"
)

type Service struct {
	queries   *models.Queries
	jwtSecret string
}

func New(queries *models.Queries, jwtSecret string) *Service {
	return &Service{
		queries:   queries,
		jwtSecret: jwtSecret,
	}
}

/* ================= REGISTER ================= */

func (s *Service) Register(
	ctx context.Context,
	name, email, password, role string,
	storeID *int64,
	phone *string,
	address *types.Address,
) (string, error) {
	
	err := utils.CheckPasswordPolicy(password)
	if err != nil {
		return "", err
	}

	hashed, err := utils.HashPassword(password)
	if err != nil {
		return "", err
	}

	addr := types.NullableAddress{Valid: false}
	if address != nil {
		addr = types.NullableAddress{
			Addr:  address,
			Valid: true,
		}
	}

	phoneSQL := sql.NullString{}
	if phone != nil {
		phoneSQL = sql.NullString{
			String: *phone,
			Valid:  true,
		}
	}

	var userID int64

	switch role {

	case "store_owner":
		user, err := s.queries.CreateStoreOwner(ctx, models.CreateStoreOwnerParams{
			Name:         name,
			Email:        email,
			PasswordHash: hashed,
			Phone:       phoneSQL,
			Address:    addr,
		})
		if err != nil {
			return "", err
		}
		userID = user.StoreOwnerID

	case "customer":
		if storeID == nil {
			return "", errors.New("store_id is required")
		}
		
		user, err := s.queries.CreateCustomer(ctx, models.CreateCustomerParams{
			StoreID:      *storeID,
			Name:         name,
			Email:        email,
			PasswordHash: hashed,
			Phone:       phoneSQL,
			Address:    addr,
		})
		if err != nil {
			return "", err
		}
		userID = user.CustomerID

	default:
		return "", errors.New("invalid role")
	}

	return utils.GenerateJWT(
		userID,
		role,
		storeID,
		s.jwtSecret,
		24*time.Hour,
	)
}

/* ================= LOGIN ================= */

func (s *Service) Login(
	ctx context.Context,
	email, password, role string,
	storeID *int64,
) (string, error) {

	var (
		userID int64
		hashed string
	)

	switch role {

	case "store_owner":
		user, err := s.queries.GetStoreOwnerByEmail(ctx, email)
		if err != nil {
			return "", errors.New("invalid credentials")
		}
		userID = user.StoreOwnerID
		hashed = user.PasswordHash

	case "customer":
		if storeID == nil {
			return "", errors.New("store_id is required")
		}

		user, err := s.queries.GetCustomerByEmail(ctx, models.GetCustomerByEmailParams{
			Email:   email,
			StoreID: *storeID,
		})
		if err != nil {
			return "", errors.New("invalid credentials")
		}
		userID = user.CustomerID
		hashed = user.PasswordHash

	default:
		return "", errors.New("invalid role")
	}

	if !utils.CheckPasswordHash(password, hashed) {
		return "", errors.New("invalid credentials")
	}

	return utils.GenerateJWT(
		userID,
		role,
		storeID,
		s.jwtSecret,
		24*time.Hour,
	)
}

func (s *Service) AdminLogin(
	ctx context.Context,
	email, password string,
) (string, error) {

	admin, err := s.queries.GetAdminByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, admin.PasswordHash) {
		return "", errors.New("invalid credentials")
	}

	return utils.GenerateJWT(
		admin.AdminID,
		"admin",
		nil, // storeID is ALWAYS nil for admin
		s.jwtSecret,
		24*time.Hour,
	)
}
