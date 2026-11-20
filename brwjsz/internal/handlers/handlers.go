package handlers

import (
	"context"

	"github.com/infinity6-ai/gox/brwjsz/internal/internal/stzfiles"
	"github.com/infinity6-ai/gox/httpz/httpzserver"
	"github.com/infinity6-ai/gox/httpz/httpzserver/httpzserverstatic"
)

const BasePath = "/api/brwjs/f"
const ControllerPath = BasePath + "controller.html"

func AddHanlers(ctx context.Context, s *httpzserver.Server) {
	s.AddHandler("GET", BasePath, httpzserverstatic.Handler(stzfiles.Name, httpzserverstatic.Options{
		Prefix:  BasePath,
		NoIndex: false,
	}))
}
