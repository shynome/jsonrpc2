package jsonrpc2

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMethodMap(t *testing.T) {
	assert := assert.New(t)

	var methods MethodMap
	assert.EqualError(methods.IsValid(), "nil MethodMap")

	methods = MethodMap{}
	assert.EqualError(methods.IsValid(), "empty MethodMap")

	methods = MethodMap{"": func(params json.RawMessage) Response { return Response{} }}
	assert.EqualError(methods.IsValid(), "empty name")

	methods = MethodMap{"test": MethodFunc(nil)}
	assert.EqualError(methods.IsValid(),
		fmt.Sprintf("nil MethodFunc for method %#v", "test"))
}

func TestMethodFuncCall(t *testing.T) {
	assert := assert.New(t)

	var buf bytes.Buffer
	logger.SetOutput(&buf) // hide output
	DebugMethodFunc = true
	var fs []MethodFunc
	fs = append(fs, func(_ json.RawMessage) Response {
		return NewErrorResponse(MethodNotFoundCode, "method not found", "test data")
	}, func(_ json.RawMessage) Response {
		return Response{}
	}, func(_ json.RawMessage) Response {
		return Response{Error: Error{Message: "e", Data: map[bool]bool{true: true}}}
	}, func(_ json.RawMessage) Response {
		return Response{Result: map[bool]bool{true: true}}
	})
	for _, f := range fs {
		res := f.call(nil)
		if assert.NotNil(res.Error) {
			assert.Equal(InternalError, res.Error)
		}
		assert.Nil(res.Result)
	}
	assert.Equal("Internal error: \"Invalid Response.Error\"\nParams: \nResponse: <-- {\"jsonrpc\":\"2.0\",\"error\":{\"code\":-32601,\"message\":\"method not found\",\"data\":\"test data\"},\"id\":null}\nInternal error: \"Both Response.Result and Response.Error are empty\"\nParams: \nResponse: <-- \nInternal error: \"Cannot marshal Response.Error.Data\"\nParams: \nResponse: <-- \nInternal error: \"Cannot marshal Response.Result\"\nParams: \nResponse: <-- \n",
		string(buf.Bytes()))

	var f MethodFunc = func(_ json.RawMessage) Response {
		return NewErrorResponse(100, "custom", "data")
	}
	res := f.call(nil)
	if assert.NotNil(res.Error) {
		assert.Equal(Error{
			Code:    100,
			Message: "custom",
			Data:    json.RawMessage(`"data"`),
		}, res.Error)
	}
	assert.Nil(res.Result)

	f = func(_ json.RawMessage) Response {
		return NewInvalidParamsErrorResponse("data")
	}
	res = f.call(nil)
	if assert.NotNil(res.Error) {
		e := InvalidParams
		e.Data = json.RawMessage(`"data"`)
		assert.Equal(e, res.Error)
	}
	assert.Nil(res.Result)
}
