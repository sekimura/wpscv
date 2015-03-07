package main

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gopherjs/go-angularjs"
	"honnef.co/go/js/xhr"
)

func doFetch(u string, onSuccess func(*FetchResult), onError func(error)) {
	parsed, err := url.Parse(u)
	if err != nil {
		onError(err)
		return
	}

	escaped := url.QueryEscape(parsed.String())
	req := xhr.NewRequest("GET", "/api/fetch?u="+escaped)
	go func() {
		err := req.Send(nil)
		if err != nil {
			onError(err)
			return
		}
		if req.Status != 200 {
			msg := req.Response.String()
			onError(fmt.Errorf("Failed to get result from the API endpoint: %v", msg))
			return
		}
		var fr FetchResult
		buf := []byte(req.Response.String())
		if err := json.Unmarshal(buf, &fr); err != nil {
			onError(err)
			return
		}
		onSuccess(&fr)
	}()
}

func FetcherCtrl(scope *angularjs.Scope) {
	scope.Set("result", nil)
	scope.Set("url", "http://sekimura.org")
	scope.Set("highlighted", nil)

	onSuccess := func(fr *FetchResult) {
		scope.Set("fetching", false)
		scope.Apply(func() {
			scope.Set("result", fr)
		})
	}

	onError := func(err error) {
		scope.Set("fetching", false)
		scope.Apply(func() {
			scope.Set("error", err.Error())
		})
	}

	scope.Set("fetch", func() {
		scope.Set("error", nil)
		scope.Set("highlighted", nil)
		scope.Set("result", nil)
		scope.Set("fetching", true)

		u := scope.Get("url").String()
		doFetch(u, onSuccess, onError)
	})

	scope.Set("highlight", func(s string) {
		scope.Set("highlighted", s)
	})
}

func main() {
	app := angularjs.NewModule("myapp", nil, nil)
	app.NewController("FetcherCtrl", FetcherCtrl)
}
