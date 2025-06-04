```
go get github.com/MegalLink/iso8583
```

### Define your specification

Currently, we support following ISO 8583 specifications:

* [Spec87ASCII](./specs/spec87ascii.go) - 1987 version of the spec with ASCII encoding
* [Spec87Hex](./specs/spec87hex.go) - 1987 version of the spec with Hex encoding

Spec87ASCII is suitable for the majority of use cases. Simply instantiate a new message using `specs.Spec87ASCII`:

```
isomessage := iso8583.NewMessage(specs.Spec87ASCII)
```
If this spec does not meet your needs, we encourage you to modify it or create your own using the information below. 

First, you need to define the format of the message fields that are described in your ISO8583 specification. Each data field has a type and its own spec. You can create a `NewBitmap`, `NewString`, or `NewNumeric` field. Each individual field spec consists of a few elements:

| Element          | Notes                                                                                                                                                                                                                       | Example                    |
|------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|----------------------------|
| `Length`         | Maximum length of field (bytes, characters or digits), for both fixed and variable lengths.                                                                                                                                 | `10`                       |
| `Description`    | Describes what the data field holds.                                                                                                                                                                                        | `"Primary Account Number"` |
| `Enc`            | Sets the encoding type (`ASCII`, `Hex`, `Binary`, `BCD`, `LBCD`, `EBCDIC`).                                                                                                                                                           | `encoding.ASCII`           |
| `Pref`           | Sets the encoding (`ASCII`, `Hex`, `Binary`, `BCD`, `EBCDIC`) of the field length and its type as fixed or variable (`Fixed`, `L`, `LL`, `LLL`, `LLLL`). The number of 'L's corresponds to the number of digits in a variable length. | `prefix.ASCII.Fixed`       |
| `Pad` (optional) | Sets padding direction and type.                                                                                                                                                                                            | `padding.Left('0')`        |

While some ISO8583 specifications do not have field 0 and field 1, we use them for MTI and Bitmap. Because technically speaking, they are just regular fields. We use field specs to describe MTI and Bitmap too. We currently use the `String` field for MTI, while we have a separate `Bitmap` field for the bitmap.

The following example creates a full specification with three individual fields (excluding MTI and Bitmap):

```go
spec := &iso8583.MessageSpec{
	Fields: map[int]field.Field{
		0: field.NewString(&field.Spec{
			Length:      4,
			Description: "Message Type Indicator",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}),
		1: field.NewBitmap(&field.Spec{
			Description: "Bitmap",
			Enc:         encoding.Hex,
			Pref:        prefix.Hex.Fixed,
		}),

		// Message fields:
		2: field.NewString(&field.Spec{
			Length:      19,
			Description: "Primary Account Number",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.LL,
		}),
		3: field.NewNumeric(&field.Spec{
			Length:      6,
			Description: "Processing Code",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
		}),
		4: field.NewString(&field.Spec{
			Length:      12,
			Description: "Transaction Amount",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
			Pad:         padding.Left('0'),
		}),
	},
}
```

### Build and pack the message

After the specification is defined, you can build a message. Having a binary representation of your message that's packed according to the provided spec lets you send it directly to a payment system! There are two ways to generate a message: typed or untyped.

Notice in the examples below, you do not need to set the bitmap value manually, as it is automatically generated for you during packing.

#### Untyped fields

If you don't care about types, and strings are fine for you, then you can easily set message fields like this:

```go
// create message with defined spec
message := NewMessage(spec)

// sets message type indicator at field 0
message.MTI("0100")

// set all message fields you need as strings
message.Field(2, "4242424242424242")
message.Field(3, "123456")
message.Field(4, "100")

// generate binary representation of the message into rawMessage
rawMessage, err := message.Pack()

// now you can send rawMessage over the wire
```

The binary message created will be `"01007000000000000000164242424242424242123456000000000100"`.

#### Typed fields

In many cases, you may want to work with specific types (numbers, strings, time, etc.) when setting message fields.

First, you need to define a struct that corresponds to the spec field types. Here is an example:

```go
// use the same types from message specification
type ISO87Data struct {
	F2 *field.StringField
	F3 *field.NumericField
	F4 *field.StringField
}

message := NewMessage(spec)
message.MTI("0100")

// now, pass data with fields into the message 
err := message.SetData(&ISO87Data{
	F2: field.NewStringValue("4242424242424242"),
	F3: field.NewNumericValue(123456),
	F4: field.NewStringValue("100"),
})

// pack the message and send it to your provider
rawMessage, err := message.Pack()
```

### Parse the message and access the data

When you have a binary (packed) message and you know the specification it follows, you can unpack it and access the data. Again, you have two options for data access: untyped and typed.

#### Untyped fields

With this approach you can easily access fields as strings:

```go
message := NewMessage(spec)
message.Unpack(rawMessage)

message.GetMTI() // MTI: 0100
message.GetString(2) // Card number: 4242424242424242
message.GetString(3) // Processing code: 123456
message.GetString(4) // Transaction amount: 100
```

#### Typed fields

To get typed field values just pass an empty struct for the data you want to access:

```go
// use the same types from message specification
type ISO87Data struct {
	F2 *field.StringField
	F3 *field.NumericField
	F4 *field.StringField
}

message := NewMessage(spec)
message.SetData(&ISO87Data{})

// let's unpack binary message
err := message.Unpack(rawMessage)

// to get access to typed data we have to get Data from the message and convert it into our ISO87Data type
data := message.Data().(*ISO87Data)

// now you have typed values
data.F2.Value // is a string "4242424242424242"
data.F3.Value // is an int 123456
data.F4.Value // is a string "100"
```

For complete code samples please check [./message_test.go](./message_test.go).

### JSON encoding

You can serialize message into JSON format:

```go
message := iso8583.NewMessage(spec)
message.MTI("0100")
message.Field(2, "4242424242424242")
message.Field(3, "123456")
message.Field(4, "100")

jsonMessage, err := json.Marshal(message)
```

it will produce following JSON:

```json
{
   "0":"0100",
   "1":"700000000000000000000000000000000000000000000000",
   "2":"4242424242424242",
   "3":123456,
   "4":"100"
}
```

### Network Header

All messages between the client/server (ISO host and endpoint) have a message
length header. It can be a 4 bytes ASCII or 2 bytes BCD encoded length. We
provide a `network.Header` interface to simplify the reading and writing of the
network header.

Following network headers are supported:

* Binary2Bytes - message length encoded in 2 bytes, e.g, {0x00 0x73} for 115
  bytes of the message
* ASCII4Bytes - message length encoded in 4 bytes ASCII, e.g., 0115 for 115
  bytes of the message
* BCD2Bytes - message length encoded in 2 bytes BCD, e.g, {0x01, 0x15} for 115
  bytes of the message
* VMLH (Visa Message Length Header) - message length encoded in 2 bytes + 2 reserved bytes

You can read network header from the network connection like this:

```go
header := network.NewBCD2BytesHeader()
_, err := header.ReadFrom(conn)
if err != nil {
	// handle error
}

// Make a buffer to hold message
buf := make([]byte, header.Length())
// Read the incoming message into the buffer.
read, err := io.ReadFull(conn, buf)
if err != nil {
	// handle error
}
if reqLen != header.Length() {
	// handle error
}

message := iso8583.NewMessage(specs.Spec87ASCII)
message.Unpack(buf)
```

Here is an example of how to write network header into network connection:

```go
header := network.NewBCD2BytesHeader()
packed, err := message.Pack()
if err != nil {
	// handle error
}
header.SetLength(len(packed))
_, err = header.WriteTo(conn)
if err != nil {
	// handle error
}
n, err := conn.Write(packed)
if err != nil {
	// handle error
}
```
