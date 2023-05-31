package serializers

import (
	"fmt"
	"reflect"

	"github.com/glothriel/grf/pkg/grfctx"
	"github.com/glothriel/grf/pkg/models"
)

type MissingSerializer[Model any] struct {
}

func (s *MissingSerializer[Model]) ToInternalValue(raw map[string]any, ctx *grfctx.Context) (models.InternalValue, error) {
	var m Model
	return nil, fmt.Errorf("View for model `%s` does not have a serializer, please set one using WithSerializer", reflect.TypeOf(m))
}

func (s *MissingSerializer[Model]) FromDB(raw map[string]any, ctx *grfctx.Context) (models.InternalValue, error) {
	var m Model
	return nil, fmt.Errorf("View for model `%s` does not have a serializer, please set one using WithSerializer", reflect.TypeOf(m))
}

func (s *MissingSerializer[Model]) ToRepresentation(intVal models.InternalValue, ctx *grfctx.Context) (Representation, error) {
	var m Model
	return nil, fmt.Errorf("View for model `%s` does not have a serializer, please set one using WithSerializer", reflect.TypeOf(m))
}
