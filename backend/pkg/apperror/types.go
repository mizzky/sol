package apperror

import "reflect"

type ValidationError struct {
	Field     string
	Value     any
	ValueType string
	Rule      string
	Message   string
}

func NewValidationError(field string, value any, rule string, message string) *ValidationError {
	v := value
	if field == "email" {
		if s, ok := value.(string); ok {
			v = maskEmail(s)
		}
	}
	var valueType string
	if value != nil {
		valueType = reflect.TypeOf(value).String()
	}
	return &ValidationError{
		Field:     field,
		Value:     v,
		ValueType: valueType,
		Rule:      rule,
		Message:   message,
	}
}

func (e *ValidationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Resource string
	ID       any
	Message  string
}

func NewNotFoundError(resource string, id any, message string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
		Message:  message,
	}
}

func (e *NotFoundError) Error() string {
	return e.Message
}

type ConflictError struct {
	Field   string
	Value   string
	Message string
}

func NewConflictError(field, value, message string) *ConflictError {
	return &ConflictError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

func (e *ConflictError) Error() string {
	return e.Message
}

type UnauthorizedError struct {
	Reason  string
	Message string
}

func NewUnauthorizedError(reason, message string) *UnauthorizedError {
	return &UnauthorizedError{
		Reason:  reason,
		Message: message,
	}
}

func (e *UnauthorizedError) Error() string {
	return e.Message
}

type ForbiddenError struct {
	RequiredRole string
	UserRole     string
	Message      string
}

func NewForbiddenError(requiredRole, userRole, message string) *ForbiddenError {
	return &ForbiddenError{
		RequiredRole: requiredRole,
		UserRole:     userRole,
		Message:      message,
	}
}

func (e *ForbiddenError) Error() string {
	return e.Message
}

type BusinessLogicError struct {
	Message string
}

func NewBusinessLogicError(message string) *BusinessLogicError {
	return &BusinessLogicError{
		Message: message,
	}
}

func (e *BusinessLogicError) Error() string {
	return e.Message
}

type InternalError struct {
	Operation string
	Cause     error
	Message   string
}

func NewInternalError(operation string, cause error, message string) *InternalError {
	return &InternalError{
		Operation: operation,
		Cause:     cause,
		Message:   message,
	}
}

func (e *InternalError) Error() string {
	return e.Message
}

func (e *InternalError) Unwrap() error {
	return e.Cause
}
