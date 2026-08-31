package httpzserver

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/infinity6-ai/gox/commonz/deferz"
)

type Server struct {
	Context      context.Context
	LocalAddress string

	server *http.Server
}

func (me *Server) Listen() {
	dfz := deferz.New()
	defer dfz.Close()
	if me.LocalAddress == "" {
		port := os.Getenv("PORT")
		if port != "" {
			me.LocalAddress = fmt.Sprintf("0.0.0.0:%s", port)
		}
	}
	if me.LocalAddress == "" {
		me.LocalAddress = "localhost:8080"
	}
	// me.server.Handler = me.handler

	// listener, err := net.Listen("tcp", me.LocalAddress)
	// errorz.Check(err)

	// util.Check(err)
	// me.listeners = []net.Listener{listener}

	// if me.customLocalAddresses != nil {
	// 	for _, addr := range me.customLocalAddresses {
	// 		listener, err := net.Listen("tcp", addr)
	// 		util.Check(err)
	// 		me.listeners = append(me.listeners, listener)
	// 	}
	// }

	// logzFilter := &logzFilter{
	// 	ServName:    me.ServName,
	// 	ServVersion: me.ServVersion,
	// }
	// me.AddFilter(logzFilter.Filter)

	// me.AddFilter(deferzfilter.Filter)

	// if me.Audit {
	// 	auditFilter := &auditfilter.AuditorFilter{ServName: me.ServName, ServVersion: me.ServVersion}
	// 	me.AddFilter(auditFilter.Filter)

	// }

	// me.AddFilter(gzipfilter.Filter)

	// me.corsFilter = &corzFilter{}
	// me.AddFilter(me.corsFilter.Filter)

	// me.AddFilter((&errorFilter{}).Filter)
	// logger.Info(me.Context, "http server started", map[string]any{"name": me.ServName, "version": me.ServVersion, "address": me.Address(), "addresses": me.customLocalAddresses})
}
