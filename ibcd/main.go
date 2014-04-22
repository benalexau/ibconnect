package main

import (
	"log"

	"github.com/benalexau/ibconnect/core"
	"github.com/benalexau/ibconnect/gateway"
	"github.com/benalexau/ibconnect/server"
)

func main() {
	c, err := core.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	ctx, err := core.NewContext(c)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	ffs := gateway.FeedFactories(c)
	gatewayController, err := gateway.NewGatewayController(ffs, ctx.DB, ctx.N, ctx.DL, c.IbGws, c.IbClientId)
	if err != nil {
		log.Fatal(err)
	}
	defer gatewayController.Close()

	terminated := handleSignals()

	handler := server.Handler(c.ErrInfo, ctx.DB, ctx.N)
	err = server.Serve(terminated, c.Address(), handler)
	if err != nil {
		log.Fatal(err)
	}

	// ensure we have terminated (useful if HTTP server commented out etc)
	<-terminated
}
