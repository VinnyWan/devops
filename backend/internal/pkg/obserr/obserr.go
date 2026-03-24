package obserr

import "errors"

type ObservableError struct {
	Code    string
	Op      string
	Message string
	Err     error
}

func New(code, op, message string) *ObservableError {
	return &ObservableError{
		Code:    code,
		Op:      op,
		Message: message,
	}
}

func Wrap(code, op, message string, err error) *ObservableError {
	return &ObservableError{
		Code:    code,
		Op:      op,
		Message: message,
		Err:     err,
	}
}

func (e *ObservableError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		if e.Op == "" {
			return e.Message
		}
		return e.Op + ": " + e.Message
	}
	if e.Op == "" {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Op + ": " + e.Message + ": " + e.Err.Error()
}

func (e *ObservableError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type Node struct {
	Type    string `json:"type"`
	Code    string `json:"code,omitempty"`
	Op      string `json:"op,omitempty"`
	Message string `json:"message"`
}

func Details(err error) map[string]interface{} {
	if err == nil {
		return nil
	}
	code := "INTERNAL_ERROR"
	message := err.Error()
	var oe *ObservableError
	if errors.As(err, &oe) {
		if oe.Code != "" {
			code = oe.Code
		}
		if oe.Message != "" {
			message = oe.Message
		}
	}
	return map[string]interface{}{
		"code":    code,
		"message": message,
		"chain":   Chain(err),
	}
}

func Chain(err error) []Node {
	chain := make([]Node, 0, 4)
	for current := err; current != nil; current = errors.Unwrap(current) {
		node := Node{
			Type: "error",
		}
		var oe *ObservableError
		if errors.As(current, &oe) && oe != nil {
			node.Type = "observable"
			node.Code = oe.Code
			node.Op = oe.Op
			if oe.Message != "" {
				node.Message = oe.Message
			} else {
				node.Message = current.Error()
			}
		} else {
			node.Message = current.Error()
		}
		chain = append(chain, node)
	}
	return chain
}
