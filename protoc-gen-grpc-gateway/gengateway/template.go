package gengateway

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/golang/glog"
	generator2 "github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/jslyzt/grpc-gateway/protoc-gen-grpc-gateway/descriptor"
	"github.com/jslyzt/grpc-gateway/utilities"
)

type param struct {
	*descriptor.File
	Imports            []descriptor.GoPackage
	UseRequestContext  bool
	RegisterFuncSuffix string
	AllowPatchFeature  bool
}

type binding struct {
	*descriptor.Binding
	Registry          *descriptor.Registry
	AllowPatchFeature bool
}

// GetBodyFieldPath returns the binding body's fieldpath.
func (b binding) GetBodyFieldPath() string {
	if b.Body != nil && len(b.Body.FieldPath) != 0 {
		return b.Body.FieldPath.String()
	}
	return "*"
}

// GetBodyFieldPath returns the binding body's struct field name.
func (b binding) GetBodyFieldStructName() (string, error) {
	if b.Body != nil && len(b.Body.FieldPath) != 0 {
		return generator2.CamelCase(b.Body.FieldPath.String()), nil
	}
	return "", errors.New("No body field found")
}

// HasQueryParam determines if the binding needs parameters in query string.
// It sometimes returns true even though actually the binding does not need.
// But it is not serious because it just results in a small amount of extra codes generated.
func (b binding) HasQueryParam() bool {
	if b.Body != nil && len(b.Body.FieldPath) == 0 {
		return false
	}
	fields := make(map[string]bool)
	for _, f := range b.Method.RequestType.Fields {
		fields[f.GetName()] = true
	}
	if b.Body != nil {
		delete(fields, b.Body.FieldPath.String())
	}
	for _, p := range b.PathParams {
		delete(fields, p.FieldPath.String())
	}
	return len(fields) > 0
}

func (b binding) QueryParamFilter() queryParamFilter {
	var seqs [][]string
	if b.Body != nil {
		seqs = append(seqs, strings.Split(b.Body.FieldPath.String(), "."))
	}
	for _, p := range b.PathParams {
		seqs = append(seqs, strings.Split(p.FieldPath.String(), "."))
	}
	return queryParamFilter{utilities.NewDoubleArray(seqs)}
}

// HasEnumPathParam returns true if the path parameter slice contains a parameter
// that maps to an enum proto field that is not repeated, if not false is returned.
func (b binding) HasEnumPathParam() bool {
	return b.hasEnumPathParam(false)
}

// HasRepeatedEnumPathParam returns true if the path parameter slice contains a parameter
// that maps to a repeated enum proto field, if not false is returned.
func (b binding) HasRepeatedEnumPathParam() bool {
	return b.hasEnumPathParam(true)
}

// hasEnumPathParam returns true if the path parameter slice contains a parameter
// that maps to a enum proto field and that the enum proto field is or isn't repeated based on the provided 'repeated' parameter.
func (b binding) hasEnumPathParam(repeated bool) bool {
	for _, p := range b.PathParams {
		if p.IsEnum() && p.IsRepeated() == repeated {
			return true
		}
	}
	return false
}

// LookupEnum looks up a enum type by path parameter.
func (b binding) LookupEnum(p descriptor.Parameter) *descriptor.Enum {
	e, err := b.Registry.LookupEnum("", p.Target.GetTypeName())
	if err != nil {
		return nil
	}
	return e
}

// FieldMaskField returns the golang-style name of the variable for a FieldMask, if there is exactly one of that type in
// the message. Otherwise, it returns an empty string.
func (b binding) FieldMaskField() string {
	var fieldMaskField *descriptor.Field
	for _, f := range b.Method.RequestType.Fields {
		if f.GetTypeName() == ".google.protobuf.FieldMask" {
			// if there is more than 1 FieldMask for this request, then return none
			if fieldMaskField != nil {
				return ""
			}
			fieldMaskField = f
		}
	}
	if fieldMaskField != nil {
		return generator2.CamelCase(fieldMaskField.GetName())
	}
	return ""
}

// queryParamFilter is a wrapper of utilities.DoubleArray which provides String() to output DoubleArray.Encoding in a stable and predictable format.
type queryParamFilter struct {
	*utilities.DoubleArray
}

func (f queryParamFilter) String() string {
	encodings := make([]string, len(f.Encoding))
	for str, enc := range f.Encoding {
		encodings[enc] = fmt.Sprintf("%q: %d", str, enc)
	}
	e := strings.Join(encodings, ", ")
	return fmt.Sprintf("&utilities.DoubleArray{Encoding: map[string]int{%s}, Base: %#v, Check: %#v}", e, f.Base, f.Check)
}

type trailerParams struct {
	Services           []*descriptor.Service
	UseRequestContext  bool
	RegisterFuncSuffix string
	AssumeColonVerb    bool
}

func applyTemplate(p param, reg *descriptor.Registry) (string, error) {
	w := bytes.NewBuffer(nil)
	if err := headerTemplate.Execute(w, p); err != nil {
		return "", err
	}
	var targetServices []*descriptor.Service

	for _, msg := range p.Messages {
		msgName := generator2.CamelCase(*msg.Name)
		msg.Name = &msgName
	}
	for _, svc := range p.Services {
		var methodWithBindingsSeen bool
		svcName := generator2.CamelCase(*svc.Name)
		svc.Name = &svcName
		for _, meth := range svc.Methods {
			glog.V(2).Infof("Processing %s.%s", svc.GetName(), meth.GetName())
			methName := generator2.CamelCase(*meth.Name)
			meth.Name = &methName
			for _, b := range meth.Bindings {
				methodWithBindingsSeen = true
				if err := handlerTemplate.Execute(w, binding{
					Binding:           b,
					Registry:          reg,
					AllowPatchFeature: p.AllowPatchFeature,
				}); err != nil {
					return "", err
				}

				// Local
				if err := localHandlerTemplate.Execute(w, binding{
					Binding:           b,
					Registry:          reg,
					AllowPatchFeature: p.AllowPatchFeature,
				}); err != nil {
					return "", err
				}
			}
		}
		if methodWithBindingsSeen {
			targetServices = append(targetServices, svc)
		}
	}
	if len(targetServices) == 0 {
		return "", errNoTargetService
	}

	assumeColonVerb := true
	if reg != nil {
		assumeColonVerb = !reg.GetAllowColonFinalSegments()
	}
	tp := trailerParams{
		Services:           targetServices,
		UseRequestContext:  p.UseRequestContext,
		RegisterFuncSuffix: p.RegisterFuncSuffix,
		AssumeColonVerb:    assumeColonVerb,
	}
	// Local
	if err := localTrailerTemplate.Execute(w, tp); err != nil {
		return "", err
	}

	if err := trailerTemplate.Execute(w, tp); err != nil {
		return "", err
	}
	return w.String(), nil
}

var (
	headerTemplate = template.Must(template.New("header").Parse(`
package {{.GoPkg.Name}}
import (
	{{range $i := .Imports}}{{if $i.Standard}}{{$i | printf "%s\n"}}{{end}}{{end}}

	{{range $i := .Imports}}{{if not $i.Standard}}{{$i | printf "%s\n"}}{{end}}{{end}}
)
`))

	handlerTemplate = template.Must(template.New("handler").Parse(`
{{if and .Method.GetClientStreaming .Method.GetServerStreaming}}
{{template "bidi-streaming-request-func" .}}
{{else if .Method.GetClientStreaming}}
{{template "client-streaming-request-func" .}}
{{else}}
{{template "client-rpc-request-func" .}}
{{end}}
`))

	_ = template.Must(handlerTemplate.New("request-func-signature").Parse(strings.Replace(`
{{if .Method.GetServerStreaming}}
func request_{{.Method.Service.GetName}}_{{.Method.GetName}}_{{.Index}}(ctx context.Context, marshaler runtime.Marshaler, client {{.Method.Service.GetName}}Client, req *http.Request, params runtime.Params) ({{.Method.Service.GetName}}_{{.Method.GetName}}Client, *runtime.ServerMetadata, error)
{{else}}
func request_{{.Method.Service.GetName}}_{{.Method.GetName}}_{{.Index}}(ctx context.Context, marshaler runtime.Marshaler, client {{.Method.Service.GetName}}Client, req *http.Request, params runtime.Params) (proto.Message, *runtime.ServerMetadata, error)
{{end}}`, "\n", "", -1)))

	_ = template.Must(handlerTemplate.New("client-streaming-request-func").Parse(`
{{template "request-func-signature" .}} {
	stream, err := client.{{.Method.GetName}}(ctx)
	if err != nil {
		grpclog.Infof("Failed to start streaming: %v", err)
		return nil, nil, err
	}
	var (
		err error
		metadata *runtime.ServerMetadata
		protoReq {{.Method.RequestType.GoType .Method.Service.File.GoPkg.Path}}
	)
	metadata, err = request.StreamSend(stream, &protoReq, marshaler, req)
	if err != nil {
		return nil, nil, err
	}
{{if .Method.GetServerStreaming}}
	return stream, metadata, nil
{{else}}
	msg, err := stream.CloseAndRecv()
	metadata.TrailerMD = stream.Trailer()
	return msg, metadata, err
{{end}}
}
`))

	_ = template.Must(handlerTemplate.New("client-rpc-request-func").Parse(`
{{$AllowPatchFeature := .AllowPatchFeature}}
{{if .HasQueryParam}}
var (
	filter_{{.Method.Service.GetName}}_{{.Method.GetName}}_{{.Index}} = {{.QueryParamFilter}}
)
{{end}}
{{template "request-func-signature" .}} {
	var protoReq {{.Method.RequestType.GoType .Method.Service.File.GoPkg.Path}}
{{if .Body}}
	reader, err := request.StreamDecode(marshaler, req, &{{.Body.AssignableExpr "protoReq"}})
	if err != nil {
		return nil, nil, err
	}
	{{- if and $AllowPatchFeature (eq (.HTTPMethod) "PATCH") (.FieldMaskField) (not (eq "*" .GetBodyFieldPath)) }}
	if protoReq.{{.FieldMaskField}} == nil || len(protoReq.{{.FieldMaskField}}.GetPaths()) == 0 {
			_, md := descriptor.ForMessage(protoReq.{{.GetBodyFieldStructName}})
			if fieldMask, err := runtime.FieldMaskFromRequestBody(reader(), md); err != nil {
				return nil, nil, status.Errorf(codes.InvalidArgument, "%v", err)
			} else {
				protoReq.{{.FieldMaskField}} = fieldMask
			}
	}
	{{end}}
{{end}}
{{if .PathParams}}
	var (
		err error
{{- if .HasEnumPathParam}}
		e int32
{{- end}}
{{- if .HasRepeatedEnumPathParam}}
		es []int32
{{- end}}
	)
	{{$binding := .}}
	{{range $param := .PathParams}}
	{{$enum := $binding.LookupEnum $param}}
{{if $param.IsNestedProto3}}
	err = runtime.PopulateFieldFromPath(&protoReq, params, {{$param | printf "%q"}})
	{{if $enum}}
		e{{if $param.IsRepeated}}s{{end}}, err = params.{{$param.ConvertFuncExpr}}({{$param | printf "%q"}}{{if $param.IsRepeated}}, {{$binding.Registry.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}}, {{$enum.GoType $param.Target.Message.File.GoPkg.Path}}_value)
	{{end}}
{{else if $enum}}
	e{{if $param.IsRepeated}}s{{end}}, err = params.{{$param.ConvertFuncExpr}}({{$param | printf "%q"}}{{if $param.IsRepeated}}, {{$binding.Registry.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}}, {{$enum.GoType $param.Target.Message.File.GoPkg.Path}}_value)
{{else}}
	{{$param.AssignableExpr "protoReq"}}, err = params.{{$param.ConvertFuncExpr}}({{$param | printf "%q"}}{{if $param.IsRepeated}}, {{$binding.Registry.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}})
{{end}}
	if err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", {{$param | printf "%q"}}, err)
	}
{{if and $enum $param.IsRepeated}}
	s := make([]{{$enum.GoType $param.Target.Message.File.GoPkg.Path}}, len(es))
	for i, v := range es {
		s[i] = {{$enum.GoType $param.Target.Message.File.GoPkg.Path}}(v)
	}
	{{$param.AssignableExpr "protoReq"}} = s
{{else if $enum}}
	{{$param.AssignableExpr "protoReq"}} = {{$enum.GoType $param.Target.Message.File.GoPkg.Path}}(e)
{{end}}
	{{end}}
{{end}}
{{if .HasQueryParam}}
	if err = req.ParseForm(); err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	if err = runtime.PopulateQueryParameters(&protoReq, req.Form, filter_{{.Method.Service.GetName}}_{{.Method.GetName}}_{{.Index}}); err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
{{end}}
	var metadata *runtime.ServerMetadata
{{if .Method.GetServerStreaming}}
	stream, err := client.{{.Method.GetName}}(ctx, &protoReq)
	if err != nil {
		return nil, nil, err
	}
	metadata, err = request.StreamGainMetadata(stream)
	return stream, metadata, nil
{{else}}
	metadata = &runtime.ServerMetadata{}
	msg, err := client.{{.Method.GetName}}(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err
{{end}}
}`))

	_ = template.Must(handlerTemplate.New("bidi-streaming-request-func").Parse(`
{{template "request-func-signature" .}} {
	stream, err := client.{{.Method.GetName}}(ctx)
	if err != nil {
		grpclog.Infof("Failed to start streaming: %v", err)
		return nil, nil, err
	}
	var protoReq {{.Method.RequestType.GoType .Method.Service.File.GoPkg.Path}}
	metadata, err = request.StreamBiDiSend(stream, &protoReq, marshaler, req)
	return stream, metadata, nil
}
`))

	localHandlerTemplate = template.Must(template.New("local-handler").Parse(`
{{if and .Method.GetClientStreaming .Method.GetServerStreaming}}
{{else if .Method.GetClientStreaming}}
{{else if .Method.GetServerStreaming}}
{{else}}
{{template "local-client-rpc-request-func" .}}
{{end}}
`))

	_ = template.Must(localHandlerTemplate.New("local-request-func-signature").Parse(strings.Replace(`
{{if .Method.GetServerStreaming}}
{{else}}
func local_request_{{.Method.Service.GetName}}_{{.Method.GetName}}_{{.Index}}(ctx context.Context, marshaler runtime.Marshaler, server {{.Method.Service.GetName}}Server, req *http.Request, params runtime.Params) (proto.Message, *runtime.ServerMetadata, error)
{{end}}`, "\n", "", -1)))

	_ = template.Must(localHandlerTemplate.New("local-client-rpc-request-func").Parse(`
{{$AllowPatchFeature := .AllowPatchFeature}}
{{template "local-request-func-signature" .}} {
	var protoReq {{.Method.RequestType.GoType .Method.Service.File.GoPkg.Path}}
{{if .Body}}
	reader, err := request.StreamDecode(marshaler, req, &{{.Body.AssignableExpr "protoReq"}})
	if err != nil {
		return nil, nil, err
	}
	{{- if and $AllowPatchFeature (eq (.HTTPMethod) "PATCH") (.FieldMaskField) (not (eq "*" .GetBodyFieldPath)) }}
	if protoReq.{{.FieldMaskField}} == nil || len(protoReq.{{.FieldMaskField}}.GetPaths()) == 0 {
			_, md := descriptor.ForMessage(protoReq.{{.GetBodyFieldStructName}})
			if fieldMask, err := runtime.FieldMaskFromRequestBody(reader(), md); err != nil {
				return nil, nil, status.Errorf(codes.InvalidArgument, "%v", err)
			} else {
				protoReq.{{.FieldMaskField}} = fieldMask
			}
	}
	{{end}}
{{end}}
{{if .PathParams}}
	var (
		err error
{{- if .HasEnumPathParam}}
		e int32
{{- end}}
{{- if .HasRepeatedEnumPathParam}}
		es []int32
{{- end}}
	)
	{{$binding := .}}
	{{range $param := .PathParams}}
	{{$enum := $binding.LookupEnum $param}}
{{if $param.IsNestedProto3}}
	err = runtime.PopulateFieldFromPath(&protoReq, params, {{$param | printf "%q"}})
	{{if $enum}}
		e{{if $param.IsRepeated}}s{{end}}, err = params.{{$param.ConvertFuncExpr}}({{$param | printf "%q"}}{{if $param.IsRepeated}}, {{$binding.Registry.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}}, {{$enum.GoType $param.Target.Message.File.GoPkg.Path}}_value)
	{{end}}
{{else if $enum}}
	e{{if $param.IsRepeated}}s{{end}}, err = params.{{$param.ConvertFuncExpr}}({{$param | printf "%q"}}{{if $param.IsRepeated}}, {{$binding.Registry.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}}, {{$enum.GoType $param.Target.Message.File.GoPkg.Path}}_value)
{{else}}
	{{$param.AssignableExpr "protoReq"}}, err = params.{{$param.ConvertFuncExpr}}({{$param | printf "%q"}}{{if $param.IsRepeated}}, {{$binding.Registry.GetRepeatedPathParamSeparator | printf "%c" | printf "%q"}}{{end}})
{{end}}
	if err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "type mismatch, parameter: %s, error: %v", {{$param | printf "%q"}}, err)
	}
{{if and $enum $param.IsRepeated}}
	s := make([]{{$enum.GoType $param.Target.Message.File.GoPkg.Path}}, len(es))
	for i, v := range es {
		s[i] = {{$enum.GoType $param.Target.Message.File.GoPkg.Path}}(v)
	}
	{{$param.AssignableExpr "protoReq"}} = s
{{else if $enum}}
	{{$param.AssignableExpr "protoReq"}} = {{$enum.GoType $param.Target.Message.File.GoPkg.Path}}(e)
{{end}}
	{{end}}
{{end}}
{{if .HasQueryParam}}
	if err := runtime.PopulateQueryParameters(&protoReq, req.URL.Query(), filter_{{.Method.Service.GetName}}_{{.Method.GetName}}_{{.Index}}); err != nil {
		return nil, nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
{{end}}
{{if .Method.GetServerStreaming}}
	// TODO
{{else}}
	msg, err := server.{{.Method.GetName}}(ctx, &protoReq)
	return msg, nil, err
{{end}}
}`))

	localTrailerTemplate = template.Must(template.New("local-trailer").Parse(`
{{$UseRequestContext := .UseRequestContext}}
{{range $svc := .Services}}
// Register{{$svc.GetName}}{{$.RegisterFuncSuffix}}Server registers the http handlers for service {{$svc.GetName}} to "mux".
// UnaryRPC     :call {{$svc.GetName}}Server directly.
// StreamingRPC :currently unsupported pending https://github.com/grpc/grpc-go/issues/906.
func Register{{$svc.GetName}}{{$.RegisterFuncSuffix}}Server(ctx context.Context, mux *runtime.ServeMux, server {{$svc.GetName}}Server) error {
	{{- if $UseRequestContext }}
	rgctx := req.Context()
	{{- else -}}
	rgctx := ctx
	{{- end }}
	{{range $m := $svc.Methods}}
	{{range $b := $m.Bindings}}
	{{if $m.GetServerStreaming}}
	mux.Handle({{$b.HTTPMethod | printf "%q"}}, pattern_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}, register.HDStreamForward(rgctx, mux, local_request_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}, nil))
	{{- else -}}
	mux.Handle({{$b.HTTPMethod | printf "%q"}}, pattern_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}, register.HDForward(rgctx, mux, local_request_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}},
	{{ if $b.ResponseBody }}
		func(resp proto.Message)proto.Message { return response_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}{resp} }))
	{{- else -}}
		nil))
	{{end}}
	{{- end }}
	{{end}}
	{{end}}
	return nil
}
{{end}}`))

	trailerTemplate = template.Must(template.New("trailer").Parse(`
{{$UseRequestContext := .UseRequestContext}}
{{range $svc := .Services}}
// Register{{$svc.GetName}}{{$.RegisterFuncSuffix}}FromEndpoint is same as Register{{$svc.GetName}}{{$.RegisterFuncSuffix}} but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func Register{{$svc.GetName}}{{$.RegisterFuncSuffix}}FromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	return register.RegFromEndpoint(ctx, mux, endpoint, opts, Register{{$svc.GetName}}{{$.RegisterFuncSuffix}})
}

// Register{{$svc.GetName}}{{$.RegisterFuncSuffix}} registers the http handlers for service {{$svc.GetName}} to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func Register{{$svc.GetName}}{{$.RegisterFuncSuffix}}(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return Register{{$svc.GetName}}{{$.RegisterFuncSuffix}}Client(ctx, mux, New{{$svc.GetName}}Client(conn))
}

// Register{{$svc.GetName}}{{$.RegisterFuncSuffix}}Client registers the http handlers for service {{$svc.GetName}}
// to "mux". The handlers forward requests to the grpc endpoint over the given implementation of "{{$svc.GetName}}Client".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "{{$svc.GetName}}Client"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "{{$svc.GetName}}Client" to call the correct interceptors.
func Register{{$svc.GetName}}{{$.RegisterFuncSuffix}}Client(ctx context.Context, mux *runtime.ServeMux, client {{$svc.GetName}}Client) error {
	{{- if $UseRequestContext }}
	rgctx := req.Context()
	{{- else -}}
	rgctx := ctx
	{{- end }}
	{{range $m := $svc.Methods}}
	{{range $b := $m.Bindings}}
	{{if $m.GetServerStreaming}}
	mux.Handle({{$b.HTTPMethod | printf "%q"}}, pattern_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}, register.HDStreamForward(rgctx, mux, request_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}, nil))
	{{- else -}}
	mux.Handle({{$b.HTTPMethod | printf "%q"}}, pattern_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}, register.HDForward(rgctx, mux, request_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}},
	{{ if $b.ResponseBody }}
		func(resp proto.Message)proto.Message { return response_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}{resp} }))
	{{- else -}}
		nil))
	{{end}}
	{{- end }}
	{{end}}
	{{end}}
	return nil
}

{{range $m := $svc.Methods}}
{{range $b := $m.Bindings}}
{{if $b.ResponseBody}}
type response_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}} struct {
	proto.Message
}

func (m response_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}}) XXX_ResponseBody() interface{} {
	response := m.Message.(*{{$m.ResponseType.GoType $m.Service.File.GoPkg.Path}})
	return {{$b.ResponseBody.AssignableExpr "response"}}
}
{{end}}
{{end}}
{{end}}

var (
	{{range $m := $svc.Methods}}
	{{range $b := $m.Bindings}}
	pattern_{{$svc.GetName}}_{{$m.GetName}}_{{$b.Index}} = runtime.MustPattern(runtime.NewPattern({{$b.PathTmpl.Version}}, {{$b.PathTmpl.OpCodes | printf "%#v"}}, {{$b.PathTmpl.Pool | printf "%#v"}}, {{$b.PathTmpl.Verb | printf "%q"}}, runtime.AssumeColonVerbOpt({{$.AssumeColonVerb}})))
	{{end}}
	{{end}}
)
{{end}}`))
)
