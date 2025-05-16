package analyzer

import "fmt"

type DetectorError struct {
	Type    string
	Message string
	Err     error
}

func (e *DetectorError) Error() string {
	return fmt.Sprintf("%s: %s: %v", e.Type, e.Message, e.Err)
}

func NewDetectorError(typ, msg string, err error) error {
	return &DetectorError{
		Type:    typ,
		Message: msg,
		Err:     err,
	}
}
