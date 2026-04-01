package stdlib

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TRC-Loop/ccolon/vm"
)

func NewHttpModule() *vm.ModuleValue {
	return &vm.ModuleValue{
		Name: "http",
		Methods: map[string]*vm.NativeFuncValue{
			"get": {
				Name: "http.get",
				Fn:   httpRequest("GET"),
			},
			"post": {
				Name: "http.post",
				Fn:   httpRequest("POST"),
			},
			"put": {
				Name: "http.put",
				Fn:   httpRequest("PUT"),
			},
			"delete": {
				Name: "http.delete",
				Fn:   httpRequest("DELETE"),
			},
			"listen": {
				Name: "http.listen",
				Fn: func(args []vm.Value) (vm.Value, error) {
					if len(args) != 2 {
						return nil, fmt.Errorf("http.listen() takes 2 arguments (port, handler)")
					}
					port, ok := args[0].(*vm.IntValue)
					if !ok {
						return nil, fmt.Errorf("http.listen() port must be an int")
					}
					handler, ok := args[1].(*vm.FuncValue)
					if !ok {
						return nil, fmt.Errorf("http.listen() handler must be a function")
					}
					_ = handler // handler will be called via VM callback
					addr := fmt.Sprintf(":%d", port.Val)
					fmt.Printf("CColon HTTP server listening on %s\n", addr)
					// We store the handler ref but can't call it without VM access.
					// The server integration is done through the VM's CallFunc.
					return nil, fmt.Errorf("http.listen() requires VM callback support (use the listen helper)")
				},
			},
		},
	}
}

func httpRequest(method string) func(args []vm.Value) (vm.Value, error) {
	return func(args []vm.Value) (vm.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("http.%s() requires at least 1 argument (url)", strings.ToLower(method))
		}
		url, ok := args[0].(*vm.StringValue)
		if !ok {
			return nil, fmt.Errorf("http.%s() url must be a string", strings.ToLower(method))
		}

		var body io.Reader
		if len(args) >= 2 {
			if b, ok := args[1].(*vm.StringValue); ok {
				body = strings.NewReader(b.Val)
			}
		}

		req, err := http.NewRequest(method, url.Val, body)
		if err != nil {
			return nil, fmt.Errorf("http.%s() failed to create request: %s", strings.ToLower(method), err)
		}

		// headers from dict argument
		if len(args) >= 3 {
			if headers, ok := args[2].(*vm.DictValue); ok {
				for key, val := range headers.Entries {
					if sv, ok := val.(*vm.StringValue); ok {
						req.Header.Set(key, sv.Val)
					}
				}
			}
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("http.%s() request failed: %s", strings.ToLower(method), err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("http.%s() failed to read response: %s", strings.ToLower(method), err)
		}

		// build response headers dict
		headerDict := &vm.DictValue{
			Entries: make(map[string]vm.Value),
			Order:   []string{},
		}
		for key := range resp.Header {
			headerDict.Entries[strings.ToLower(key)] = &vm.StringValue{Val: resp.Header.Get(key)}
			headerDict.Order = append(headerDict.Order, strings.ToLower(key))
		}

		return &vm.DictValue{
			Entries: map[string]vm.Value{
				"status":  &vm.IntValue{Val: int64(resp.StatusCode)},
				"body":    &vm.StringValue{Val: string(respBody)},
				"headers": headerDict,
			},
			Order: []string{"status", "body", "headers"},
		}, nil
	}
}
