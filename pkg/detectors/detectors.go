package detectors

import (
	"errors"
	"fmt"

	"github.com/glothriel/grf/pkg/fields"
)

var ErrFieldShouldBeSkipped = errors.New("field is a relationship")

type relationshipDetector[Model any] struct {
	internalChild       ToInternalValueDetector
	representationChild ToRepresentationDetector[Model]
}

func (p *relationshipDetector[Model]) ToInternalValue(fieldName string) (fields.InternalValueFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings.isRelation {
		return nil, ErrFieldShouldBeSkipped
	}
	return p.internalChild.ToInternalValue(fieldName)
}

func (p *relationshipDetector[Model]) ToRepresentation(fieldName string) (fields.RepresentationFunc, error) {
	fieldSettings := getFieldSettings[Model](fieldName)
	if fieldSettings == nil {
		return nil, fmt.Errorf("Field `%s` is not present in the model", fieldName)

	}
	if fieldSettings.isRelation {
		return nil, ErrFieldShouldBeSkipped
	}
	return p.representationChild.ToRepresentation(fieldName)
}
