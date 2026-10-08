package graphqlmapper

import (
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func PointerString(s *wrapperspb.StringValue) *string {
	if s == nil {
		return nil
	}
	v := s.Value
	return &v
}
