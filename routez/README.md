# `routez` API Implementation and Client Generation

This document explains how to implement APIs using the `routez` package and how to generate type-safe clients for them using `apiclientz`, demonstrated through the `sampleroutez/routezsamplefraction` example.

## Getting Started

To use the `routez` package and its related utilities, first add it to your project:

```bash
go get github.com/infinity6-ai/gox/routez
```

This will download the necessary packages. The key packages you will be working with are:

*   `github.com/infinity6-ai/gox/routez/routez`: The core package for registering `apiz.Api` definitions with an HTTP server.
*   `github.com/infinity6-ai/gox/routez/apiz`: Provides the `Api` struct for combining schema and handler logic.
*   `github.com/infinity6-ai/gox/schemaz/schemaz`: Used to define the API contract, including paths, methods, and data shapes.
*   `github.com/infinity6-ai/gox/routez/apiclientz`: A utility for generating type-safe Go clients from your API definitions.

## `routez` API Implementation

The `routez` package simplifies API definition by allowing developers to declare API schemas and handlers in a structured way.

### Core Concepts for API Implementation

1.  **The `*ReqResp` Struct**: A central struct that encapsulates all incoming request data (path parameters, query parameters, headers, body) and outgoing response data (headers, body).
2.  **The `GetDataRefs()` Method**: A method on the `*ReqResp` struct that maps its fields to the `apiz.DataRefs` structure, which `routez` uses internally to bind HTTP request/response elements.
3.  **The `Schema()` Function**: Defines the API's contract using `schemaz.Api`, specifying HTTP method, path, and detailed schemas for request and response elements.
4.  **The `Api()` Function**: Combines the defined schema with the API's business logic (handler function) into a single `apiz.Api` object, ready for registration with an HTTP server.

---

### Example: `FractionReqResp` Struct

This struct acts as a single source of truth for all data related to a specific API call. It contains nested structs for organizing request and response components.

```go
type Result struct {
	Display string `json:"display"`
	Result  string `json:"result"`
}

type FractionReq struct {
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
	Precision   int     `json:"precision"`
	TraceId     string  `json:"x-i6-trace-id"`
	Reason      string  `json:"reason"`
}

type FractionResp struct {
	TraceMessage string  `json:"x-i6-trace-message"`
	Result       *Result `json:"result"`
}

type FractionReqResp struct {
	Req  *FractionReq
	Resp *FractionResp
}
```

### Example: `GetDataRefs()` Method

This method is crucial for `routez` to understand how to populate the `FractionReqResp` struct from an incoming HTTP request and extract data for the HTTP response. It returns an `apiz.DataRefs` object, where each field points to a specific part of the `FractionReqResp` struct.

```go
func (f *FractionReqResp) GetDataRefs() *apiz.DataRefs {
	if f.Req == nil {
		f.Req = &FractionReq{}
	}
	if f.Resp == nil {
		f.Resp = &FractionResp{}
	}
	if f.Resp.Result == nil {
		f.Resp.Result = &Result{}
	}
	return &apiz.DataRefs{
		PathParams: &struct {
			Numerator   *float64 `json:"numerator"`
			Denominator *float64 `json:"denominator"`
		}{
			&f.Req.Numerator,
			&f.Req.Denominator,
		},
		QueryParams: &struct {
			Precision *int `json:"precision"`
		}{
			&f.Req.Precision,
		},
		ReqHeaders: &struct {
			TraceId *string `json:"x-i6-trace-id"`
		}{
			&f.Req.TraceId,
		},
		ReqBody: &struct {
			Reason *string `json:"reason"`
		}{
			&f.Req.Reason,
		},
		RespHeaders: &struct {
			TraceMessage *string `json:"x-i6-trace-message"`
		}{
			&f.Resp.TraceMessage,
		},
		RespBody: f.Resp.Result,
	}
}
```

### Example: `Schema()` Function

This function defines the API's public contract. It specifies the HTTP method, the URL path (including path parameters), and the types and descriptions for all expected request and response elements.

```go
func Schema() *schemaz.Api {
	return &schemaz.Api{
		Id: "samplefraction",

		Desc: schemaz.Desc{
			Name:     "Sample Fraction",
			Summary:  "Calculates the result of a fraction with a given precision.",
			Markdown: "# Sample Fraction API\n\nThis API demonstrates a simple fraction calculation. It accepts a numerator and a denominator as path parameters, a precision as a query parameter, and returns the result in a JSON object.",
		},

		Method: "POST",
		Path:   "/api/gox/routez/sample/fraction/{numerator}/{denominator}",

		ReqParams: []schemaz.Field{
			{Name: "numerator", Desc: schemaz.Desc{Summary: "numerator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
			{Name: "denominator", Desc: schemaz.Desc{Summary: "denominator"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
		},

		ReqQuery: []schemaz.Field{
			{Name: "precision", Desc: schemaz.Desc{Summary: "precision"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
		},

		ReqHeaders: []schemaz.Field{
			{Name: "x_i6_trace_id", Desc: schemaz.Desc{Summary: "trace id"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
		},

		ReqBody: &schemaz.Spec{
			Type: schemaz.TypeObject,
			Fields: []schemaz.Field{
				{Name: "reason", Desc: schemaz.Desc{Summary: "reason"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
			},
		},

		RespHeaders: []schemaz.Field{
			{Name: "x_i6_trace_message", Desc: schemaz.Desc{Summary: "trace message"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
		},

		RespBody: &schemaz.Spec{
			Type: schemaz.TypeObject,
			Fields: []schemaz.Field{
				{Name: "display", Desc: schemaz.Desc{Summary: "fraction display"}, Spec: schemaz.Spec{Type: schemaz.TypeString}},
				{Name: "result", Desc: schemaz.Desc{Summary: "fraction result"}, Spec: schemaz.Spec{Type: schemaz.TypeNumber}},
			},
		},
	}
}
```

### Example: `Api()` Function and Handler

This is where the API comes to life. It takes the `schemaz.Api` definition and couples it with a `Handler` function that contains the actual business logic. The `Handler` receives the `FractionReqResp` struct, already populated with request data by `routez`, and is responsible for filling the response fields.

```go
func Api() *apiz.Api[*FractionReqResp] {
	return &apiz.Api[*FractionReqResp]{
		Schema: Schema(),
		Handler: func(ctx context.Context, reqResp *FractionReqResp) (int, error) {
			reqResp.Resp.TraceMessage = "reason: " + reqResp.Req.Reason + ", trace: " + reqResp.Req.TraceId
			reqResp.Resp.Result.Display = fmt.Sprintf(fmt.Sprintf("%%.%df/%%.%df", int(reqResp.Req.Precision), int(reqResp.Req.Precision)), reqResp.Req.Numerator, reqResp.Req.Denominator)
			reqResp.Resp.Result.Result = strconv.FormatFloat(reqResp.Req.Numerator/reqResp.Req.Denominator, 'f', reqResp.Req.Precision, 64)

			return 201, nil
		},
	}
}
```

### Registering the API

To make this API accessible, you would typically register it with an `httpzserver` instance. The `Api()` function returns an `apiz.Api` object that can be passed directly to `routez.Register()`.

```go
// Example usage (from a test file or main application setup)
// Assuming 's' is an httpzserver.Server instance

// Import necessary packages:
// "github.com/infinity6-ai/gox/httpz/httpzserver"
// "github.com/infinity6-ai/gox/routez/routez"
// "github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"

// s := httpzserver.New(ctx, httpzserver.Options{})
// defer s.Close()
// s.Listen()
// s.Start()

// routez.Register(s, routezsamplefraction.Api())

// The API is now available at the path defined in Schema():
// POST /api/gox/routez/sample/fraction/{numerator}/{denominator}
```

---

## API Client Generation (`apiclientz`)

The `apiclientz` package provides a way to generate type-safe client functions for your `routez` APIs, simplifying API consumption and ensuring consistency between server and client.

### Core Concepts for Client Generation

1.  **`apiclientz.Get()` Function**: This function takes an `httpzclient.Client` and an `apiz.Api` definition to return a client-side handler function. This handler takes the same `*ReqResp` struct as the server-side API handler, allowing for a consistent API interface.
2.  **`parseRequest()`**: (Internal) Converts the `*ReqResp` struct into an `httpzrequest.Req` suitable for sending over HTTP. It uses the `GetDataRefs()` method of the `*ReqResp` struct to correctly map data to path parameters, query parameters, headers, and the request body.
3.  **`writeResponse()`**: (Internal) Populates the `Resp` fields of the `*ReqResp` struct with data received from the HTTP response. It uses `GetDataRefs()` to correctly map response headers and the response body back into the `*ReqResp` struct.

### Example: Using `apiclientz.Get()`

Here's how you can use `apiclientz.Get()` to create and use an API client:

```go
// From schemazsamplefraction_test.go
// Assuming 'c' is an httpzclient.Client instance configured to reach your server
// and routezsamplefraction.Api() returns the apiz.Api definition for your endpoint.

// Import necessary packages:
// "github.com/infinity6-ai/gox/httpz/httpzclient"
// "github.com/infinity6-ai/gox/routez/apiclientz"
// "github.com/infinity6-ai/gox/routez/sampleroutez/routezsamplefraction"

// c := httpzclient.New(ctx, httpzclient.Options{})
// // Assume s.Base() is the base URL of the httpzserver
// // c := httpzclient.New(ctx, httpzclient.Options{
// // 	BaseUrl: s.Base(),
// // })

ac := apiclientz.Get(c, routezsamplefraction.Api())

reqResp := &routezsamplefraction.FractionReqResp{
	Req: &routezsamplefraction.FractionReq{
		Numerator:   10,
		Denominator: 3,
		Precision:   3,
		TraceId:     "xx",
		Reason:      "myreason",
	},
}

// Call the generated client function with the reqResp struct
status, err := ac(ctx, reqResp)

// After the call, reqResp.Resp will be populated with the response data
// For example:
//   require.Equal(t, 201, status)
//   require.Equal(t, "reason: myreason, trace: xx", reqResp.Resp.TraceMessage)
//   require.Equal(t, &routezsamplefraction.Result{
//   	Display: "10.000/3.000",
//   	Result:  "3.333",
//   }, reqResp.Resp.Result)
// 
```

This comprehensive approach allows for efficient and type-safe development of both API servers and their corresponding clients in Go.
