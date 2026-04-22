package apperror

import (
	"errors"
	"net/http"
	"testing"
)

func TestToHTTP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantMsg    string
	}{
		{
			name:       "nil errorは500共通メッセージ",
			err:        nil,
			wantStatus: http.StatusInternalServerError,
			wantMsg:    InternalServerMessageCommon,
		},
		{
			name:       "Validation: Message優先",
			err:        NewValidationError("email", "a@example.com", "format", "custom validation"),
			wantStatus: http.StatusBadRequest,
			wantMsg:    "custom validation",
		},
		{
			name:       "Validation: Field map email",
			err:        NewValidationError("email", "a@example.com", "format", ""),
			wantStatus: http.StatusBadRequest,
			wantMsg:    ValidationMessageEmail,
		},
		{
			name:       "Validation: 未知fieldは generic",
			err:        NewValidationError("unkown", "x", "rule", ""),
			wantStatus: http.StatusBadRequest,
			wantMsg:    ValidationMessageGeneric,
		},
		{
			name:       "Conflict: Message優先",
			err:        NewConflictError("sku", "ABC", "custom conflict"),
			wantStatus: http.StatusConflict,
			wantMsg:    "custom conflict",
		},
		{
			name:       "Conflict: field map qty",
			err:        NewConflictError("qty", "1", ""),
			wantStatus: http.StatusConflict,
			wantMsg:    ConflictMessageQty,
		},
		{
			name:       "Conflict: 未知fieldはfallback",
			err:        NewConflictError("unkown", "x", ""),
			wantStatus: http.StatusConflict,
			wantMsg:    ConflictMessageGeneric,
		},
		{
			name:       "NotFound: resource map cart_item",
			err:        NewNotFoundError("cart_item", 1, ""),
			wantStatus: http.StatusNotFound,
			wantMsg:    NotFoundMessageCartItem,
		},
		{
			name:       "NotFound: 未知resourceはfallback",
			err:        NewNotFoundError("unknown", 1, ""),
			wantStatus: http.StatusNotFound,
			wantMsg:    NotFoundMessageGeneric,
		},
		{
			name:       "Unauthorized: Message優先",
			err:        NewUnauthorizedError("invalid_credentials", "custom unauthorized"),
			wantStatus: http.StatusUnauthorized,
			wantMsg:    "custom unauthorized",
		},
		{
			name:       "Unauthorized: Message空はfallback",
			err:        NewUnauthorizedError("token_not_found", ""),
			wantStatus: http.StatusUnauthorized,
			wantMsg:    UnauthorizedMessageGeneric,
		},
		{
			name:       "Forbidden: Message空はfallback",
			err:        NewForbiddenError("admin", "user", ""),
			wantStatus: http.StatusForbidden,
			wantMsg:    ForbiddenMessageGeneric,
		},
		{
			name:       "BusinessLogic: Message空はfallback",
			err:        NewBusinessLogicError(""),
			wantStatus: http.StatusBadRequest,
			wantMsg:    BusinessLogicMessageGeneric,
		},
		{
			name:       "Internal: Message優先",
			err:        NewInternalError("CreateUser", errors.New("db"), "custom internal error"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    "custom internal error",
		},
		{
			name:       "Internal: Message空は common fallback",
			err:        NewInternalError("CreateUser", errors.New("db"), ""),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    InternalServerMessageCommon,
		},
		{
			name:       "未知errorは500 common",
			err:        errors.New("unknown"),
			wantStatus: http.StatusInternalServerError,
			wantMsg:    InternalServerMessageCommon,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotStatus, gotMsg := ToHTTP(tt.err)
			if gotStatus != tt.wantStatus {
				t.Fatalf("status mismatch: got=%d want=%d", gotStatus, tt.wantStatus)
			}
			if gotMsg != tt.wantMsg {
				t.Fatalf("message mismatch: got=%q want=%q", gotMsg, tt.wantMsg)
			}
		})
	}
}
