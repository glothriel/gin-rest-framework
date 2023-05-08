package models

import (
	"fmt"
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"
	uuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Base contains common columns for all tables.
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *BaseModel) BeforeCreate(tx *gorm.DB) error {
	base.ID = uuid.NewV4()
	return nil
}

type InternalValue[Model any] struct {
	Map map[string]interface{}
}

func (i *InternalValue[Model]) Fields() (map[string]interface{}, error) {
	return i.Map, nil
}

func (i *InternalValue[Model]) AsModel() (Model, error) {
	var entity Model
	decoder, decoderErr := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "json",
		Result:  &entity,
		Squash:  true,
	})

	if decoderErr != nil {
		return entity, decoderErr
	}
	decodeErr := decoder.Decode(i.Map)
	if decodeErr != nil {
		logrus.Errorf("Type of the error is %T", decodeErr)
		return entity, fmt.Errorf(
			"Failed to convert internal value `%v` to model: %w", i.Map, decodeErr,
		)
	}
	return entity, nil
}

// Uses reflect package to do a shallow translation of the model to a map wrapped in an InternalValue. We
// could use mapstructure, but it does a deep translation, which is not what we want.
func InternalValueFromModel[Model any](entity Model) (*InternalValue[Model], error) {
	v := reflect.ValueOf(entity)
	out := make(map[string]interface{})
	fields := reflect.VisibleFields(reflect.TypeOf(entity))
	for _, field := range fields {
		if !field.Anonymous {
			out[field.Tag.Get("json")] = v.FieldByName(field.Name).Interface()
		}
	}
	return &InternalValue[Model]{
		Map: out,
	}, nil
}
