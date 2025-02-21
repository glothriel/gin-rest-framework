package detectors

import (
	"database/sql"
	"testing"

	"github.com/glothriel/grf/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/datatypes"
)

func testSqlNullModelsToRepresentation[Model any](t *testing.T, value any, expected any) {
	// given
	detector := DefaultToRepresentationDetector[Model]()

	// when
	toRepresentation, err := detector.ToRepresentation("data")

	// then
	assert.NoError(t, err)
	assert.NotNil(t, toRepresentation)

	// when
	repr, reprErr := toRepresentation(map[string]any{"data": value}, "data", nil)

	// then
	assert.NoError(t, reprErr)
	assert.Equal(t, expected, repr)

	// when
	_, reprErr = toRepresentation(map[string]any{"data": someStruct{}}, "data", nil)

	// then
	assert.Error(t, reprErr)
}

func TestToRepresentation_SQLNullInt16(t *testing.T) {
	type int16Model struct {
		Data sql.NullInt16 `json:"data"`
	}

	testSqlNullModelsToRepresentation[int16Model](
		t,
		sql.NullInt16{
			Int16: 42,
			Valid: true,
		},
		int16(42),
	)
	testSqlNullModelsToRepresentation[int16Model](t, sql.NullInt16{
		Int16: 0,
		Valid: false,
	}, nil)
}
func TestToRepresentation_SQLNullInt32(t *testing.T) {
	type int32Model struct {
		Data sql.NullInt32 `json:"data"`
	}

	testSqlNullModelsToRepresentation[int32Model](
		t,
		sql.NullInt32{
			Int32: 42,
			Valid: true,
		},
		int32(42),
	)
	testSqlNullModelsToRepresentation[int32Model](t, sql.NullInt32{
		Int32: 0,
		Valid: false,
	}, nil)
}

func TestToRepresentation_SQLNullInt64(t *testing.T) {
	type int64Model struct {
		Data sql.NullInt64 `json:"data"`
	}

	testSqlNullModelsToRepresentation[int64Model](
		t,
		sql.NullInt64{
			Int64: 42,
			Valid: true,
		},
		int64(42),
	)
	testSqlNullModelsToRepresentation[int64Model](t, sql.NullInt64{
		Int64: 0,
		Valid: false,
	}, nil)
}

func TestToRepresentation_SQLNullFloat64(t *testing.T) {
	type float64Model struct {
		Data sql.NullFloat64 `json:"data"`
	}

	testSqlNullModelsToRepresentation[float64Model](
		t,
		sql.NullFloat64{
			Float64: 3.14,
			Valid:   true,
		},
		float64(3.14),
	)
	testSqlNullModelsToRepresentation[float64Model](t, sql.NullFloat64{
		Float64: 0,
		Valid:   false,
	}, nil)
}

func TestToRepresentation_SQLNullString(t *testing.T) {
	type stringModel struct {
		Data sql.NullString `json:"data"`
	}

	testSqlNullModelsToRepresentation[stringModel](
		t,
		sql.NullString{
			String: "John",
			Valid:  true,
		},
		"John",
	)
	testSqlNullModelsToRepresentation[stringModel](t, sql.NullString{
		String: "",
		Valid:  false,
	}, nil)
}

func TestToRepresentation_SQLNullByte(t *testing.T) {
	type byteModel struct {
		Data sql.NullByte `json:"data"`
	}

	testSqlNullModelsToRepresentation[byteModel](
		t,
		sql.NullByte{
			Byte:  52,
			Valid: true,
		},
		"4",
	)
	testSqlNullModelsToRepresentation[byteModel](t, sql.NullByte{
		Byte:  0,
		Valid: false,
	}, nil)
}

func TestToRepresentation_SQLNullBool(t *testing.T) {
	type boolModel struct {
		Data sql.NullBool `json:"data"`
	}

	testSqlNullModelsToRepresentation[boolModel](
		t,
		sql.NullBool{
			Bool:  true,
			Valid: true,
		},
		true,
	)
	testSqlNullModelsToRepresentation[boolModel](t, sql.NullBool{
		Bool:  false,
		Valid: false,
	}, nil)
}

func TestToRepresentationGormDatatypesJSON(t *testing.T) {
	// given
	type gormDatatypesJSONModel struct {
		Data datatypes.JSON `json:"data"`
	}
	detector := &gormDataTypesJSONToRepresentationProvider[gormDatatypesJSONModel]{}

	// when
	toRepresentation, toRepresentationErr := detector.ToRepresentation("data")
	representation, representationErr := toRepresentation(models.InternalValue{
		"value": datatypes.JSON([]byte(`{"key": "value"}`)),
	}, "value", nil)

	// then
	assert.NoError(t, toRepresentationErr)
	assert.NoError(t, representationErr)
	assert.Equal(t, map[string]any{"key": "value"}, representation)

}

func TestToRepresentationGormDatatypesNotDatatypesJSON(t *testing.T) {
	// given
	type gormDatatypesJSONModel struct {
		Data datatypes.JSON `json:"data"`
	}
	detector := &gormDataTypesJSONToRepresentationProvider[gormDatatypesJSONModel]{}

	// when
	toRepresentation, toRepresentationErr := detector.ToRepresentation("data")
	_, representationErr := toRepresentation(models.InternalValue{
		"value": "suprajs bijacz",
	}, "value", nil)

	// then
	assert.NoError(t, toRepresentationErr)
	assert.Error(t, representationErr)

}

func TestToRepresentationGormDatatypesNotValidJSON(t *testing.T) {
	// given
	type gormDatatypesJSONModel struct {
		Data datatypes.JSON `json:"data"`
	}
	detector := &gormDataTypesJSONToRepresentationProvider[gormDatatypesJSONModel]{}

	// when
	toRepresentation, toRepresentationErr := detector.ToRepresentation("data")
	_, representationErr := toRepresentation(models.InternalValue{
		"value": datatypes.JSON([]byte(`{"key": "value",}`)),
	}, "value", nil)

	// then
	assert.NoError(t, toRepresentationErr)
	assert.Error(t, representationErr)

}
