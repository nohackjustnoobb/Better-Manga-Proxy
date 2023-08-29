package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/valyala/fasthttp"
)

func log(ctx *fasthttp.RequestCtx) {
	fmt.Printf("[%s] %q %q -> %q %d\n", ctx.Time().Format("2006-01-02 15:04:05"), ctx.Method(), ctx.RemoteIP(), ctx.RequestURI(), ctx.Response.StatusCode())
}

func updateTimestamp(hash string) {
	os.Chtimes("cache/"+hash+".jpeg", time.Now(), time.Now())
}

func initClient() {
	// init fasthttp client
	client = &fasthttp.Client{
		NoDefaultUserAgentHeader: true,
		DisablePathNormalizing:   true,
	}
}

func fetchSettings() {
	// get proxy settings
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(os.Getenv("ADDRESS") + "driver/proxy")
	req.Header.SetMethod(fasthttp.MethodGet)

	resp := fasthttp.AcquireResponse()
	err := client.Do(req, resp)
	fasthttp.ReleaseRequest(req)

	if err == nil {
		json.Unmarshal([]byte(resp.Body()), &settings)
	}
	fasthttp.ReleaseResponse(resp)
}
