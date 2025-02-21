package detectors

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

type someStruct struct{}

func testSqlNullModelsToInternalValue[Model any](t *testing.T, value any, expected any) {
	// given
	detector := DefaultToInternalValueDetector[Model]()

	// when
	internalValue, err := detector.ToInternalValue("data")

	// then
	assert.NoError(t, err)
	assert.NotNil(t, internalValue)

	// when
	iv, ivErr := internalValue(map[string]any{"data": value}, "data", nil)

	// then
	assert.NoError(t, ivErr)
	assert.Equal(t, expected, iv)

	// when
	_, ivErr = internalValue(map[string]any{"data": someStruct{}}, "data", nil)

	// then
	assert.Error(t, ivErr)
}

func TestToInternalValue_SQLNullInt16(t *testing.T) {
	type int16Model struct {
		Data sql.NullInt16 `json:"data"`
	}

	testSqlNullModelsToInternalValue[int16Model](t, float64(42), sql.NullInt16{
		Int16: 42,
		Valid: true,
	})
	testSqlNullModelsToInternalValue[int16Model](t, nil, sql.NullInt16{
		Int16: 0,
		Valid: false,
	})
}

func TestToInternalValue_SQLNullInt32(t *testing.T) {
	type int32Model struct {
		Data sql.NullInt32 `json:"data"`
	}

	testSqlNullModelsToInternalValue[int32Model](t, float64(42), sql.NullInt32{
		Int32: 42,
		Valid: true,
	})
	testSqlNullModelsToInternalValue[int32Model](t, nil, sql.NullInt32{
		Int32: 0,
		Valid: false,
	})
}

func TestToInternalValue_SQLNullInt64(t *testing.T) {
	type int64Model struct {
		Data sql.NullInt64 `json:"data"`
	}

	testSqlNullModelsToInternalValue[int64Model](t, float64(42), sql.NullInt64{
		Int64: 42,
		Valid: true,
	})
	testSqlNullModelsToInternalValue[int64Model](t, nil, sql.NullInt64{
		Int64: 0,
		Valid: false,
	})
}

func TestToInternalValue_SQLNullFloat64(t *testing.T) {
	type float64Model struct {
		Data sql.NullFloat64 `json:"data"`
	}

	testSqlNullModelsToInternalValue[float64Model](t, float64(42), sql.NullFloat64{
		Float64: float64(42),
		Valid:   true,
	})
	testSqlNullModelsToInternalValue[float64Model](t, nil, sql.NullFloat64{
		Float64: 0,
		Valid:   false,
	})
}

func TestToInternalValue_SQLNullString(t *testing.T) {
	type stringModel struct {
		Data sql.NullString `json:"data"`
	}

	testSqlNullModelsToInternalValue[stringModel](t, "42", sql.NullString{
		String: "42",
		Valid:  true,
	})
	testSqlNullModelsToInternalValue[stringModel](t, nil, sql.NullString{
		String: "",
		Valid:  false,
	})
}

func TestToInternalValue_SQLNullBool(t *testing.T) {
	type boolModel struct {
		Data sql.NullBool `json:"data"`
	}

	testSqlNullModelsToInternalValue[boolModel](t, true, sql.NullBool{
		Bool:  true,
		Valid: true,
	})
	testSqlNullModelsToInternalValue[boolModel](t, false, sql.NullBool{
		Bool:  false,
		Valid: true,
	})
	testSqlNullModelsToInternalValue[boolModel](t, nil, sql.NullBool{
		Bool:  false,
		Valid: false,
	})
}

func TestToInternalValue_SQLNullByte(t *testing.T) {
	type byteModel struct {
		Data sql.NullByte `json:"data"`
	}

	testSqlNullModelsToInternalValue[byteModel](t, "4", sql.NullByte{
		Byte:  byte(52), // ASCII code for '4'
		Valid: true,
	})
	testSqlNullModelsToInternalValue[byteModel](t, nil, sql.NullByte{
		Byte:  byte(0),
		Valid: false,
	})
}

type someUnmarshaler struct {
	T string
}

func (s *someUnmarshaler) UnmarshalText(text []byte) error {
	s.T = string(text)
	return nil
}

func TestToInternalValue_TextUnmarshaler(t *testing.T) {

	type textUnmarshalerModel struct {
		Data someUnmarshaler `json:"data"`
	}

	testSqlNullModelsToInternalValue[textUnmarshalerModel](
		t, "42", &someUnmarshaler{T: "42"},
	)

}

func TestToInternalValue_String(t *testing.T) {
	type stringModel struct {
		Data string `json:"data"`
	}

	testSqlNullModelsToInternalValue[stringModel](t, "42", "42")
}
func TestToInternalValue_Int8(t *testing.T) {
	type int8Model struct {
		Data int8 `json:"data"`
	}

	testSqlNullModelsToInternalValue[int8Model](t, float64(42), int8(42))
}

func TestToInternalValue_Int16(t *testing.T) {
	type int16Model struct {
		Data int16 `json:"data"`
	}

	testSqlNullModelsToInternalValue[int16Model](t, float64(42), int16(42))
}

func TestToInternalValue_Int32(t *testing.T) {
	type int32Model struct {
		Data int32 `json:"data"`
	}

	testSqlNullModelsToInternalValue[int32Model](t, float64(42), int32(42))
}

func TestToInternalValue_Int64(t *testing.T) {
	type int64Model struct {
		Data int64 `json:"data"`
	}

	testSqlNullModelsToInternalValue[int64Model](t, float64(42), int64(42))
}

func TestToInternalValue_Int(t *testing.T) {
	type intModel struct {
		Data int `json:"data"`
	}

	testSqlNullModelsToInternalValue[intModel](t, float64(42), int(42))
}

func TestToInternalValue_Uint8(t *testing.T) {
	type uint8Model struct {
		Data uint8 `json:"data"`
	}

	testSqlNullModelsToInternalValue[uint8Model](t, float64(42), uint8(42))
}

func TestToInternalValue_Uint16(t *testing.T) {
	type uint16Model struct {
		Data uint16 `json:"data"`
	}

	testSqlNullModelsToInternalValue[uint16Model](t, float64(42), uint16(42))
}

func TestToInternalValue_Uint32(t *testing.T) {
	type uint32Model struct {
		Data uint32 `json:"data"`
	}

	testSqlNullModelsToInternalValue[uint32Model](t, float64(42), uint32(42))
}

func TestToInternalValue_Uint64(t *testing.T) {
	type uint64Model struct {
		Data uint64 `json:"data"`
	}

	testSqlNullModelsToInternalValue[uint64Model](t, float64(42), uint64(42))
}

func TestToInternalValueDataTypesJSON(t *testing.T) {
	// given
	type jsonModel struct {
		Data datatypes.JSON `json:"data"`
	}
	detector := DefaultToInternalValueDetector[jsonModel]()

	// when
	internalValueFunc, internalValieFuncErr := detector.ToInternalValue("data")
	iv, ivErr := internalValueFunc(
		map[string]any{"data": map[string]any{"key": "value"}},
		"data",
		nil,
	)

	// then
	assert.NoError(t, internalValieFuncErr)
	assert.NoError(t, ivErr)
	assert.Equal(t, datatypes.JSON(`{"key":"value"}`), iv)
}

func TestToInternalValueDataTypesJSONNotJSONSerializable(t *testing.T) {
	// given
	type jsonModel struct {
		Data datatypes.JSON `json:"data"`
	}
	detector := DefaultToInternalValueDetector[jsonModel]()

	// when
	internalValueFunc, internalValueFuncErr := detector.ToInternalValue("data")
	_, ivErr := internalValueFunc(
		map[string]any{"data": make(chan int)},
		"data",
		nil,
	)

	// then
	assert.NoError(t, internalValueFuncErr)
	assert.Error(t, ivErr)
}
