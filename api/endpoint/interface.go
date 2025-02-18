package endpoints

import (
	"github.com/danielgtaylor/huma/v2"
)

type Endpoint interface {
	RegisterRoutes(huma.API, string)
}
