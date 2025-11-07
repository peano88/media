package http

import (
	"context"
	"net/http"

	"github.com/peano88/medias/internal/domain"
)

type TagCreator interface {
	Execute(context.Context, domain.Tag) (domain.Tag, error)
}

func HandlePostTags(tc TagCreator) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {

		req, err := JSONIn[createTagRequest](rw, r)
		if err != nil {
			return
		}

		// There is no upfront validation that can be done here.

		tag := domain.Tag{
			Name:        req.Name,
			Description: req.Description,
		}

		// Execute business logic
		createdTag, err := tc.Execute(r.Context(), tag)
		if err != nil {
			handleExecutorError(r.Context(), rw, err)
			return
		}

		rw.Header().Set("Location", BasePath+"/"+createdTag.ID.String())

		resp := createTagResponse{
			Data: tagData{
				ID:          createdTag.ID.String(),
				Name:        createdTag.Name,
				Description: createdTag.Description,
				CreatedAt:   createdTag.CreatedAt,
				UpdatedAt:   createdTag.UpdatedAt,
			},
		}

		JSONOut(rw, http.StatusCreated, resp)
	}
}
