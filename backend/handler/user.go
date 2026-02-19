package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"
	"sol_coffeesys/backend/pkg/respond"
	"sol_coffeesys/backend/pkg/validation"
	"strconv"

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
		return "", fmt.Errorf("パスワードのハッシュ化に失敗しました: %w", err)
	}
	return string(hashed), nil
}

func RegisterUserHandler(q db.Querier) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req RegisterRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}
		// バリデーションチェック
		if err := validation.ValidateRegisterRequest(req.Name, req.Email, req.Password); err != nil {
			switch {
			case errors.Is(err, validation.ErrInvalidName):
				respond.RespondError(c, http.StatusBadRequest, "名前は必須です")
			case errors.Is(err, validation.ErrInvalidEmail):
				respond.RespondError(c, http.StatusBadRequest, "メールアドレスの形式が正しくありません")
			case errors.Is(err, validation.ErrInvalidPassword):
				respond.RespondError(c, http.StatusBadRequest, "パスワードの形式が正しくありません")
			default:
				c.Error(err)
				respond.RespondError(c, http.StatusBadRequest, "入力が不正です")
			}
			return
		}

		hashed, err := HashPassword(req.Password)
		if err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "パスワードのハッシュ化に失敗しました")
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
					respond.RespondError(c, http.StatusBadRequest, "このメールアドレスは既に登録されています")
					return
				}
			}
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
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
			fmt.Printf("Bind Error: %v\n", err)
			respond.RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}

		// バリデーションチェック
		if err := validation.ValidateEmail(req.Email); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "メールアドレスの形式が正しくありません")
			return
		}
		if err := validation.ValidatePassword(req.Password); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "パスワードの形式が正しくありません")
			return
		}

		user, err := q.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil {
			respond.RespondError(c, http.StatusUnauthorized, "メールアドレスまたはパスワードが正しくありません")
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			respond.RespondError(c, http.StatusUnauthorized, "メールアドレスまたはパスワードが正しくありません")
			return
		}

		token, err := tokenGenerator.GenerateToken(user.ID)
		// token, err := auth.GenerateToken(int32(user.ID))
		if err != nil {
			respond.RespondError(c, http.StatusInternalServerError, "トークンの生成に失敗しました")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "ログイン成功",
			"token":   token,
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
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
			respond.RespondError(c, http.StatusBadRequest, "無効なユーザーIDです")
			return
		}
		raw, exists := c.Get("userID")
		if !exists {
			respond.RespondError(c, http.StatusUnauthorized, "認証が必要です")
			return
		}
		adminID := raw.(int64)

		if adminID == userID {
			respond.RespondError(c, http.StatusBadRequest, "自分自身のロールは変更できません")
			return
		}

		var req SetUserRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "リクエストが不正です")
			return
		}

		if err := validation.ValidateRole(req.Role); err != nil {
			respond.RespondError(c, http.StatusBadRequest, "無効なロール")
			return
		}

		user, err := q.UpdateUserRole(c.Request.Context(), db.UpdateUserRoleParams{
			ID:   userID,
			Role: req.Role,
		})
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respond.RespondError(c, http.StatusNotFound, "ユーザーが見つかりません")
				return
			}
			respond.RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
