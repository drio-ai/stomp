package frame

import (
	"errors"
	"math"
)

const (
	DefaultMaxHeaders         uint32 = 64
	DefaultMaxHeaderNameSize  uint32 = 256
	DefaultMaxHeaderValueSize uint32 = 4096
	DefaultMaxBodySize        uint32 = 1048576

	// We are not going down the path of 0 implies
	// check is off. If the caller wants to disable
	// some checks, they can set the corresponding
	// value to ConstraintsMax. This is a high enough
	// value to cover all exceptional cases.
	ConstraintsMax uint32 = math.MaxUint32
)

var (
	ErrTooManyHeaders      = errors.New("too many headers")
	ErrHeaderNameTooLarge  = errors.New("header name too large")
	ErrHeaderValueTooLarge = errors.New("header value too large")
	ErrBodyTooLarge        = errors.New("body too large")
	ErrInvalidContentType  = errors.New("invalid content type")
	ErrMissingContentType  = errors.New("missing content type")
)

// Constraints that will be applied on a
// frame received from the client.
type Constraints struct {
	// Must be enabled for constraints to be applied.
	// Defaults to disabled
	Enabled bool

	// Limit constraints.
	MaxHeaders         uint32
	MaxHeaderNameSize  uint32
	MaxHeaderValueSize uint32
	MaxBodySize        uint32

	// Allowed Content Types.
	EnforceAllowedContentTypes bool
	AllowedContentTypes        map[string]bool
}

// Returns an initialized instance of Constraints
// with sane defaults. It is recommended to call
// this to initialize Constraints instance.
func NewConstraints() *Constraints {
	return &Constraints{
		MaxHeaders:         DefaultMaxHeaders,
		MaxHeaderNameSize:  DefaultMaxHeaderNameSize,
		MaxHeaderValueSize: DefaultMaxHeaderValueSize,
		MaxBodySize:        DefaultMaxBodySize,
	}
}

// Given the frame, checks if header count does not exceed threshold
func (c *Constraints) ValidateMaxHeaders(f *Frame) error {
	if c != nil && c.Enabled {
		hLen := uint32(f.Header.Len())
		if hLen > c.MaxHeaders {
			return ErrTooManyHeaders
		}
	}

	return nil
}

// Ensures header name length does not exceed limit
func (c *Constraints) ValidateHeaderNameLen(nLen int) error {
	if c != nil && c.Enabled && uint32(nLen) > c.MaxHeaderNameSize {
		return ErrHeaderNameTooLarge
	}

	return nil
}

// Ensures header value length does not exceed limit
func (c *Constraints) ValidateHeaderValueLen(vLen int) error {
	if c != nil && c.Enabled && uint32(vLen) <= c.MaxHeaderValueSize {
		return ErrHeaderValueTooLarge
	}

	return nil
}

// Ensures body size does not exceed limit
func (c *Constraints) ValidateBodyLen(bLen int) error {
	if c != nil && c.Enabled && uint32(bLen) > c.MaxBodySize {
		return ErrBodyTooLarge
	}

	return nil
}

// Check if passed Content-Type header value is allowed
func (c *Constraints) ValidateContentType(f *Frame) error {
	if c != nil && c.Enabled && c.EnforceAllowedContentTypes {
		// Get content type header if present
		contentType, present := f.Header.Contains(ContentType)

		// Is there a content-type header
		if present {
			// We are not going to validate the format of the content-type
			// header. We will leave that to the application.
			// Value must be in the map and set to true.
			// NOTE: If the content-type header is present and has an empty
			// value, we will still check to make sure it is allowed.
			v, e := c.AllowedContentTypes[contentType]
			if !e || !v {
				return ErrInvalidContentType
			}
		} else {
			// There must be a Content-Type header if there is a body
			// and the frame command is either of SEND, MESSAGE or ERROR.
			if len(f.Body) > 0 && (f.Command == SEND || f.Command == MESSAGE || f.Command == ERROR) {
				return ErrMissingContentType
			}
		}
	}

	return nil
}
