package detectors

import (
	"database/sql"
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"github.com/glothriel/grf/pkg/types"
	"gorm.io/datatypes"
)

// ToRepresentationDetector is an interface that allows to detect the representation function for a given field
// using well known interfaces, knowledge of stdlib types or any other way
type ToRepresentationDetector[Model any] interface {
	ToRepresentation(fieldName string) (fields.RepresentationFunc, error)
}

func DefaultToRepresentationDetector[Model any]() ToRepresentationDetector[Model] {
	return &missingFieldSkippingToRepresentationDetector[Model]{
		child: &relationshipDetector[Model]{
			representationChild: &chainingToRepresentationDetector[Model]{
				children: []ToRepresentationDetector[Model]{
					&usingGRFRepresentableToRepresentationProvider[Model]{},
					&timeTimeToRepresentationProvider[Model]{},
					&gormDataTypesJSONToRepresentationProvider[Model]{},
					&fromTypeMapperToRepresentationProvider[Model]{
						mapper:         types.Mapper(),
						modelTypeNames: FieldTypes[Model](),
					},
					&encodingTextMarshalerToRepresentationProvider[Model]{},
					&usingSqlNullFieldToRepresentationProvider[Model, sql.NullBool]{
						valueFunc: func(v sql.NullBool) any {
							if v.Valid {
								return v.Bool
							}
							return nil
						},
					},
					&usingSqlNullFieldToRepresentationProvider[Model, sql.NullInt16]{
						valueFunc: func(v sql.NullInt16) any {
							if v.Valid {
								return v.Int16
							}
							return nil
						},
					},
					&usingSqlNullFieldToRepresentationProvider[Model, sql.NullInt32]{
						valueFunc: func(v sql.NullInt32) any {
							if v.Valid {
								return v.Int32
							}
							return nil
						},
					},
					&usingSqlNullFieldToRepresentationProvider[Model, sql.NullInt64]{
						valueFunc: func(v sql.NullInt64) any {
							if v.Valid {
								return v.Int64
							}
							return nil
						},
					},
					&usingSqlNullFieldToRepresentationProvider[Model, sql.NullString]{
						valueFunc: func(v sql.NullString) any {
							if v.Valid {
								return v.String
							}
							return nil
						},
					},
					&usingSqlNullFieldToRepresentationProvider[Model, sql.NullFloat64]{
						valueFunc: func(v sql.NullFloat64) any {
							if v.Valid {
								return v.Float64
							}
							return nil
						},
					},
					&usingSqlNullFieldToRepresentationProvider[Model, sql.NullByte]{
						valueFunc: func(v sql.NullByte) any {
							if v.Valid {
								return string(v.Byte)
							}
							return nil
						},
					},
				},
			},
		},
	}
}

type missingFieldSkippingToRepresentationDetector[Model any] struct {
	child ToRepresentationDetector[Model]
}

func (p missingFieldSkippingToRepresentationDetector[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	childDetector, childErr := p.child.ToRepresentation(fieldName)
	if childErr != nil {
		return nil, childErr
	}
	return func(m models.InternalValue, s string, c *gin.Context) (any, error) {
		if _, ok := m[s]; !ok {
			return nil, fields.NewErrorFieldIsNotPresentInPayload(s)
		}
		return childDetector(m, s, c)
	}, nil
}

type usingGRFRepresentableToRepresentationProvider[Model any] struct{}

func (p usingGRFRepresentableToRepresentationProvider[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings != nil && fieldSettings.isGRFRepresentable {
		return ConvertFuncToRepresentationFuncAdapter(
			func(v any) (any, error) {
				return v.(fields.GRFRepresentable).ToRepresentation()
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a GRFRepresentable", fieldName)
}

type timeTimeToRepresentationProvider[Model any] struct{}

func (p timeTimeToRepresentationProvider[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {

	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.itsType.Name() == "Time" && fieldSettings.itsType.PkgPath() == "time" {
		return ConvertFuncToRepresentationFuncAdapter(
			func(v any) (any, error) {
				vAsTime, ok := v.(time.Time)
				if ok {
					return vAsTime.Format("2006-01-02T15:04:05Z"), nil
				}
				return nil, fmt.Errorf("Field `%s` is not a time.Time", fieldName)
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a time.Time", fieldName)
}

type usingSqlNullFieldToRepresentationProvider[Model any, sqlNullType any] struct {
	valueFunc func(sqlNullType) any
}

func (p usingSqlNullFieldToRepresentationProvider[Model, sqlNullType]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	isNullSqlType := false
	var entity Model
	for _, field := range reflect.VisibleFields(reflect.TypeOf(entity)) {
		jsonTag := field.Tag.Get("json")
		if jsonTag == fieldName {
			var theTypeAsAny any
			reflectedInstance := reflect.New(reflect.TypeOf(reflect.ValueOf(entity).FieldByName(field.Name).Interface())).Elem()
			if reflectedInstance.CanAddr() {
				theTypeAsAny = reflectedInstance.Addr().Interface()
			} else {
				theTypeAsAny = reflectedInstance.Interface()
			}
			_, isNullSqlType = theTypeAsAny.(*sqlNullType)
			if isNullSqlType {
				break
			}
		}
	}
	var zeroType sqlNullType
	if isNullSqlType {
		return ConvertFuncToRepresentationFuncAdapter(
			func(v any) (any, error) {
				vAsNullType, ok := v.(sqlNullType)
				if !ok {
					return nil, fmt.Errorf("Field `%s` is not a %T", fieldName, zeroType)
				}
				return p.valueFunc(vAsNullType), nil
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a sql.NullType", fieldName)
}

// fromTypeMapperToRepresentationDetector uses field type mapper to detect the representation function for a given field
type fromTypeMapperToRepresentationProvider[Model any] struct {
	mapper         *types.FieldTypeMapper
	modelTypeNames map[string]string
}

func (p fromTypeMapperToRepresentationProvider[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	ftmToRepresentation, toRepresentationErr := p.mapper.ToRepresentation(p.modelTypeNames[fieldName])
	if toRepresentationErr != nil {
		return nil, toRepresentationErr
	}
	return ConvertFuncToRepresentationFuncAdapter(ftmToRepresentation), nil
}

type encodingTextMarshalerToRepresentationProvider[Model any] struct{}

func (p encodingTextMarshalerToRepresentationProvider[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings != nil && fieldSettings.isEncodingTextMarshaler {
		return ConvertFuncToRepresentationFuncAdapter(
			func(v any) (any, error) {
				marshalledBytes, marshallErr := v.(encoding.TextMarshaler).MarshalText()
				if marshallErr != nil {
					return nil, marshallErr
				}
				return string(marshalledBytes), nil
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a encoding.TextMarshaler", fieldName)
}

type gormDataTypesJSONToRepresentationProvider[Model any] struct{}

func (p gormDataTypesJSONToRepresentationProvider[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings != nil && fieldSettings.isDataTypesJSON {
		return ConvertFuncToRepresentationFuncAdapter(
			func(v any) (any, error) {
				vAsJSON, ok := v.(datatypes.JSON)
				if !ok {
					return nil, fmt.Errorf("Field `%s` is not a datatypes.JSON", fieldName)
				}
				rawJSON, marshalJSONErr := vAsJSON.MarshalJSON()
				var ret any
				unmarshalErr := json.Unmarshal(rawJSON, &ret)
				if unmarshalErr != nil || marshalJSONErr != nil {
					return nil, fmt.Errorf(
						"Failed to unmarshal field `%s` from JSON: %w, %w",
						fieldName,
						marshalJSONErr,
						unmarshalErr,
					)
				}
				return ret, nil
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a encoding.TextMarshaler", fieldName)
}

type chainingToRepresentationDetector[Model any] struct {
	children []ToRepresentationDetector[Model]
}

func (p chainingToRepresentationDetector[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	for _, child := range p.children {
		representation, representationErr := child.ToRepresentation(fieldName)
		if representationErr == nil {
			return representation, nil
		}
	}
	var m Model
	return nil, fmt.Errorf("No representation function could be found for field `%T`.`%s`", m, fieldName)
}
