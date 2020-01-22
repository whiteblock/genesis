/**
 * Copyright 2019 Whiteblock Inc. All rights reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package entity

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/getlantern/deepcopy"
	"github.com/imdario/mergo"
)

// ResultType is the type of the result
type ResultType int

// Result is the result of executing the command, contains a type and possibly an error
type Result struct {
	// Error is where the error is stored if this result is not a successful result
	Error error
	// Type is the type of result
	Type ResultType

	// Meta is additional information which can be added for debugging purposes
	Meta map[string]interface{}

	//Caller is the location in which it was first created
	Caller string

	//Delay is the delay for the next round of execution
	Delay time.Duration
}

// IsAllDone checks whether the request is completely finished. If true, then the completion
// protocol should be followed
func (res Result) IsAllDone() bool {
	return res.Error == nil && res.Type == AllDoneType
}

// IsSuccess returns whether or not the result indicates success. Both AllDoneType and
// SuccessType count as being successful
func (res Result) IsSuccess() bool {
	return res.Error == nil
}

// IsTrap returns whether or not the result indicates a trap was raised.
func (res Result) IsTrap() bool {
	return res.Type == TrapType
}

// IsFatal returns true if there is an errr and it is marked as a fatal error,
// meaning it should not be reattempted
func (res Result) IsFatal() bool {
	return res.Error != nil && res.Type == FatalType
}

// IsIgnore returns true if this should be ignored
func (res Result) IsIgnore() bool {
	return res.Type == IgnoreType
}

// CopyTo copies this result's data into another result
func (res Result) CopyTo(out *Result) {
	if out == nil {
		out = new(Result)
	}
	*out = res
	deepcopy.Copy(&out.Meta, res.Meta)
}

// Trap turns this result into a trapping result
func (res Result) Trap() (out Result) {
	res.CopyTo(&out)
	out.Type = TrapType
	return out
}

// Fatal turns this result into a fatal result, useful for when you want to change
// the resulting action, but want to preserve the information in the result.
// If no args given, keeps the original error
func (res Result) Fatal(err ...error) (out Result) {
	res.CopyTo(&out)
	out.Type = FatalType
	if len(err) > 0 {
		out.Error = err[0]
	}
	return out
}

// IsRequeue returns true if this result indicates that the command should be retried at a
// later time
func (res Result) IsRequeue() bool {
	return res.Type == RequeueType || !res.IsSuccess() && !res.IsFatal()
}

// IsDelayed returns true if this result is a delay result with a delay greater than 0
func (res Result) IsDelayed() bool {
	return res.Type == DelayType && res.Delay > 0
}

// InjectMeta allows for chaining on New...Result for the return statement
func (res Result) InjectMeta(meta map[string]interface{}) (out Result) {
	res.CopyTo(&out)
	mergo.Map(&out.Meta, meta)
	return out
}

// MarshalJSON allows Result to customize the marshaling into JSON
func (res Result) MarshalJSON() ([]byte, error) {
	resType := ""
	switch res.Type {
	case SuccessType:
		resType = "Success"
	case AllDoneType:
		resType = "AllDone"
	case TooSoonType:
		resType = "TooSoon"
	case FatalType:
		resType = "Fatal"
	case ErrorType:
		resType = "Error"
	case RequeueType:
		resType = "Requeue"
	case TrapType:
		resType = "Trap"
	case DelayType:
		resType = "Delay"
	default:
		resType = "Unknown"
	}

	jRes := map[string]interface{}{
		"type":   resType,
		"meta":   res.Meta,
		"caller": res.Caller,
	}
	if res.Error != nil {
		jRes["error"] = res.Error.Error()
	} else {
		jRes["error"] = nil
	}
	return json.Marshal(jRes)
}

const (
	//SuccessType is the type of a successful result
	SuccessType ResultType = iota + 1

	// AllDoneType is the type for when all of the commands executed successfully and
	// there are not anymore commands to execute
	AllDoneType

	// TooSoonType is the type of a result from a cmd which tried to execute too soon
	TooSoonType

	// FatalType is the type of a result which indicates a fatal error
	FatalType

	// ErrorType is the generic error type
	ErrorType

	// RequeueType is when there is no error but should requeue something
	RequeueType

	// TrapType indicates that the task is a trap, and the commands should just be acked without
	// further action
	TrapType

	// IgnoreType indicates that the given payload should be dropped immediately without further action
	IgnoreType

	// DelayType indicates that the given payload should continue as if a SuccessType was returned, but should
	// be requeued on a time delay according to the value in Delay
	DelayType
)

func getCaller(n int) string {
	_, file, line, ok := runtime.Caller(n)
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// NewResult creates a success result if err == nil other an error result,
func NewResult(err interface{}, depth ...int) Result {
	n := 2
	if len(depth) > 0 {
		n += depth[0]
	}
	if err == nil {
		return Result{Type: SuccessType, Error: nil,
			Meta: map[string]interface{}{}, Caller: getCaller(n)}
	}
	return Result{Type: ErrorType, Error: fmt.Errorf("%v", err),
		Meta: map[string]interface{}{}, Caller: getCaller(n)}
}

// NewSuccessResult indicates a successful result
func NewSuccessResult() Result {
	return Result{Type: SuccessType, Error: nil,
		Meta: map[string]interface{}{}, Caller: getCaller(2)}
}

// NewDelayResult indicates a time delay result, aka a success result with a delay
func NewDelayResult(delay time.Duration) Result {
	return Result{Type: DelayType, Error: nil, Delay: delay,
		Meta: map[string]interface{}{}, Caller: getCaller(2)}
}

// NewTrapResult creates a new Trapping result
func NewTrapResult() Result {
	return Result{Type: TrapType, Error: nil,
		Meta: map[string]interface{}{}, Caller: getCaller(2)}
}

// NewFatalResult creates a fatal error result. Commands with fatal errors are not retried
func NewFatalResult(err interface{}) Result {
	return Result{Type: FatalType, Error: fmt.Errorf("%v", err),
		Meta: map[string]interface{}{}, Caller: getCaller(2)}
}

// NewErrorResult creates a result which indicates a non-fatal error.
// Commands with this result should be requeued.
func NewErrorResult(err interface{}) Result {
	return Result{Type: ErrorType, Error: fmt.Errorf("%v", err),
		Meta: map[string]interface{}{}, Caller: getCaller(2)}
}

// NewIgnoreResult creates a result which indicates to just ack the message, and ignore it
func NewIgnoreResult(err interface{}) Result {
	return Result{Type: IgnoreType, Error: fmt.Errorf("%v", err),
		Meta: map[string]interface{}{}, Caller: getCaller(2)}
}

// NewAllDoneResult creates a result for the all done condition
func NewAllDoneResult() Result {
	return Result{Type: AllDoneType, Error: nil,
		Meta: map[string]interface{}{}, Caller: getCaller(2)}
}

// NewRequeueResult creates a new requeue result non-error
func NewRequeueResult() Result {
	return Result{Type: RequeueType, Error: nil,
		Meta: map[string]interface{}{}, Caller: getCaller(2)}
}
