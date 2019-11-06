package internal

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

func ConvertValidationErrors(errMessages map[string]string) *errdetails.BadRequest {
	if len(errMessages) == 0 {
		return nil
	}

	fvs := make([]*errdetails.BadRequest_FieldViolation, 0, len(errMessages))
	for field, message := range errMessages {
		fvs = append(fvs, &errdetails.BadRequest_FieldViolation{
			Field:       field,
			Description: message,
		})
	}

	br := &errdetails.BadRequest{
		FieldViolations: fvs,
	}
	return br
}
