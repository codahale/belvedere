package check

import (
	"encoding/json"
	"fmt"
)

type FailedOperationError struct {
	Message json.Marshaler
}

func (e *FailedOperationError) Error() string {
	j, _ := e.Message.MarshalJSON()
	return fmt.Sprintf("operation failed: %s", string(j))
}
