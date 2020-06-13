package delivery

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"techpark_db/internal/pkg/service/repository"
)

type ServiceHandler struct {
	ServiceRepo repository.Repository
}

func (uh ServiceHandler) Status(ctx *fasthttp.RequestCtx) {
	status, err := uh.ServiceRepo.Status()
	if err != nil {
		log.Print(err)
		ctx.Error( err.Error(), http.StatusInternalServerError)
		return
	}

	ctx.Response.Header.Set("Content-Type", "application/json")

	if err := json.NewEncoder(ctx.Response.BodyWriter()).Encode(status); err != nil {
		log.Print(err)
		ctx.Error( "failed to encode threads to json", http.StatusInternalServerError)
		return
	}
}

func (uh ServiceHandler) Clear(ctx *fasthttp.RequestCtx) {
	err := uh.ServiceRepo.Clear()
	if err != nil {
		log.Print(err)
		ctx.Error( err.Error(), http.StatusInternalServerError)
		return
	}
}
