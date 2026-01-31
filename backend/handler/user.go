package handler

import (
	"errors"
	"net/http"
	"sol_coffeesys/backend/db"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

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
