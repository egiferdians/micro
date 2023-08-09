package errcode

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New grpc error code
func New(statusCode codes.Code, err error) error {
	return status.Error(statusCode, err.Error())
}