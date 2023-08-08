// Package views provides a set of functions that can be used to create views for models.
package views

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateModelFunc is a function that creates a new model
func CreateModelFunc[Model any](settings ModelViewSettings[Model]) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var rawElement map[string]any
		if parseErr := ctx.ShouldBindJSON(&rawElement); parseErr != nil {
			WriteError(ctx, parseErr)
			return
		}
		effectiveSerializer := settings.CreateSerializer
		if effectiveSerializer == nil {
			effectiveSerializer = settings.DefaultSerializer
		}
		internalValue, fromRawErr := effectiveSerializer.ToInternalValue(rawElement, ctx)
		if fromRawErr != nil {
			WriteError(ctx, fromRawErr)
			return
		}
		internalValue, createErr := settings.QueryDriver.CRUD().Create(ctx, internalValue)
		if createErr != nil {
			WriteError(ctx, createErr)
			return
		}
		representation, serializeErr := effectiveSerializer.ToRepresentation(internalValue, ctx)
		if serializeErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": serializeErr.Error(),
			})
			return
		}
		ctx.JSON(http.StatusCreated, representation)
	}
}
