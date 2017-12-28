package gremlin

import "encoding/json"

const (
	OpAuthentication = "authentication"
	OpBytecode       = "bytecode"
	OpEval           = "eval"
	OpClose          = "close"
	OpGather         = "gather"
	OpKeys           = "keys"
)

type Request struct {
	RequestID string      `json:"requestId"`
	Op        string      `json:"op"`
	Processor string      `json:"processor"`
	Args      interface{} `json:"args"`
}

type AuthenticationInput struct {
	SASL          string `json:"sasl,omitempty"`
	SASLMechanism string `json:"saslMechanism,omitempty"`
}

type EvalInput struct {
	Script    string                 `json:"gremlin,omitempty"`
	Bindings  map[string]interface{} `json:"bindings,omitempty"`
	Aliases   map[string]interface{} `json:"aliases,omitempty"`
	BatchSize uint64                 `json:"batchSize,omitempty"`

	Language                string `json:"language,omitempty"`
	ScriptEvaluationTimeout uint64 `json:"scriptEvaluationTimeout,omitempty"`
}

type RawOutput []byte

func (out RawOutput) Scan(dst interface{}) error {
	return json.Unmarshal(out, dst)
}

type RequestArgsSessionAuthentication struct {
	*AuthenticationInput
}

type RequestArgsSessionEval struct {
	*EvalInput
	Session           string `json:"session,omitempty"`
	ManageTransaction bool   `json:"manageTransaction,omitempty"`
}

type RequestArgsSessionClose struct {
	Session string `json:"session,omitempty"`
	Force   bool   `json:"force,omitempty"`
}

type RequestArgsTraversalAuthentication struct {
	SASL string `json:"sasl,omitempty"`
}

type RequestArgsTraversalBytecode struct {
	Gremlin string            `json:"gremlin,omitempty"`
	Aliases map[string]string `json:"aliases,omitempty"`
}

type RequestArgsTraversalClose struct {
	SideEffect string `json:"sideEffect,omitempty"`
}

type RequestArgsTraversalGather struct {
	SideEffect    string                 `json:"sideEffect,omitempty"`
	SideEffectKey string                 `json:"sideEffectKey,omitempty"`
	Aliases       map[string]interface{} `json:"aliases,omitempty"`
}

type RequestArgsTraversalKeys struct {
	SideEffect string `json:"sideEffect,omitempty"`
}

type Response struct {
	RequestID string          `json:"requestId,omitempty"`
	Status    *ResponseStatus `json:"status,omitempty"`
	Result    *ResponseResult `json:"result,omitempty"`
}

type ResponseStatus struct {
	Code       int                    `json:"code,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	Message    string                 `json:"message,omitempty"`
}

type ResponseResult struct {
	Data json.RawMessage        `json:"data,omitempty"`
	Meta map[string]interface{} `json:"meta,omitempty"`
}

const (
	StatusSuccess                  = 200
	StatusNoContent                = 204
	StatusPartialContent           = 206
	StatusUnauthorized             = 401
	StatusAuthenticate             = 407
	StatusMalformedRequest         = 498
	StatusInvalidRequestArguments  = 499
	StatusServerError              = 500
	StatusScriptEvaluationError    = 597
	StatusServerTimeout            = 598
	StatusServerSerializationError = 599
)

var StatusMessages = map[int]string{
	StatusUnauthorized:             "Unauthorized",
	StatusAuthenticate:             "Authenticate",
	StatusMalformedRequest:         "Malformed Request",
	StatusInvalidRequestArguments:  "Invalid Request Arguments",
	StatusServerError:              "Server Error",
	StatusScriptEvaluationError:    "Script Evaluation Error",
	StatusServerTimeout:            "Server Timeout",
	StatusServerSerializationError: "Server Serialization Error",
}