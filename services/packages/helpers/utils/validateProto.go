package helpers_utils

import (
	protovalidate "buf.build/go/protovalidate"
	"google.golang.org/protobuf/proto"
)

// Validate Protobuf request
func ValidateProto[P proto.Message](req P) error {
	validator, err := protovalidate.New()
	if err != nil {
		return err
	}
	return validator.Validate(req)
}
