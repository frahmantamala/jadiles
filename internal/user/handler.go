// this is for using handler http
package user

// import (
// 	"context"
// 	"net/http"

// 	"github.com/frahmantamala/jadiles/internal"
// 	v1 "github.com/frahmantamala/jadiles/pkg/openapi/v1"
// 	"github.com/go-chi/render"
// )

// type ServiceAPI interface {
// 	Register(ctx context.Context, req *RegisterParams) (User, error)
// 	Logout(ctx context.Context, accessToken string) error
// }

// type Handler struct {
// 	service ServiceAPI
// }

// func NewHandler(service ServiceAPI) *Handler {
// 	return &Handler{
// 		service: service,
// 	}
// }

// func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
// 	req, err := newRegisterParams(r)
// 	if err != nil {
// 		internal.HandleEndpointError(w, r, err)
// 		return
// 	}

// 	err = req.Validate(r.Context())
// 	if err != nil {
// 		internal.HandleEndpointError(w, r, err)
// 		return
// 	}

// 	u, err := h.service.Register(r.Context(), req)
// 	if err != nil {
// 		internal.HandleEndpointError(w, r, err)
// 		return
// 	}

// 	resp := v1.RegisterResponse{}
// 	resp.Data.UserProfile = u.ToV1()

// 	render.Status(r, http.StatusCreated)
// 	render.JSON(w, r, resp)
// }
