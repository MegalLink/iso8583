package iso8583

import (
	"github.com/MegalLink/iso8583/field"
)

// MessageInterface defines the public methods available for the Message type
type MessageInterface interface {
	// Data returns the underlying data structure
	Data() interface{}

	// SetData sets the data structure that will be used to populate message fields
	SetData(data interface{}) error

	// Bitmap returns the message bitmap
	Bitmap() *field.Bitmap

	// MTI sets the Message Type Indicator
	SetMTI(val string)

	// GetMTI returns the Message Type Indicator
	GetMTI() (string, error)

	// GetSpec returns the message specification
	GetSpec() *MessageSpec

	// Field sets a string field value by ID
	Field(id int, val string) error

	// BinaryField sets a binary field value by ID
	BinaryField(id int, val []byte) error

	// GetString returns a string representation of the field value by ID
	GetString(id int) (string, error)

	// GetBytes returns the raw bytes of the field value by ID
	GetBytes(id int) ([]byte, error)

	// GetField returns the field by ID
	GetField(id int) field.Field

	// GetFields returns a map of all set fields
	GetFields() map[int]field.Field

	// Pack packs the message into bytes
	Pack() ([]byte, error)

	// Unpack unpacks the message from bytes
	Unpack(src []byte) error

	// MarshalJSON implements json.Marshaler interface
	MarshalJSON() ([]byte, error)
}
