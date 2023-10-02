package field

import (
	"reflect"
	"testing"

	"github.com/moov-io/iso8583/encoding"
	"github.com/moov-io/iso8583/padding"
	"github.com/moov-io/iso8583/prefix"
	"github.com/stretchr/testify/require"
)

func TestNumericField(t *testing.T) {
	spec := &Spec{
		Length:      10,
		Description: "Field",
		Enc:         encoding.ASCII,
		Pref:        prefix.ASCII.Fixed,
		Pad:         padding.Left(' '),
	}
	numeric := NewNumeric(spec)

	numeric.SetBytes([]byte("100"))
	require.Equal(t, 100, numeric.Value())

	packed, err := numeric.Pack()
	require.NoError(t, err)
	require.Equal(t, "       100", string(packed))

	length, err := numeric.Unpack([]byte("      9876"))
	require.NoError(t, err)
	require.Equal(t, 10, length)

	b, err := numeric.Bytes()
	require.NoError(t, err)
	require.Equal(t, "9876", string(b))

	require.Equal(t, 9876, numeric.Value())

	numeric = NewNumeric(spec)
	numeric.Marshal(NewNumericValue(9876))
	packed, err = numeric.Pack()
	require.NoError(t, err)
	require.Equal(t, "      9876", string(packed))

	numeric = NewNumeric(spec)
	data := NewNumericValue(0)
	numeric.Marshal(data)
	length, err = numeric.Unpack([]byte("      9876"))
	require.NoError(t, err)
	require.Equal(t, 10, length)
	require.Equal(t, 9876, numeric.Value())

	numeric = NewNumeric(spec)
	numeric.SetValue(9876)

	require.Equal(t, 9876, numeric.Value())
}

func TestNumericNil(t *testing.T) {
	var str *Numeric = nil

	bs, err := str.Bytes()
	require.NoError(t, err)
	require.Nil(t, bs)

	value, err := str.String()
	require.NoError(t, err)
	require.Equal(t, "", value)

	n := str.Value()
	require.Equal(t, 0, n)
}

func TestNumericPack(t *testing.T) {
	t.Run("returns error for zero value when fixed length and no padding specified", func(t *testing.T) {
		spec := &Spec{
			Length:      10,
			Description: "Field",
			Enc:         encoding.ASCII,
			Pref:        prefix.ASCII.Fixed,
		}
		numeric := NewNumeric(spec)
		_, err := numeric.Pack()

		// zero value for Numeric is 0, so we have default field length 1
		require.EqualError(t, err, "failed to encode length: field length: 1 should be fixed: 10")
	})
}

func TestNumericFieldUnmarshal(t *testing.T) {
	str := NewNumericValue(123456)

	val1 := &Numeric{}
	err := str.Unmarshal(val1)
	require.NoError(t, err)
	require.Equal(t, 123456, val1.Value())

	var val2 string
	err = str.Unmarshal(&val2)
	require.NoError(t, err)
	require.Equal(t, "123456", val2)

	var val3 int
	err = str.Unmarshal(&val3)
	require.NoError(t, err)
	require.Equal(t, 123456, val3)

	val4 := reflect.ValueOf(&val2).Elem()
	err = str.Unmarshal(val4)
	require.NoError(t, err)
	require.Equal(t, "123456", val4.String())

	val5 := reflect.ValueOf(&val3).Elem()
	err = str.Unmarshal(val5)
	require.NoError(t, err)
	require.Equal(t, 123456, int(val5.Int()))

	val6 := reflect.ValueOf(val2)
	err = str.Unmarshal(val6)
	require.Error(t, err)
	require.Equal(t, "cannot set reflect.Value of type string", err.Error())

	val7 := reflect.ValueOf(&val2)
	err = str.Unmarshal(val7)
	require.Error(t, err)
	require.Equal(t, "cannot set reflect.Value of type ptr", err.Error())

	err = str.Unmarshal(nil)
	require.Error(t, err)
	require.Equal(t, "unsupported type: expected *Numeric, *int, or reflect.Value, got <nil>", err.Error())
}

func TestNumericFieldMarshal(t *testing.T) {
	str := NewNumericValue(0)
	vNumeric := NewNumericValue(123456)
	str.Marshal(vNumeric)
	require.Equal(t, 123456, vNumeric.Value())

	str.Marshal(&vNumeric)
	require.Equal(t, 123456, vNumeric.Value())

	vstring := "123456"
	str.Marshal(vstring)
	require.Equal(t, 123456, vNumeric.Value())

	str.Marshal(&vstring)
	require.Equal(t, 123456, vNumeric.Value())

	vint := 123456
	str.Marshal(vint)
	require.Equal(t, 123456, vNumeric.Value())

	str.Marshal(&vint)
	require.Equal(t, 123456, vNumeric.Value())

	err := str.Marshal(nil)
	require.NoError(t, err)

	err = str.Marshal([]byte("123456"))
	require.Error(t, err)
	require.Equal(t, "data does not match require *Numeric or (int, *int, string, *string) type", err.Error())
}

func TestNumericFieldWithNotANumber(t *testing.T) {
	numeric := NewNumeric(&Spec{
		Length:      10,
		Description: "Field",
		Enc:         encoding.ASCII,
		Pref:        prefix.ASCII.Fixed,
		Pad:         padding.Left(' '),
	})

	err := numeric.SetBytes([]byte("hello"))
	require.Error(t, err)
	require.EqualError(t, err, "failed to convert into number")
	require.Equal(t, 0, numeric.Value())

	packed, err := numeric.Pack()
	require.NoError(t, err)
	require.Equal(t, "         0", string(packed))

	_, err = numeric.Unpack([]byte("hhhhhhhhhh"))
	require.Error(t, err)
	require.EqualError(t, err, "failed to set bytes: failed to convert into number")
}

func TestNumericFieldZeroLeftPaddedZero(t *testing.T) {
	numeric := NewNumeric(&Spec{
		Length:      4,
		Description: "Field",
		Enc:         encoding.ASCII,
		Pref:        prefix.ASCII.Fixed,
		Pad:         padding.Left('0'),
	})

	numeric.SetBytes([]byte("0"))
	require.Equal(t, 0, numeric.Value())

	packed, err := numeric.Pack()

	require.NoError(t, err)
	require.Equal(t, "0000", string(packed))

	length, err := numeric.Unpack([]byte("0000"))

	require.NoError(t, err)
	require.Equal(t, 4, length)

	bs, err := numeric.Bytes()
	require.NoError(t, err)
	require.Equal(t, "0", string(bs))

	require.Equal(t, 0, numeric.Value())
}

func TestNumericSetBytesSetsDataOntoDataStruct(t *testing.T) {
	numeric := NewNumeric(&Spec{
		Length:      1,
		Description: "Field",
		Enc:         encoding.ASCII,
		Pref:        prefix.ASCII.Fixed,
	})

	data := &Numeric{}
	err := numeric.Marshal(data)
	require.NoError(t, err)

	err = numeric.SetBytes([]byte("9"))
	require.NoError(t, err)

	require.Equal(t, 9, numeric.Value())
}

func TestNumericJSONMarshal(t *testing.T) {
	numeric := NewNumericValue(1)
	marshalledJSON, err := numeric.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, "1", string(marshalledJSON))
}

func TestNumericJSONUnmarshal(t *testing.T) {
	input := []byte(`4000`)

	numeric := NewNumeric(&Spec{
		Length:      4,
		Description: "Field",
		Enc:         encoding.ASCII,
		Pref:        prefix.ASCII.Fixed,
	})

	require.NoError(t, numeric.UnmarshalJSON(input))
	require.Equal(t, 4000, numeric.Value())
}
