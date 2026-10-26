//go:build gohp_pcre2 && cgo

package pcre2

/*
#include "bridge.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Match runs one byte-oriented PCRE2 match through the opt-in C backend. The
// C boundary copies both inputs and all returned data, so no C or Go pointer is
// retained after this call returns.
func Match(request Request) (Result, error) {
	if err := validate(request); err != nil {
		return Result{}, err
	}

	response := C.pcre2_bridge_match(bytesPointer(request.Pattern), C.size_t(len(request.Pattern)),
		bytesPointer(request.Subject), C.size_t(len(request.Subject)), C.size_t(request.Offset),
		C.uint32_t(request.Limits.Match), C.uint32_t(request.Limits.Depth), C.uint32_t(request.Limits.HeapKiB))
	if response == nil {
		return Result{}, &EngineError{Code: int(C.PCRE2_BRIDGE_ERROR_ALLOC), Message: "cannot allocate bridge response"}
	}
	defer C.pcre2_bridge_response_free(response)

	status := int(C.pcre2_bridge_response_status(response))
	if status == int(C.PCRE2_BRIDGE_COMPILE) {
		return Result{}, &CompileError{
			Code:    int(C.pcre2_bridge_response_error_code(response)),
			Offset:  int(C.pcre2_bridge_response_error_offset(response)),
			Message: C.GoString(C.pcre2_bridge_response_message(response)),
		}
	}

	if status != int(C.PCRE2_BRIDGE_OK) {
		// An allocation/profile/bounds failure may leave partially built C
		// vectors. Never expose or traverse them, even on an error path.
		if status == int(C.PCRE2_BRIDGE_ERROR_PROFILE) || status == int(C.PCRE2_BRIDGE_ERROR_SIZE) {
			return Result{}, fmt.Errorf("%w: %s", ErrInvalidInput, C.GoString(C.pcre2_bridge_response_message(response)))
		}
		return Result{}, &EngineError{
			Code:    int(C.pcre2_bridge_response_return_code(response)),
			Message: C.GoString(C.pcre2_bridge_response_message(response)),
		}
	}
	result := copyResult(response)
	if result.ReturnCode != NoMatch && result.ReturnCode != MatchLimit && result.ReturnCode <= 0 {
		return result, &EngineError{
			Code:    result.ReturnCode,
			Message: C.GoString(C.pcre2_bridge_response_message(response)),
		}
	}
	return result, nil
}

func validate(request Request) error {
	switch {
	case request.Profile != ByteDotAll:
		return fmt.Errorf("%w: unsupported profile %q", ErrInvalidInput, request.Profile)
	case len(request.Pattern) > MaxPatternBytes:
		return fmt.Errorf("%w: pattern is %d bytes (maximum %d)", ErrInvalidInput, len(request.Pattern), MaxPatternBytes)
	case len(request.Subject) > MaxSubjectBytes:
		return fmt.Errorf("%w: subject is %d bytes (maximum %d)", ErrInvalidInput, len(request.Subject), MaxSubjectBytes)
	case request.Offset < 0 || request.Offset > len(request.Subject):
		return fmt.Errorf("%w: byte offset %d is outside subject", ErrInvalidInput, request.Offset)
	case request.Limits.Match == 0 || request.Limits.Depth == 0 || request.Limits.HeapKiB == 0:
		return fmt.Errorf("%w: all limits must be non-zero", ErrInvalidInput)
	case request.Limits.Match > MaxMatchLimit || request.Limits.Depth > MaxDepthLimit || request.Limits.HeapKiB > MaxHeapKiB:
		return fmt.Errorf("%w: limits exceed byte-dotall-v1 ceilings", ErrInvalidInput)
	}
	return nil
}

func bytesPointer(bytes []byte) *C.uchar {
	if len(bytes) == 0 {
		return nil
	}
	return (*C.uchar)(unsafe.Pointer(&bytes[0]))
}

func copyResult(response *C.pcre2_bridge_response) Result {
	result := Result{
		Version:      C.GoString(C.pcre2_bridge_response_version(response)),
		ReturnCode:   int(C.pcre2_bridge_response_return_code(response)),
		CaptureCount: int(C.pcre2_bridge_response_capture_count(response)),
	}
	nameCount := uint32(C.pcre2_bridge_response_name_count(response))
	result.Names = make([]Name, 0, nameCount)
	for i := uint32(0); i < nameCount; i++ {
		nameLength := int(C.pcre2_bridge_response_name_length(response, C.uint32_t(i)))
		result.Names = append(result.Names, Name{
			Number: int(C.pcre2_bridge_response_name_number(response, C.uint32_t(i))),
			Name:   C.GoStringN(C.pcre2_bridge_response_name(response, C.uint32_t(i)), C.int(nameLength)),
		})
	}
	resultCount := uint32(C.pcre2_bridge_response_result_count(response))
	result.Captures = make([]Capture, 0, resultCount)
	for i := uint32(0); i < resultCount; i++ {
		present := C.pcre2_bridge_response_capture_present(response, C.uint32_t(i)) != 0
		capture := Capture{Number: int(i), Start: -1, End: -1, Bytes: []byte{}}
		if present {
			length := int(C.pcre2_bridge_response_capture_length(response, C.uint32_t(i)))
			capture.Present = true
			capture.Start = int(C.pcre2_bridge_response_capture_start(response, C.uint32_t(i)))
			capture.End = int(C.pcre2_bridge_response_capture_end(response, C.uint32_t(i)))
			capture.Bytes = C.GoBytes(unsafe.Pointer(C.pcre2_bridge_response_capture_bytes(response, C.uint32_t(i))), C.int(length))
		}
		result.Captures = append(result.Captures, capture)
	}
	return result
}
