package detectors

import (
	"database/sql"
	"encoding"
	"reflect"

	"github.com/glothriel/grf/pkg/fields"
	"github.com/glothriel/grf/pkg/models"
	"gorm.io/datatypes"
)

// Prints a summary with the fields of the model obtained using reflection
func FieldTypes[Model any]() map[string]string {
	ret := make(map[string]string)

	var m Model
	fields := reflect.VisibleFields(reflect.TypeOf(m))
	for _, field := range fields {
		if !field.Anonymous {
			ret[field.Tag.Get("json")] = field.Type.String()
		}
	}
	return ret
}

// Prints a summary with the fields of the model obtained using reflection
func FieldNames[Model any]() map[string]string {
	ret := make(map[string]string)

	var m Model
	fields := reflect.VisibleFields(reflect.TypeOf(m))
	for _, field := range fields {
		if !field.Anonymous {
			ret[field.Tag.Get("json")] = field.Name
		}
	}
	return ret
}

func Fields[Model any]() []string {
	fieldNames := []string{}
	var m Model
	fields := reflect.VisibleFields(reflect.TypeOf(m))
	for _, field := range fields {
		if !field.Anonymous && field.Tag.Get("json") != "" {
			fieldNames = append(fieldNames, field.Tag.Get("json"))
		}
	}
	return fieldNames
}

type fieldSettings struct {
	itsType                   reflect.Type
	isEncodingTextMarshaler   bool
	isEncodingTextUnmarshaler bool

	isGRFRepresentable bool
	isGRFParsable      bool
	isRelation         bool
	isDataTypesJSON    bool

	isSqlNullInt32 bool
}

func getFieldSettings[Model any](fieldName string) *fieldSettings {
	var entity Model
	var settings *fieldSettings
	for _, field := range reflect.VisibleFields(reflect.TypeOf(entity)) {
		jsonTag := field.Tag.Get("json")
		if jsonTag == fieldName {
			var theTypeAsAny any
			reflectedInstance := reflect.New(reflect.TypeOf(reflect.ValueOf(entity).FieldByName(field.Name).Interface())).Elem()

			settingsFromTag := models.ParseTag(field)
			_, fieldMarkedAsRelation := settingsFromTag[models.TagIsRelation]

			if reflectedInstance.CanAddr() {
				theTypeAsAny = reflectedInstance.Addr().Interface()
			} else {
				theTypeAsAny = reflectedInstance.Interface()
			}

			_, isEncodingTextMarshaler := theTypeAsAny.(encoding.TextMarshaler)
			_, isEncodingTextUnmarshaler := theTypeAsAny.(encoding.TextUnmarshaler)
			_, isGRFRepresentable := theTypeAsAny.(fields.GRFRepresentable)
			_, isGRFParsable := theTypeAsAny.(fields.GRFParsable)
			_, isSQLNull32 := theTypeAsAny.(*sql.NullInt32)
			_, isDataTypesJSON := theTypeAsAny.(*datatypes.JSON)

			settings = &fieldSettings{
				itsType: reflect.TypeOf(
					reflectedInstance.Interface(),
				),
				isEncodingTextMarshaler:   isEncodingTextMarshaler,
				isEncodingTextUnmarshaler: isEncodingTextUnmarshaler,
				isGRFRepresentable:        isGRFRepresentable,
				isGRFParsable:             isGRFParsable,
				isRelation:                fieldMarkedAsRelation,
				isDataTypesJSON:           isDataTypesJSON,
				isSqlNullInt32:            isSQLNull32,
			}
		}
	}
	return settings
}
