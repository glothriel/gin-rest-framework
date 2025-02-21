package integration

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/models"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSQLite(t *testing.T) {
	DoTestTypes(t, sqlite.Open(":memory:"))
}

type BoolModel struct {
	models.BaseModel
	Value bool `json:"value" gorm:"column:value"`
}

type StringModel struct {
	models.BaseModel
	Value string `json:"value" gorm:"column:value"`
}

type IntModel struct {
	models.BaseModel
	Value int `json:"value" gorm:"column:value"`
}

type UintModel struct {
	models.BaseModel
	Value uint `json:"value" gorm:"column:value"`
}

type FloatModel struct {
	models.BaseModel
	Value float64 `json:"value" gorm:"column:value"`
}

type StringSliceModel struct {
	models.BaseModel
	Value models.SliceField[string] `json:"value" gorm:"column:value;type:json"`
}

type FloatSliceModel struct {
	models.BaseModel
	Value models.SliceField[float64] `json:"value" gorm:"column:value;type:json"`
}

type TimeModel struct {
	models.BaseModel
	Value time.Time `json:"value" gorm:"column:value;type:timestamp"`
}

type BoolSliceModel struct {
	models.BaseModel
	Value models.SliceField[bool] `json:"value" gorm:"column:value;type:json"`
}

type AnySliceModel struct {
	models.BaseModel
	Value models.SliceField[any] `json:"value" gorm:"column:value;type:json"`
}

type NullBoolModel struct {
	models.BaseModel
	Value sql.NullBool `json:"value" gorm:"column:value"`
}

type NullStringModel struct {
	models.BaseModel
	Value sql.NullString `json:"value" gorm:"column:value"`
}

type DecimalModel struct {
	models.BaseModel
	Value decimal.Decimal `json:"value" gorm:"column:value"`
}

type NullFloat64Model struct {
	models.BaseModel
	Value sql.NullFloat64 `json:"value" gorm:"column:value"`
}

type NestedJSONModel struct {
	models.BaseModel
	Value datatypes.JSON `json:"value" gorm:"column:value"`
}

func DoTestTypes(t *testing.T, dialector gorm.Dialector) { // nolint: funlen
	tests := []struct {
		name        string
		baseURL     string
		skipOn      map[string]bool
		okBodies    []map[string]any
		errorBodies []map[string]any
		okResponses []map[string]any
		router      func() *gin.Engine
	}{
		{
			name:    "Bool type",
			baseURL: "/bool_field",
			okBodies: []map[string]any{
				{"value": true},
				{"value": false},
			},
			okResponses: []map[string]any{
				{"value": true},
				{"value": false},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": "hueble"},
				{"value": "true"},
				{"value": "True"},
			},
			router: func() *gin.Engine {
				return registerModel[BoolModel]("/bool_field", dialector)
			},
		},
		{
			name:    "String type",
			baseURL: "/string_field",
			okBodies: []map[string]any{
				{"value": "hello world"},
				{"value": ""},
			},
			okResponses: []map[string]any{
				{"value": "hello world"},
				{"value": ""},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
			},
			router: func() *gin.Engine {
				return registerModel[StringModel]("/string_field", dialector)
			},
		},
		{
			name:    "Integer type",
			baseURL: "/int_field",
			okBodies: []map[string]any{
				{"value": 1337},
				{"value": -1337},
				{"value": 13.00},
			},
			okResponses: []map[string]any{
				{"value": 1337.0},
				{"value": -1337.0},
				{"value": 13.0},
			},
			errorBodies: []map[string]any{
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
				{"value": "hello world"},
			},
			router: func() *gin.Engine {
				return registerModel[IntModel]("/int_field", dialector)
			},
		},
		{
			name:    "Unsigned Integer type",
			baseURL: "/uint_field",
			okBodies: []map[string]any{
				{"value": 1337},
				{"value": 13.00},
			},
			okResponses: []map[string]any{
				{"value": 1337.0},
				{"value": 13.0},
			},
			errorBodies: []map[string]any{
				{"value": -1337},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
				{"value": "hello world"},
			},
			router: func() *gin.Engine {
				return registerModel[UintModel]("/uint_field", dialector)
			},
		},
		{
			name:    "Float type",
			baseURL: "/float_field",
			okBodies: []map[string]any{
				{"value": 1.337},
				{"value": 1337},
			},
			okResponses: []map[string]any{
				{"value": 1.337},
				{"value": 1337.0},
			},
			errorBodies: []map[string]any{
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
				{"value": "hello world"},
			},
			router: func() *gin.Engine {
				return registerModel[FloatModel]("/float_field", dialector)
			},
		},
		{
			name:    "String Slice type",
			baseURL: "/string_slice_field",
			okBodies: []map[string]any{
				{"value": []string{"hello world"}},
			},
			okResponses: []map[string]any{
				{"value": []any{"hello world"}},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []interface{}{"1", 2}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
			},
			router: func() *gin.Engine {
				return registerModel[StringSliceModel]("/string_slice_field", dialector)
			},
		},
		{
			name:    "Float Slice type",
			baseURL: "/float_slice_field",
			okBodies: []map[string]any{
				{"value": []float64{1.337}},
				{"value": []int{1, 2, 3}},
			},
			okResponses: []map[string]any{
				{"value": []any{1.337}},
				{"value": []any{1.0, 2.0, 3.0}},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []interface{}{"1", 2}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": false},
			},
			router: func() *gin.Engine {
				return registerModel[FloatSliceModel]("/float_slice_field", dialector)
			},
		},
		{
			name:    "time.Time type",
			baseURL: "/time_field",
			okBodies: []map[string]any{
				{"value": "2021-01-01T00:00:00Z"},
				{"value": "2021-01-01T00:00:00+00:00"},
			},
			okResponses: []map[string]any{
				{"value": "2021-01-01T00:00:00Z"},
				{"value": "2021-01-01T00:00:00Z"},
			},
			errorBodies: []map[string]any{
				{"value": "2021-01-01"},
				{"value": 2021},
				{"value": "01-01-2021T00:00:00Z"},
				{"value": "2021-13-01T00:00:00Z"},
				{"value": "2021-01-32T00:00:00Z"},
				{"value": "2021-01-01T24:00:00Z"},
				{"value": "2021-01-01T00:60:00Z"},
				{"value": "2021-01-01T00:00:60Z"},
				{"value": "2021-01-01T00:00"},
				{"value": "2021-01-01 00:00:00Z"},
				{"value": "2021-01-01T00:00:00"},
				{"value": "not-a-date"},
				{"value": ""},
				{"value": nil},
			},
			router: func() *gin.Engine {
				return registerModel[TimeModel]("/time_field", dialector)
			},
		},
		{
			name:    "Bool Slice type",
			baseURL: "/bool_slice_field",
			okBodies: []map[string]any{
				{"value": []bool{true}},
				{"value": []bool{false}},
			},
			okResponses: []map[string]any{
				{"value": []any{true}},
				{"value": []any{false}},
			},
			errorBodies: []map[string]any{
				{"value": []int{1, 2, 3}},
				{"value": []interface{}{"1", 2}},
				{"value": []string{"true", "false"}},
				{"value": map[string]string{"foo": "bar"}},
				{"value": true},
				{"value": false},
			},
			router: func() *gin.Engine {
				return registerModel[BoolSliceModel]("/bool_slice_field", dialector)
			},
		},
		{
			name:    "Any Slice type",
			baseURL: "/any_slice_field",
			okBodies: []map[string]any{
				{"value": []interface{}{true, "1", 2, 3.45}},
				{"value": []int{1, 2, 3}},
				{"value": []interface{}{
					map[string]string{"foo": "bar"},
					map[string]string{"bar": "baz"},
					1,
					true,
				}},
				{"value": []string{"true", "false"}},
			},
			okResponses: []map[string]any{
				{"value": []interface{}{true, "1", 2.0, 3.45}},
				{"value": []interface{}{1.0, 2.0, 3.0}},
				{"value": []interface{}{
					any(map[string]any{"foo": "bar"}),
					any(map[string]any{"bar": "baz"}),
					1.0,
					true,
				}},
				{"value": []interface{}{"true", "false"}},
			},
			errorBodies: []map[string]any{
				{"value": map[string]any{"foo": "bar"}},
				{"value": true},
				{"value": false},
				{"value": "ads"},
				{"value": 1},
				{"value": 1.23},
			},
			router: func() *gin.Engine {
				return registerModel[AnySliceModel]("/any_slice_field", dialector)
			},
		},
		{
			name:    "Nullable Bool type",
			baseURL: "/null_bool_field",
			okBodies: []map[string]any{
				{"value": true},
				{"value": nil},
			},
			okResponses: []map[string]any{
				{"value": true},
				{"value": nil},
			},
			router: func() *gin.Engine {
				return registerModel[NullBoolModel]("/null_bool_field", dialector)
			},
		},
		{
			name:    "Nullable String type",
			baseURL: "/null_string_field",
			okBodies: []map[string]any{
				{"value": "hello world"},
				{"value": nil},
			},
			okResponses: []map[string]any{
				{"value": "hello world"},
				{"value": nil},
			},
			router: func() *gin.Engine {
				return registerModel[NullStringModel]("/null_string_field", dialector)
			},
		},
		{
			name:    "Decimal type",
			baseURL: "/decimal_field",
			okBodies: []map[string]any{
				{"value": "1.337"},
				{"value": "1337"},
				{"value": "13.37"},
				{"value": "0"},
			},
			okResponses: []map[string]any{
				{"value": "1.337"},
				{"value": "1337"},
				{"value": "13.37"},
				{"value": "0"},
			},
			errorBodies: []map[string]any{
				{"value": 1},
				{"value": 1.337},
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": "hello world"},
				{"value": "1,337"},
				{"value": "1.3.37"},
			},
			router: func() *gin.Engine {
				return registerModel[DecimalModel]("/decimal_field", dialector)
			},
		},
		{
			name:    "Nullable Float type",
			baseURL: "/null_float64_field",
			okBodies: []map[string]any{
				{"value": 1.337},
				{"value": nil},
			},
			okResponses: []map[string]any{
				{"value": 1.337},
				{"value": nil},
			},
			router: func() *gin.Engine {
				return registerModel[NullFloat64Model]("/null_float64_field", dialector)
			},
		},
		{
			name:    "Nested JSON field",
			baseURL: "/nested_json_field",
			okBodies: []map[string]any{
				{"value": map[string]string{"foo": "bar"}},
				{"value": []int{1, 2, 3}},
				{"value": []bool{true, false}},
				{"value": true},
				{"value": "hello world"},
				{"value": "1,337"},
				{"value": "1.3.37"},
			},
			okResponses: []map[string]any{
				{"value": map[string]any{"foo": "bar"}},
				{"value": []any{1.0, 2.0, 3.0}},
				{"value": []any{true, false}},
				{"value": true},
				{"value": "hello world"},
				{"value": "1,337"},
				{"value": "1.3.37"},
			},
			router: func() *gin.Engine {
				return registerModel[NestedJSONModel]("/nested_json_field", dialector)
			},
		},
		{
			name:    "Nested JSON field numeric (only postgres supports them)",
			baseURL: "/nested_json_field",
			skipOn: map[string]bool{
				"sqlite": true,
			},
			okBodies: []map[string]any{
				{"value": 1},
				{"value": 13.37},
			},
			okResponses: []map[string]any{
				{"value": 1.0},
				{"value": 13.37},
			},
			router: func() *gin.Engine {
				return registerModel[NestedJSONModel]("/nested_json_field", dialector)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if len(tt.okBodies) != len(tt.okResponses) {
				t.Fatalf("Test case %s: number of bodies and responses should be equal", tt.name)
			}

			skip, ok := tt.skipOn[dialector.Name()]
			if ok && skip {
				t.Skipf("Skipping test %s for dialector %s", tt.name, dialector.Name())
			}

			for _, errorBody := range tt.errorBodies {
				newRequestTestCase(
					t,
					fmt.Sprintf("%s error: %v", tt.name, errorBody),
				).Req(
					newRequest("POST", tt.baseURL, errorBody),
				).ExCode(
					http.StatusBadRequest,
				).Run(tt.router())
			}

			for i, okBody := range tt.okBodies {
				router := tt.router()
				resourceID := newRequestTestCase(
					t,
					fmt.Sprintf("%s create", tt.name),
				).Req(
					newRequest("POST", tt.baseURL, okBody),
				).ExCode(
					http.StatusCreated,
				).ExJson(
					tt.okResponses[i],
				).Run(router)

				newRequestTestCase(
					t,
					tt.name,
				).Req(
					newRequest("GET", tt.baseURL, nil),
				).ExCode(
					http.StatusOK,
				).ExJson(
					[]any{tt.okResponses[i]},
				).Run(router)

				newRequestTestCase(
					t,
					tt.name,
				).Req(
					newRequest("GET", fmt.Sprintf("%s/%s", tt.baseURL, resourceID), nil),
				).ExCode(
					http.StatusOK,
				).ExJson(
					tt.okResponses[i],
				).Run(router)

				newRequestTestCase(
					t,
					tt.name,
				).Req(
					newRequest("DELETE", fmt.Sprintf("%s/%s", tt.baseURL, resourceID), nil),
				).ExCode(
					http.StatusNoContent,
				).Run(router)

				newRequestTestCase(
					t,
					tt.name,
				).Req(
					newRequest("GET", tt.baseURL, nil),
				).ExCode(
					http.StatusOK,
				).ExJson(
					[]any{},
				).Run(router)
			}
		})
	}
}

type PostgresTestSuite struct {
	suite.Suite
	postgres *embeddedpostgres.EmbeddedPostgres
	DSN      string
}

func TestPostgres(t *testing.T) {
	suite.Run(t, new(PostgresTestSuite))
}

func (s *PostgresTestSuite) SetupSuite() {
	s.postgres = embeddedpostgres.NewDatabase()
	startErr := s.postgres.Start()
	if startErr == nil || strings.Contains(startErr.Error(), "process already listening") {
		return
	}
	require.NoError(s.T(), startErr)
}

func (s *PostgresTestSuite) TearDownSuite() {
	stopErr := s.postgres.Stop()
	if stopErr == nil || strings.Contains(stopErr.Error(), "server has not been started") {
		return
	}
	require.NoError(s.T(), stopErr)
}

func (s *PostgresTestSuite) SetupTest() {
	db, connectErr := sql.Open(
		"postgres", "host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable",
	)
	require.NoError(s.T(), connectErr)
	_, dropErr := db.Exec("DROP DATABASE IF EXISTS tests")
	require.NoError(s.T(), dropErr)
	_, createErr := db.Exec("CREATE DATABASE tests")
	require.NoError(s.T(), createErr)
	s.DSN = "host=localhost port=5432 user=postgres password=postgres dbname=tests sslmode=disable TimeZone=UTC"
}

func (s *PostgresTestSuite) TestPostgres() {
	DoTestTypes(s.T(), postgres.Open(s.DSN))
}
