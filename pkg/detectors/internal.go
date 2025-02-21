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
	"github.com/glothriel/grf/pkg/types"
	"gorm.io/datatypes"
)

type ToInternalValueDetector interface {
	ToInternalValue(fieldName string) (fields.InternalValueFunc, error)
}

type missingFieldSkippingToInternalValueDetector[Model any] struct {
	child ToInternalValueDetector
}

func (p *missingFieldSkippingToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	childDetector, childErr := p.child.ToInternalValue(fieldName)
	if childErr != nil {
		return nil, childErr
	}
	return func(m map[string]any, s string, c *gin.Context) (any, error) {
		if _, ok := m[s]; !ok {
			return nil, fields.NewErrorFieldIsNotPresentInPayload(s)
		}
		return childDetector(m, s, c)
	}, nil
}

type usingGRFParsableToInternalValueDetector[Model any] struct{}

func (p *usingGRFParsableToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.isGRFParsable {
		return ConvertFuncToInternalValueFuncAdapter(
			func(v any) (any, error) {
				typedV := reflect.New(fieldSettings.itsType).Interface()
				theErr := typedV.(fields.GRFParsable).FromRepresentation(v)
				return typedV, theErr
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a GRFParsable", fieldName)
}

type isoTimeTimeToInternalValueDetector[Model any] struct{}

func (p *isoTimeTimeToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.itsType.Name() == "Time" && fieldSettings.itsType.PkgPath() == "time" {
		return ConvertFuncToInternalValueFuncAdapter(
			func(v any) (any, error) {
				vStr, isString := v.(string)
				if isString {
					t, err := time.Parse(time.RFC3339, vStr)
					if err != nil {
						return nil, err
					}
					return t, nil
				}
				return nil, fmt.Errorf("Field `%s` is not a string", fieldName)
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a time.Time", fieldName)
}

type fromTypeMapperToInternalValueDetector[Model any] struct {
	mapper         *types.FieldTypeMapper
	modelTypeNames map[string]string
}

func (p *fromTypeMapperToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	ftmToInternalValue, toInternalValueErr := p.mapper.ToInternalValue(p.modelTypeNames[fieldName])
	if toInternalValueErr != nil {
		return nil, toInternalValueErr
	}
	return ConvertFuncToInternalValueFuncAdapter(ftmToInternalValue), nil
}

type encodingTextUnmarshalerToInternalValueDetector[Model any] struct{}

func (p *encodingTextUnmarshalerToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.isEncodingTextUnmarshaler {
		return ConvertFuncToInternalValueFuncAdapter(
			func(v any) (any, error) {
				var realFieldValue any = reflect.New(fieldSettings.itsType).Interface()
				vStr, ok := v.(string)
				if ok {
					fv, ok := realFieldValue.(encoding.TextUnmarshaler)
					if !ok {
						return nil, fmt.Errorf("Field `%s` is not a encoding.TextUnmarshaler", fieldName)
					}
					unmarshalErr := fv.UnmarshalText([]byte(vStr))
					return fv, unmarshalErr
				}
				return nil, fmt.Errorf("Field `%s` is not a string", fieldName)
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a encoding.TextUnmarshaler", fieldName)
}

type gormDataTypesJSONToInternalValueDetector[Model any] struct{}

func (p *gormDataTypesJSONToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.isDataTypesJSON {
		return ConvertFuncToInternalValueFuncAdapter(
			func(v any) (any, error) {
				vAsBytes, marshalErr := json.Marshal(v)
				var dtj datatypes.JSON
				unmarshalErr := dtj.UnmarshalJSON(vAsBytes)
				if unmarshalErr != nil || marshalErr != nil {
					return nil, fmt.Errorf(
						"Failed to unmarshal field `%s` from JSON: %w, %w",
						fieldName,
						marshalErr,
						unmarshalErr,
					)
				}
				return dtj, nil
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a encoding.TextUnmarshaler", fieldName)
}

type usingSqlNullFieldToInternalValueDetector[Model any, sqlNullType any] struct {
	valueFunc func(v any) (any, error)
}

func (p *usingSqlNullFieldToInternalValueDetector[Model, sqlNullType]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
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
		return ConvertFuncToInternalValueFuncAdapter(
			func(v any) (any, error) {
				if v == nil {
					return zeroType, nil
				}
				return p.valueFunc(v)
			},
		), nil
	}
	return nil, fmt.Errorf("Field `%s` is not a sql.NullInt32", fieldName)
}

type chainingToInternalValueDetector[Model any] struct {
	children []ToInternalValueDetector
}

func (p *chainingToInternalValueDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	for _, child := range p.children {
		internalValue, internalValueErr := child.ToInternalValue(fieldName)
		if internalValueErr == nil {
			return internalValue, nil
		}
	}
	var m Model
	return nil, fmt.Errorf("No internal value function could be found for field `%T`.`%s`", m, fieldName)
}

func DefaultToInternalValueDetector[Model any]() ToInternalValueDetector {
	return &missingFieldSkippingToInternalValueDetector[Model]{
		child: &relationshipDetector[Model]{
			internalChild: &chainingToInternalValueDetector[Model]{
				children: []ToInternalValueDetector{
					&usingGRFParsableToInternalValueDetector[Model]{},
					&isoTimeTimeToInternalValueDetector[Model]{},
					&gormDataTypesJSONToInternalValueDetector[Model]{},
					&fromTypeMapperToInternalValueDetector[Model]{
						mapper:         types.Mapper(),
						modelTypeNames: FieldTypes[Model](),
					},
					&encodingTextUnmarshalerToInternalValueDetector[Model]{},
					&usingSqlNullFieldToInternalValueDetector[Model, sql.NullBool]{
						valueFunc: func(v any) (any, error) {
							vAsBool, ok := v.(bool)
							if !ok {
								return nil, fmt.Errorf("`%s` is not a bool, it is a %T", v, v)
							}
							return sql.NullBool{
								Bool:  vAsBool,
								Valid: true,
							}, nil
						},
					},
					&usingSqlNullFieldToInternalValueDetector[Model, sql.NullInt16]{
						valueFunc: func(v any) (any, error) {
							vAsFloat64, ok := v.(float64)
							if !ok {
								return nil, fmt.Errorf("`%s` is not a float64, it is a %T", v, v)
							}
							return sql.NullInt16{
								Int16: int16(vAsFloat64),
								Valid: true,
							}, nil
						},
					},
					&usingSqlNullFieldToInternalValueDetector[Model, sql.NullInt32]{
						valueFunc: func(v any) (any, error) {
							vAsFloat64, ok := v.(float64)
							if !ok {
								return nil, fmt.Errorf("`%s` is not a float64, it is a %T", v, v)
							}
							return sql.NullInt32{
								Int32: int32(vAsFloat64),
								Valid: true,
							}, nil
						},
					},
					&usingSqlNullFieldToInternalValueDetector[Model, sql.NullInt64]{
						valueFunc: func(v any) (any, error) {
							vAsFloat64, ok := v.(float64)
							if !ok {
								return nil, fmt.Errorf("`%s` is not a float64, it is a %T", v, v)
							}
							return sql.NullInt64{
								Int64: int64(vAsFloat64),
								Valid: true,
							}, nil
						},
					},
					&usingSqlNullFieldToInternalValueDetector[Model, sql.NullFloat64]{
						valueFunc: func(v any) (any, error) {
							vAsFloat64, ok := v.(float64)
							if !ok {
								return nil, fmt.Errorf("`%s` is not a float64, it is a %T", v, v)
							}
							return sql.NullFloat64{
								Float64: vAsFloat64,
								Valid:   true,
							}, nil
						},
					},
					&usingSqlNullFieldToInternalValueDetector[Model, sql.NullString]{
						valueFunc: func(v any) (any, error) {
							vAsString, ok := v.(string)
							if !ok {
								return nil, fmt.Errorf("`%s` is not a string, it is a %T", v, v)
							}
							return sql.NullString{
								String: vAsString,
								Valid:  true,
							}, nil
						},
					},
					&usingSqlNullFieldToInternalValueDetector[Model, sql.NullByte]{
						valueFunc: func(v any) (any, error) {
							vAsStr, ok := v.(string)
							if !ok {
								return nil, fmt.Errorf("`%s` is not a string, it is a %T", v, v)
							}
							if len(vAsStr) != 1 {
								return nil, fmt.Errorf("`%s` is not a string of length 1", vAsStr)
							}
							return sql.NullByte{
								Byte:  byte(vAsStr[0]),
								Valid: true,
							}, nil
						},
					},
				},
			},
		},
	}
}
