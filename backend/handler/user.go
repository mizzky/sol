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

func RegisterHandler(q *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが正しくありません"})
			return
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

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
					c.JSON(http.StatusBadRequest, gin.H{"error": "このメールアドレスは既に登録されています"})
					return
				}
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "予期せぬエラー"})
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

func LoginHandler(q *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			fmt.Printf("Bind Error: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが正しくありません"})
			return
		}

		user, err := q.GetUserByEmail(c.Request.Context(), req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
			return
		}
		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "メールアドレスまたはパスワードが正しくありません"})
			return
		}

		token, err := auth.GenerateToken(int32(user.ID))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "トークンの生成に失敗しました"})
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
