package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/apperror"
	"sol_coffeesys/backend/pkg/respond"
	"sol_coffeesys/backend/pkg/validation"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// ＋＋ユーザー登録機能＋＋
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var BcryptGenerateFromPassword = bcrypt.GenerateFromPassword

func HashPassword(password string) (string, error) {
	hashed, err := BcryptGenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", apperror.NewInternalError("HashPassword", err, apperror.InternalServerMessagePassword)
	}
	return string(hashed), nil
}

func RegisterUserHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req RegisterRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			respond.RespondWithError(c, apperror.NewValidationError("request", nil, "bind", apperror.ValidationMessageRequest))
			return
		}
		// バリデーションチェック
		if err := validation.ValidateRegisterRequest(req.Name, req.Email, req.Password); err != nil {
			switch {
			case errors.Is(err, validation.ErrInvalidName):
				respond.RespondWithError(c, apperror.NewValidationError("name", nil, "", ""))
			case errors.Is(err, validation.ErrInvalidEmail):
				respond.RespondWithError(c, apperror.NewValidationError("email", nil, "", ""))
			case errors.Is(err, validation.ErrInvalidPassword):
				respond.RespondWithError(c, apperror.NewValidationError("password", nil, "", ""))
			default:
				c.Error(err)
				respond.RespondWithError(c, apperror.NewValidationError("request", nil, "", ""))
			}
			return
		}

		hashed, err := HashPassword(req.Password)
		if err != nil {
			c.Error(err)
			respond.RespondWithError(c, err)
			return
		}

		user, err := q.CreateUser(c.Request.Context(), db.CreateUserParams{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: string(hashed),
			Role:         "member",
		})
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) {
				if pqErr.Code == "23505" {
					respond.RespondWithError(c, apperror.NewValidationError("email", req.Email, "", apperror.ValidationMessageConflictedEmail))
					return
				}
			}
			respond.RespondWithError(c, apperror.NewInternalError("CreateUser", err, apperror.InternalServerMessageCommon))
			return
		}

		// 登録成功		migrate -path db/migrations -database "postgres://user:password@db:5432/coffeesys_db?sslmode=disable" up
		c.JSON(http.StatusCreated, user)
	}
}

// ＋＋ログイン機能＋＋
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func LoginUserHandler(q db.Querier, tokenGenerator auth.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.Error(err)
			respond.RespondWithError(c, apperror.NewValidationError("request", nil, "bind", apperror.ValidationMessageRequest))
			return
		}

		// バリデーションチェック
		if err := validation.ValidateEmail(req.Email); err != nil {
			respond.RespondWithError(c, apperror.NewValidationError("email", nil, "", ""))
			return
		}
		if err := validation.ValidatePassword(req.Password); err != nil {
			respond.RespondWithError(c, apperror.NewValidationError("password", nil, "", ""))
			return
		}

		user, err := q.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageEmailOrPassword))
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageEmailOrPassword))
			return
		}

		token, err := tokenGenerator.GenerateToken(user.ID)
		// token, err := auth.GenerateToken(int32(user.ID))
		if err != nil {
			respond.RespondWithError(c, apperror.NewInternalError("GenerateToken", err, apperror.InternalServerMessageGenToken))
			return
		}

		refreshToken, _, expiresAt, err := GenerateRefreshToken(c.Request.Context(), q, user.ID)
		if err != nil {
			respond.RespondWithError(c, apperror.NewInternalError("GenerateRefreshToken", err, apperror.InternalServerMessageRefresh))
			return
		}

		//  Cookieをセット
		accessCookie := &http.Cookie{
			Name:     "access_token",
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			Expires:  time.Now().Add(15 * time.Minute),
			MaxAge:   15 * 60,
			SameSite: http.SameSiteLaxMode,
		}

		refreshCookie := &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			HttpOnly: true,
			Path:     "/api/refresh",
			Expires:  expiresAt,
			MaxAge:   60 * 60 * 24 * 14,
			SameSite: http.SameSiteStrictMode,
		}

		if gin.Mode() == gin.ReleaseMode {
			accessCookie.Secure = true
			refreshCookie.Secure = true
		}

		http.SetCookie(c.Writer, accessCookie)
		http.SetCookie(c.Writer, refreshCookie)

		c.JSON(http.StatusOK, gin.H{
			"message": "ログイン成功",
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
			},
		})
	}
}

// ＋＋権限変更機能＋＋
type SetUserRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

func SetUserRoleHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		userID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			respond.RespondWithError(c, apperror.NewValidationError("id", nil, "", ""))
			return
		}
		raw, exists := c.Get("userID")
		if !exists {
			respond.RespondWithError(c, apperror.NewUnauthorizedError("", apperror.UnauthorizedMessageAuth))
			return
		}
		adminID := raw.(int64)

		if adminID == userID {
			respond.RespondWithError(c, apperror.NewBusinessLogicError(apperror.BusinessLogicMessageRole))
			return
		}

		var req SetUserRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respond.RespondWithError(c, apperror.NewValidationError("request", nil, "bind", apperror.ValidationMessageRequest))
			return
		}

		if err := validation.ValidateRole(req.Role); err != nil {
			respond.RespondWithError(c, apperror.NewValidationError("role", nil, "", ""))
			return
		}

		user, err := q.UpdateUserRole(c.Request.Context(), db.UpdateUserRoleParams{
			ID:   userID,
			Role: req.Role,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondWithError(c, apperror.NewNotFoundError("user", userID, ""))
				return
			}
			respond.RespondWithError(c, apperror.NewInternalError("UpdateUserRole", err, apperror.InternalServerMessageCommon))
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
