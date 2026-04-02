package stdlib

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/TRC-Loop/ccolon/vm"
)

func NewHttpModule(machine *vm.VM) *vm.ModuleValue {
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

					addr := fmt.Sprintf(":%d", port.Val)
					fmt.Printf("CColon HTTP server listening on %s\n", addr)

					mux := http.NewServeMux()
					mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						body, _ := io.ReadAll(r.Body)
						r.Body.Close()

						// Build request dict
						headerDict := &vm.DictValue{
							Entries: make(map[string]vm.Value),
							Order:   []string{},
						}
						for key := range r.Header {
							lk := strings.ToLower(key)
							headerDict.Entries[lk] = &vm.StringValue{Val: r.Header.Get(key)}
							headerDict.Order = append(headerDict.Order, lk)
						}

						reqDict := &vm.DictValue{
							Entries: map[string]vm.Value{
								"method":  &vm.StringValue{Val: r.Method},
								"path":    &vm.StringValue{Val: r.URL.Path},
								"body":    &vm.StringValue{Val: string(body)},
								"headers": headerDict,
								"query":   &vm.StringValue{Val: r.URL.RawQuery},
							},
							Order: []string{"method", "path", "body", "headers", "query"},
						}

						result, err := machine.CallFunc(handler, []vm.Value{reqDict})
						if err != nil {
							http.Error(w, "handler error: "+err.Error(), 500)
							return
						}

						// Handle response
						switch resp := result.(type) {
						case *vm.StringValue:
							w.Header().Set("Content-Type", "text/plain")
							w.WriteHeader(200)
							w.Write([]byte(resp.Val))
						case *vm.DictValue:
							// Expect {status: int, body: string, headers: dict}
							status := 200
							respBody := ""
							if s, ok := resp.Entries["status"]; ok {
								if iv, ok := s.(*vm.IntValue); ok {
									status = int(iv.Val)
								}
							}
							if b, ok := resp.Entries["body"]; ok {
								if sv, ok := b.(*vm.StringValue); ok {
									respBody = sv.Val
								}
							}
							if h, ok := resp.Entries["headers"]; ok {
								if hd, ok := h.(*vm.DictValue); ok {
									for key, val := range hd.Entries {
										if sv, ok := val.(*vm.StringValue); ok {
											w.Header().Set(key, sv.Val)
										}
									}
								}
							}
							w.WriteHeader(status)
							w.Write([]byte(respBody))
						case *vm.NilValue:
							w.WriteHeader(204)
						default:
							w.Header().Set("Content-Type", "text/plain")
							w.WriteHeader(200)
							w.Write([]byte(result.String()))
						}
					})

					// This blocks, so it won't return until the server stops
					err := http.ListenAndServe(addr, mux)
					if err != nil {
						return nil, fmt.Errorf("http.listen() failed: %s", err)
					}
					return &vm.NilValue{}, nil
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
