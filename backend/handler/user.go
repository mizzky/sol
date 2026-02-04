package handler

import (
	"errors"
	"fmt"
	"net/http"
	"sol_coffeesys/backend/auth"
	"sol_coffeesys/backend/db"

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
			RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}

		if req.Name == "" {
			RespondError(c, http.StatusBadRequest, "名前は必須です")
			return
		}

		hashed, err := HashPassword(req.Password)

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
					RespondError(c, http.StatusBadRequest, "このメールアドレスは既に登録されています")
					return
				}
			}
			RespondError(c, http.StatusInternalServerError, "予期せぬエラーが発生しました")
			return
		}

		// 登録成功
		c.JSON(http.StatusCreated, user)
	}
}

// ＋＋　ログイン機能　＋＋
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func LoginUserHandler(q db.Querier, tokenGenerator auth.TokenGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("Bind Error: %v\n", err)
			RespondError(c, http.StatusBadRequest, "リクエスト形式が正しくありません")
			return
		}

		user, err := q.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil {
			RespondError(c, http.StatusUnauthorized, "メールアドレスまたはパスワードが正しくありません")
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			RespondError(c, http.StatusUnauthorized, "メールアドレスまたはパスワードが正しくありません")
			return
		}

		token, err := tokenGenerator.GenerateToken(user.ID)
		// token, err := auth.GenerateToken(int32(user.ID))
		if err != nil {
			RespondError(c, http.StatusInternalServerError, "トークンの生成に失敗しました")
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
