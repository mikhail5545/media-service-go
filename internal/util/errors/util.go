package errors

import (
	"errors"
	"fmt"
	"net/http"

	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Details any    `json:"details,omitempty"`
	} `json:"error"`
}

func MapServiceError(err error) (int, ErrorResponse) {
	var resp ErrorResponse

	switch {
	case errors.Is(err, serviceerrors.ErrAlreadyExists):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrAlreadyExists]
		resp.Error.Message = "Already exists"
		resp.Error.Details = err.Error()
		return http.StatusConflict, resp
	case errors.Is(err, serviceerrors.ErrCanceled):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrCanceled]
		resp.Error.Message = "Context canceled"
		resp.Error.Details = err.Error()
		return http.StatusGatewayTimeout, resp
	case errors.Is(err, serviceerrors.ErrConflict):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrConflict]
		resp.Error.Message = "Resource conflict"
		resp.Error.Details = err.Error()
		return http.StatusConflict, resp
	case errors.Is(err, serviceerrors.ErrInvalidArgument):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrInvalidArgument]
		resp.Error.Message = "Invalid argument"
		resp.Error.Details = err.Error()
		return http.StatusBadRequest, resp
	case errors.Is(err, serviceerrors.ErrNotFound):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrNotFound]
		resp.Error.Message = "Resource not found"
		resp.Error.Details = err.Error()
		return http.StatusNotFound, resp
	case errors.Is(err, serviceerrors.ErrPermissionDenied):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrPermissionDenied]
		resp.Error.Message = "Permission denied"
		resp.Error.Details = err.Error()
		return http.StatusForbidden, resp
	case errors.Is(err, serviceerrors.ErrTooManyRequests):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrTooManyRequests]
		resp.Error.Message = "Too many requests"
		resp.Error.Details = err.Error()
		return http.StatusTooManyRequests, resp
	case errors.Is(err, serviceerrors.ErrUnimplemented):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrUnimplemented]
		resp.Error.Message = "Unimplemented"
		resp.Error.Details = err.Error()
		return http.StatusNotImplemented, resp
	case errors.Is(err, serviceerrors.ErrUnavailable):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrUnavailable]
		resp.Error.Message = "Service unavailable"
		resp.Error.Details = err.Error()
		return http.StatusServiceUnavailable, resp
	case errors.Is(err, serviceerrors.ErrValidationFailed):
		resp.Error.Code = serviceerrors.ErrorAliases[serviceerrors.ErrValidationFailed]
		resp.Error.Message = "Validation failed"
		resp.Error.Details = err.Error()
		return http.StatusUnprocessableEntity, resp
	default:
		resp.Error.Code = "INTERNAL_SERVER_ERROR"
		resp.Error.Message = "Internal server error"
		resp.Error.Details = err.Error()
		return http.StatusInternalServerError, resp
	}
}

func HandleRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.NotFound:
		return serviceerrors.NewNotFoundError(err.Error())
	case codes.InvalidArgument:
		return serviceerrors.NewValidationFailedError(err.Error())
	case codes.AlreadyExists:
		return serviceerrors.NewAlreadyExistsError(err.Error())
	case codes.PermissionDenied:
		return serviceerrors.NewPermissionDeniedError(err.Error())
	case codes.Internal:
		fallthrough
	default:
		return fmt.Errorf("gRPC client internal error: %w", err)
	}
}

func ToGRPCCode(err error) error {
	if err == nil {
		return nil
	}

	if _, ok := status.FromError(err); ok {
		return err
	}

	switch {
	case errors.Is(err, serviceerrors.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, serviceerrors.ErrCanceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, serviceerrors.ErrConflict):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, serviceerrors.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, serviceerrors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, serviceerrors.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, serviceerrors.ErrTooManyRequests):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, serviceerrors.ErrUnimplemented):
		return status.Error(codes.Unimplemented, err.Error())
	case errors.Is(err, serviceerrors.ErrUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	case errors.Is(err, serviceerrors.ErrValidationFailed):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
