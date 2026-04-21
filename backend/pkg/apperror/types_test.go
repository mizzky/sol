package apperror

import (
	"errors"
	"testing"
)

func TestValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		field     string
		value     any
		rule      string
		message   string
		wantType  string
		wantValue any
		wantError string
	}{
		{
			name:      "異常系: emailフィールドの値がマスクされる",
			field:     "email",
			value:     "user@example.com",
			rule:      "format",
			message:   ValidationMessageEmail,
			wantType:  "string",
			wantValue: "u****@example.com",
			wantError: ValidationMessageEmail,
		},
		{
			name:      "異常系: emailローカルパートが1文字でもマスクされる",
			field:     "email",
			value:     "a@example.com",
			rule:      "format",
			message:   ValidationMessageEmail,
			wantType:  "string",
			wantValue: "a****@example.com",
			wantError: ValidationMessageEmail,
		},
		{
			name:      "異常系: email以外のフィールドは値をそのまま保持する",
			field:     "price",
			value:     0,
			rule:      "positive",
			message:   ValidationMessagePrice,
			wantType:  "int",
			wantValue: 0,
			wantError: ValidationMessagePrice,
		},
		{
			name:      "異常系: メール重複はValidationErrorとして扱いマスクされる",
			field:     "email",
			value:     "dup@example.com",
			rule:      "unique",
			message:   ValidationMessageConflictedEmail,
			wantType:  "string",
			wantValue: "d****@example.com",
			wantError: ValidationMessageConflictedEmail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := NewValidationError(tt.field, tt.value, tt.rule, tt.message)
			if err == nil {
				t.Fatal("NewValidationError() returned nil")
			}

			if err.Field != tt.field {
				t.Fatalf("Field = %q, want %q", err.Field, tt.field)
			}
			if err.Value != tt.wantValue {
				t.Fatalf("Value = %#v, want %#v", err.Value, tt.wantValue)
			}
			if err.ValueType != tt.wantType {
				t.Fatalf("ValueType = %q, want %q", err.ValueType, tt.wantType)
			}
			if err.Rule != tt.rule {
				t.Fatalf("Rule = %q, want %q", err.Rule, tt.rule)
			}
			if err.Message != tt.message {
				t.Fatalf("Message = %q, want %q", err.Message, tt.message)
			}
			if got := err.Error(); got != tt.wantError {
				t.Fatalf("Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestNewConflictError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		field     string
		value     string
		message   string
		wantError string
	}{
		{
			name:      "異常系：SKU重複は値をそのまま保持する",
			field:     "sku",
			value:     "SKU-001",
			message:   ConflictMessageSku,
			wantError: ConflictMessageSku,
		},
		{
			name:      "異常系：在庫不足は値をそのまま保持する",
			field:     "stock",
			value:     "requested:5, available:2",
			message:   ConflictMessageQty,
			wantError: ConflictMessageQty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := NewConflictError(tt.field, tt.value, tt.message)
			if err == nil {
				t.Fatal("NewConflictError() returned nil")
			}
			if err.Field != tt.field {
				t.Fatalf("Field = %q, want %q", err.Field, tt.field)
			}
			if err.Value != tt.value {
				t.Fatalf("Value = %q, want %q", err.Value, tt.value)
			}
			if err.Message != tt.message {
				t.Fatalf("Message = %q, want %q", err.Message, tt.message)
			}
			if got := err.Error(); got != tt.wantError {
				t.Fatalf("Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestNewInternalError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation string
		cause     error
		message   string
	}{
		{
			name:      "異常系: operationとcauseを保持しUnwrapで取り出せる",
			operation: "CreateUser",
			cause:     errors.New("db timeout"),
			message:   InternalServerMessageCommon,
		},
		{
			name:      "異常系: errors.Isでcauseを辿れる",
			operation: "GenerateToken",
			cause:     errors.New("key not found"),
			message:   InternalServerMessageCommon,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := NewInternalError(tt.operation, tt.cause, tt.message)
			if err == nil {
				t.Fatal("NewInternalError() returned nil")
			}
			if err.Operation != tt.operation {
				t.Fatalf("Operation = %q, want %q", err.Operation, tt.operation)
			}
			if err.Cause != tt.cause {
				t.Fatalf("Cause = %v, want %v", err.Cause, tt.cause)
			}
			if err.Message != tt.message {
				t.Fatalf("Message = %q, want %q", err.Message, tt.message)
			}
			if got := err.Error(); got != tt.message {
				t.Fatalf("Error() = %q, want %q", got, tt.message)
			}

			if errors.Unwrap(err) != tt.cause {
				t.Fatalf("errors.Unwrap() = %v, want %v", errors.Unwrap(err), tt.cause)
			}
			if !errors.Is(err, tt.cause) {
				t.Fatal("errors.Is() = false, want true")
			}
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		resource  string
		id        any
		message   string
		wantError string
	}{
		{
			name:      "異常系: 商品IDで見つからない場合",
			resource:  "product",
			id:        int64(1),
			message:   NotFoundMessageProduct,
			wantError: NotFoundMessageProduct,
		},
		{
			name:      "異常系: ユーザーIDで見つからない場合",
			resource:  "user",
			id:        int64(42),
			message:   NotFoundMessageUser,
			wantError: NotFoundMessageUser,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := NewNotFoundError(tt.resource, tt.id, tt.message)
			if err == nil {
				t.Fatal("NewNotFoundError() returned nil")
			}
			if err.Resource != tt.resource {
				t.Fatalf("Resource = %q, want %q", err.Resource, tt.resource)
			}
			if err.ID != tt.id {
				t.Fatalf("ID = %#v, want %#v", err.ID, tt.id)
			}
			if err.Message != tt.message {
				t.Fatalf("Message = %q, want %q", err.Message, tt.message)
			}
			if got := err.Error(); got != tt.wantError {
				t.Fatalf("Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		reason    string
		message   string
		wantError string
	}{
		{
			name:      "異常系: トークン未提供",
			reason:    "token_not_found",
			message:   UnauthorizedMessageAuth,
			wantError: UnauthorizedMessageAuth,
		},
		{
			name:      "異常系: メール・パスワード不一致",
			reason:    "invalid_credentials",
			message:   UnauthorizedMessageEmailOrPassword,
			wantError: UnauthorizedMessageEmailOrPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := NewUnauthorizedError(tt.reason, tt.message)
			if err == nil {
				t.Fatal("NewUnauthorizedError() returned nil")
			}
			if err.Reason != tt.reason {
				t.Fatalf("Reason = %q, want %q", err.Reason, tt.reason)
			}
			if err.Message != tt.message {
				t.Fatalf("Message = %q, want %q", err.Message, tt.message)
			}
			if got := err.Error(); got != tt.wantError {
				t.Fatalf("Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestNewForbiddenError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		requiredRole string
		userRole     string
		message      string
		wantError    string
	}{
		{
			name:         "異常系: memberがadmin権限のエンドポイントにアクセス",
			requiredRole: "admin",
			userRole:     "member",
			message:      ForbiddenMessageAdmin,
			wantError:    ForbiddenMessageAdmin,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := NewForbiddenError(tt.requiredRole, tt.userRole, tt.message)
			if err == nil {
				t.Fatal("NewForbiddenError() returned")
			}
			if err.RequiredRole != tt.requiredRole {
				t.Fatalf("RequiredRole = %q, want %q", err.RequiredRole, tt.requiredRole)
			}
			if err.UserRole != tt.userRole {
				t.Fatalf("UserROle = %q, want %q", err.UserRole, tt.userRole)
			}
			if err.Message != tt.message {
				t.Fatalf("Message = %q, want %q", err.Message, tt.message)
			}
			if got := err.Error(); got != tt.wantError {
				t.Fatalf("Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}

func TestNewBusinessLogicError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		message   string
		wantError string
	}{
		{
			name:      "異常系: 自分自身のロール変更",
			message:   BusinessLogicMessageRole,
			wantError: BusinessLogicMessageRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := NewBusinessLogicError(tt.message)
			if err == nil {
				t.Fatal("NewBusinessLogicError() returned nil")
			}
			if err.Message != tt.message {
				t.Fatalf("Message = %q, want %q", err.Message, tt.message)
			}
			if got := err.Error(); got != tt.wantError {
				t.Fatalf("Error() = %q, want %q", got, tt.wantError)
			}
		})
	}
}
