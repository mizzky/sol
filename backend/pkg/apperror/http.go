package apperror

import (
	"errors"
	"net/http"
)

var validationMessages = map[string]string{
	"email":    ValidationMessageEmail,
	"password": ValidationMessagePassword,
	"name":     ValidationMessageName,
	"sku":      ValidationMessageSku,
	"id":       ValidationMessageID,
	"order":    ValidationMessageOrder,
	"status":   ValidationMessageStatus,
	"price":    ValidationMessagePrice,
	"request":  ValidationMessageRequest,
	"role":     ValidationMessageRole,
	"cart":     ValidationMessageCart,
	"category": ValidationMessageCategory,
	"qty":      ValidationMessageQty,
}

var conflictMessages = map[string]string{
	"qty": ConflictMessageQty,
	"sku": ConflictMessageSku,
}

var notFoundMessages = map[string]string{
	"product":   NotFoundMessageProduct,
	"cart":      NotFoundMessageCart,
	"cart_item": NotFoundMessageCartItem,
	"user":      NotFoundMessageUser,
	"category":  NotFoundMessageCategory,
	"order":     NotFoundMessageOrder,
}

func ToHTTP(err error) (status int, message string) {
	if err == nil {
		return http.StatusInternalServerError, InternalServerMessageCommon
	}

	var ve *ValidationError
	if errors.As(err, &ve) {
		if ve.Message != "" {
			return http.StatusBadRequest, ve.Message
		}
		if m, ok := validationMessages[ve.Field]; ok {
			return http.StatusBadRequest, m
		}
		return http.StatusBadRequest, ValidationMessageGeneric
	}

	var ce *ConflictError
	if errors.As(err, &ce) {
		if ce.Message != "" {
			return http.StatusConflict, ce.Message
		}
		if m, ok := conflictMessages[ce.Field]; ok {
			return http.StatusConflict, m
		}
		return http.StatusConflict, ConflictMessageGeneric
	}

	var ne *NotFoundError
	if errors.As(err, &ne) {
		if ne.Message != "" {
			return http.StatusNotFound, ne.Message
		}
		if m, ok := notFoundMessages[ne.Resource]; ok {
			return http.StatusNotFound, m
		}
		return http.StatusNotFound, NotFoundMessageGeneric
	}

	var ue *UnauthorizedError
	if errors.As(err, &ue) {
		if ue.Message != "" {
			return http.StatusUnauthorized, ue.Message
		}
		return http.StatusUnauthorized, UnauthorizedMessageGeneric
	}

	var fe *ForbiddenError
	if errors.As(err, &fe) {
		if fe.Message != "" {
			return http.StatusForbidden, fe.Message
		}
		return http.StatusForbidden, ForbiddenMessageGeneric
	}

	var be *BusinessLogicError
	if errors.As(err, &be) {
		if be.Message != "" {
			return http.StatusBadRequest, be.Message
		}
		return http.StatusBadRequest, BusinessLogicMessageGeneric
	}

	var ie *InternalError
	if errors.As(err, &ie) {
		if ie.Message != "" {
			return http.StatusInternalServerError, ie.Message
		}
		return http.StatusInternalServerError, InternalServerMessageCommon
	}

	// fallback
	return http.StatusInternalServerError, InternalServerMessageCommon
}
