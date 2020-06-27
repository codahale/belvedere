package check

import (
	"encoding/json"
	"fmt"
)

type failedOperationError struct {
	Message json.Marshaler
}

func (e *failedOperationError) Error() string {
	j, _ := e.Message.MarshalJSON()
	return fmt.Sprintf("operation failed: %s", string(j))
}
