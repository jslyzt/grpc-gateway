package descriptor

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	gogen "github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/jslyzt/grpc-gateway/protoc-gen-grpc-gateway/httprule"
)

// IsWellKnownType returns true if the provided fully qualified type name is considered 'well-known'.
func IsWellKnownType(typeName string) bool {
	_, ok := wellKnownTypeConv[typeName]
	return ok
}

// GoPackage represents a golang package
type GoPackage struct {
	Path  string // the package path to the package.
	Name  string // the package name of the package
	Alias string // an alias of the package unique within the current invokation of grpc-gateway generator
}

// Standard returns whether the import is a golang standard package.
func (p GoPackage) Standard() bool {
	return !strings.Contains(p.Path, ".")
}

// String returns a string representation of this package in the form of import line in golang.
func (p GoPackage) String() string {
	if p.Alias == "" {
		return fmt.Sprintf("%q", p.Path)
	}
	return fmt.Sprintf("%s %q", p.Alias, p.Path)
}

// File wraps descriptor.FileDescriptorProto for richer features.
type File struct {
	*descriptor.FileDescriptorProto
	GoPkg    GoPackage  // the go package of the go file generated from this file..
	Messages []*Message // the list of messages defined in this file.
	Enums    []*Enum    // the list of enums defined in this file.
	Services []*Service // the list of services defined in this file.
}

// proto2 determines if the syntax of the file is proto2.
func (f *File) proto2() bool {
	return f.Syntax == nil || f.GetSyntax() == "proto2"
}

// Message describes a protocol buffer message types
type Message struct {
	File   *File    // the file where the message is defined
	Outers []string // a list of outer messages if this message is a nested type.
	*descriptor.DescriptorProto
	Fields []*Field
	Index  int // proto path index of this message in File.
}

// FQMN returns a fully qualified message name of this message.
func (m *Message) FQMN() string {
	components := []string{""}
	if m.File.Package != nil {
		components = append(components, m.File.GetPackage())
	}
	components = append(components, m.Outers...)
	components = append(components, m.GetName())
	return strings.Join(components, ".")
}

// GoType returns a go type name for the message type.
// It prefixes the type name with the package alias if its belonging package is not "currentPackage".
func (m *Message) GoType(currentPackage string) string {
	var components []string
	components = append(components, m.Outers...)
	components = append(components, m.GetName())

	name := strings.Join(components, "_")
	if m.File.GoPkg.Path == currentPackage {
		return name
	}
	pkg := m.File.GoPkg.Name
	if alias := m.File.GoPkg.Alias; alias != "" {
		pkg = alias
	}
	return fmt.Sprintf("%s.%s", pkg, name)
}

// Enum describes a protocol buffer enum types
type Enum struct {
	File   *File    // the file where the enum is defined
	Outers []string // a list of outer messages if this enum is a nested type.
	*descriptor.EnumDescriptorProto
	Index int
}

// FQEN returns a fully qualified enum name of this enum.
func (e *Enum) FQEN() string {
	components := []string{""}
	if e.File.Package != nil {
		components = append(components, e.File.GetPackage())
	}
	components = append(components, e.Outers...)
	components = append(components, e.GetName())
	return strings.Join(components, ".")
}

// GoType returns a go type name for the enum type.
// It prefixes the type name with the package alias if its belonging package is not "currentPackage".
func (e *Enum) GoType(currentPackage string) string {
	var components []string
	components = append(components, e.Outers...)
	components = append(components, e.GetName())

	name := strings.Join(components, "_")
	if e.File.GoPkg.Path == currentPackage {
		return name
	}
	pkg := e.File.GoPkg.Name
	if alias := e.File.GoPkg.Alias; alias != "" {
		pkg = alias
	}
	return fmt.Sprintf("%s.%s", pkg, name)
}

// Service wraps descriptor.ServiceDescriptorProto for richer features.
type Service struct {
	File *File // the file where this service is defined
	*descriptor.ServiceDescriptorProto
	Methods []*Method // the list of methods defined in this service.
}

// FQSN returns the fully qualified service name of this service.
func (s *Service) FQSN() string {
	components := []string{""}
	if s.File.Package != nil {
		components = append(components, s.File.GetPackage())
	}
	components = append(components, s.GetName())
	return strings.Join(components, ".")
}

// Method wraps descriptor.MethodDescriptorProto for richer features.
type Method struct {
	Service *Service // the service which this method belongs to
	*descriptor.MethodDescriptorProto
	RequestType  *Message // the message type of requests to this method
	ResponseType *Message // the message type of responses from this method
	Bindings     []*Binding
}

// FQMN returns a fully qualified rpc method name of this method.
func (m *Method) FQMN() string {
	components := []string{}
	components = append(components, m.Service.FQSN())
	components = append(components, m.GetName())
	return strings.Join(components, ".")
}

// Binding describes how an HTTP endpoint is bound to a gRPC method.
type Binding struct {
	Method       *Method           // the method which the endpoint is bound to
	Index        int               // a zero-origin index of the binding in the target method
	PathTmpl     httprule.Template // path template where this method is mapped to
	HTTPMethod   string            // the HTTP method which this method is mapped to
	PathParams   []Parameter       // the list of parameters provided in HTTP request paths
	Body         *Body             // describes parameters provided in HTTP request body
	ResponseBody *Body             // describes field in response struct to marshal in HTTP response body
}

// ExplicitParams returns a list of explicitly bound parameters of "b",
// i.e. a union of field path for body and field paths for path parameters.
func (b *Binding) ExplicitParams() []string {
	var result []string
	if b.Body != nil {
		result = append(result, b.Body.FieldPath.String())
	}
	for _, p := range b.PathParams {
		result = append(result, p.FieldPath.String())
	}
	return result
}

// Field wraps descriptor.FieldDescriptorProto for richer features.
type Field struct {
	Message      *Message // the message type which this field belongs to
	FieldMessage *Message // the message type of the field
	*descriptor.FieldDescriptorProto
}

// Parameter is a parameter provided in http requests
type Parameter struct {
	FieldPath         // a path to a proto field which this parameter is mapped to
	Target    *Field  // the proto field which this parameter is mapped to
	Method    *Method // the method which this parameter is used for
}

// ConvertFuncExpr returns a go expression of a converter function.
// The converter function converts a string into a value for the parameter.
func (p Parameter) ConvertFuncExpr() (string, error) {
	tbl := proto3ConvertFuncs
	if !p.IsProto2() && p.IsRepeated() {
		tbl = proto3RepeatedConvertFuncs
	} else if p.IsProto2() && !p.IsRepeated() {
		tbl = proto2ConvertFuncs
	} else if p.IsProto2() && p.IsRepeated() {
		tbl = proto2RepeatedConvertFuncs
	}
	typ := p.Target.GetType()
	conv, ok := tbl[typ]
	if !ok {
		conv, ok = wellKnownTypeConv[p.Target.GetTypeName()]
	}
	if !ok {
		return "", fmt.Errorf("unsupported field type %s of parameter %s in %s.%s", typ, p.FieldPath, p.Method.Service.GetName(), p.Method.GetName())
	}
	return conv, nil
}

// IsEnum returns true if the field is an enum type, otherwise false is returned.
func (p Parameter) IsEnum() bool {
	return p.Target.GetType() == descriptor.FieldDescriptorProto_TYPE_ENUM
}

// IsRepeated returns true if the field is repeated, otherwise false is returned.
func (p Parameter) IsRepeated() bool {
	return p.Target.GetLabel() == descriptor.FieldDescriptorProto_LABEL_REPEATED
}

// IsProto2 returns true if the field is proto2, otherwise false is returned.
func (p Parameter) IsProto2() bool {
	return p.Target.Message.File.proto2()
}

// Body describes a http (request|response) body to be sent to the (method|client).
// This is used in body and response_body options in google.api.HttpRule
type Body struct {
	FieldPath FieldPath // a path to a proto field which the (request|response) body is mapped to.
}

// AssignableExpr returns an assignable expression in Go to be used to initialize method request object.
// It starts with "msgExpr", which is the go expression of the method request object.
func (b Body) AssignableExpr(msgExpr string) string {
	return b.FieldPath.AssignableExpr(msgExpr)
}

// FieldPath is a path to a field from a request message.
type FieldPath []FieldPathComponent

// String returns a string representation of the field path.
func (p FieldPath) String() string {
	var components []string
	for _, c := range p {
		components = append(components, c.Name)
	}
	return strings.Join(components, ".")
}

// IsNestedProto3 indicates whether the FieldPath is a nested Proto3 path.
func (p FieldPath) IsNestedProto3() bool {
	if len(p) > 1 && !p[0].Target.Message.File.proto2() {
		return true
	}
	return false
}

// AssignableExpr is an assignable expression in Go to be used to assign a value to the target field.
// It starts with "msgExpr", which is the go expression of the method request object.
func (p FieldPath) AssignableExpr(msgExpr string) string {
	l := len(p)
	if l == 0 {
		return msgExpr
	}

	var preparations []string
	components := msgExpr
	for i, c := range p {
		// Check if it is a oneOf field.
		if c.Target.OneofIndex != nil {
			index := c.Target.OneofIndex
			msg := c.Target.Message
			oneOfName := gogen.CamelCase(msg.GetOneofDecl()[*index].GetName())
			oneofFieldName := msg.GetName() + "_" + c.AssignableExpr()

			components = components + "." + oneOfName
			s := `if %s == nil {
				%s =&%s{}
			} else if _, ok := %s.(*%s); !ok {
				return nil, metadata, grpc.Errorf(codes.InvalidArgument, "expect type: *%s, but: %%t\n",%s)
			}`

			preparations = append(preparations, fmt.Sprintf(s, components, components, oneofFieldName, components, oneofFieldName, oneofFieldName, components))
			components = components + ".(*" + oneofFieldName + ")"
		}

		if i == l-1 {
			components = components + "." + c.AssignableExpr()
			continue
		}
		components = components + "." + c.ValueExpr()
	}

	preparations = append(preparations, components)
	return strings.Join(preparations, "\n")
}

// FieldPathComponent is a path component in FieldPath
type FieldPathComponent struct {
	Name   string // a name of the proto field which this component corresponds to.
	Target *Field // the proto field which this component corresponds to.
}

// AssignableExpr returns an assignable expression in go for this field.
func (c FieldPathComponent) AssignableExpr() string {
	return gogen.CamelCase(c.Name)
}

// ValueExpr returns an expression in go for this field.
func (c FieldPathComponent) ValueExpr() string {
	if c.Target.Message.File.proto2() {
		return fmt.Sprintf("Get%s()", gogen.CamelCase(c.Name))
	}
	return gogen.CamelCase(c.Name)
}

var (
	proto3ConvertFuncs = map[descriptor.FieldDescriptorProto_Type]string{
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:  "Float64",
		descriptor.FieldDescriptorProto_TYPE_FLOAT:   "Float32",
		descriptor.FieldDescriptorProto_TYPE_INT64:   "Int64",
		descriptor.FieldDescriptorProto_TYPE_UINT64:  "Uint64",
		descriptor.FieldDescriptorProto_TYPE_INT32:   "Int32",
		descriptor.FieldDescriptorProto_TYPE_FIXED64: "Uint64",
		descriptor.FieldDescriptorProto_TYPE_FIXED32: "Uint32",
		descriptor.FieldDescriptorProto_TYPE_BOOL:    "Bool",
		descriptor.FieldDescriptorProto_TYPE_STRING:  "String",
		// FieldDescriptorProto_TYPE_GROUP
		// FieldDescriptorProto_TYPE_MESSAGE
		descriptor.FieldDescriptorProto_TYPE_BYTES:    "Bytes",
		descriptor.FieldDescriptorProto_TYPE_UINT32:   "Uint32",
		descriptor.FieldDescriptorProto_TYPE_ENUM:     "Enum",
		descriptor.FieldDescriptorProto_TYPE_SFIXED32: "Int32",
		descriptor.FieldDescriptorProto_TYPE_SFIXED64: "Int64",
		descriptor.FieldDescriptorProto_TYPE_SINT32:   "Int32",
		descriptor.FieldDescriptorProto_TYPE_SINT64:   "Int64",
	}

	proto3RepeatedConvertFuncs = map[descriptor.FieldDescriptorProto_Type]string{
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:  "Float64Slice",
		descriptor.FieldDescriptorProto_TYPE_FLOAT:   "Float32Slice",
		descriptor.FieldDescriptorProto_TYPE_INT64:   "Int64Slice",
		descriptor.FieldDescriptorProto_TYPE_UINT64:  "Uint64Slice",
		descriptor.FieldDescriptorProto_TYPE_INT32:   "Int32Slice",
		descriptor.FieldDescriptorProto_TYPE_FIXED64: "Uint64Slice",
		descriptor.FieldDescriptorProto_TYPE_FIXED32: "Uint32Slice",
		descriptor.FieldDescriptorProto_TYPE_BOOL:    "BoolSlice",
		descriptor.FieldDescriptorProto_TYPE_STRING:  "StringSlice",
		// FieldDescriptorProto_TYPE_GROUP
		// FieldDescriptorProto_TYPE_MESSAGE
		descriptor.FieldDescriptorProto_TYPE_BYTES:    "BytesSlice",
		descriptor.FieldDescriptorProto_TYPE_UINT32:   "Uint32Slice",
		descriptor.FieldDescriptorProto_TYPE_ENUM:     "EnumSlice",
		descriptor.FieldDescriptorProto_TYPE_SFIXED32: "Int32Slice",
		descriptor.FieldDescriptorProto_TYPE_SFIXED64: "Int64Slice",
		descriptor.FieldDescriptorProto_TYPE_SINT32:   "Int32Slice",
		descriptor.FieldDescriptorProto_TYPE_SINT64:   "Int64Slice",
	}

	proto2ConvertFuncs = map[descriptor.FieldDescriptorProto_Type]string{
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:  "Float64P",
		descriptor.FieldDescriptorProto_TYPE_FLOAT:   "Float32P",
		descriptor.FieldDescriptorProto_TYPE_INT64:   "Int64P",
		descriptor.FieldDescriptorProto_TYPE_UINT64:  "Uint64P",
		descriptor.FieldDescriptorProto_TYPE_INT32:   "Int32P",
		descriptor.FieldDescriptorProto_TYPE_FIXED64: "Uint64P",
		descriptor.FieldDescriptorProto_TYPE_FIXED32: "Uint32P",
		descriptor.FieldDescriptorProto_TYPE_BOOL:    "BoolP",
		descriptor.FieldDescriptorProto_TYPE_STRING:  "StringP",
		// FieldDescriptorProto_TYPE_GROUP
		// FieldDescriptorProto_TYPE_MESSAGE
		// FieldDescriptorProto_TYPE_BYTES
		// TODO(yugui) Handle bytes
		descriptor.FieldDescriptorProto_TYPE_UINT32:   "Uint32P",
		descriptor.FieldDescriptorProto_TYPE_ENUM:     "EnumP",
		descriptor.FieldDescriptorProto_TYPE_SFIXED32: "Int32P",
		descriptor.FieldDescriptorProto_TYPE_SFIXED64: "Int64P",
		descriptor.FieldDescriptorProto_TYPE_SINT32:   "Int32P",
		descriptor.FieldDescriptorProto_TYPE_SINT64:   "Int64P",
	}

	proto2RepeatedConvertFuncs = map[descriptor.FieldDescriptorProto_Type]string{
		descriptor.FieldDescriptorProto_TYPE_DOUBLE:  "Float64Slice",
		descriptor.FieldDescriptorProto_TYPE_FLOAT:   "Float32Slice",
		descriptor.FieldDescriptorProto_TYPE_INT64:   "Int64Slice",
		descriptor.FieldDescriptorProto_TYPE_UINT64:  "Uint64Slice",
		descriptor.FieldDescriptorProto_TYPE_INT32:   "Int32Slice",
		descriptor.FieldDescriptorProto_TYPE_FIXED64: "Uint64Slice",
		descriptor.FieldDescriptorProto_TYPE_FIXED32: "Uint32Slice",
		descriptor.FieldDescriptorProto_TYPE_BOOL:    "BoolSlice",
		descriptor.FieldDescriptorProto_TYPE_STRING:  "StringSlice",
		// FieldDescriptorProto_TYPE_GROUP
		// FieldDescriptorProto_TYPE_MESSAGE
		// FieldDescriptorProto_TYPE_BYTES
		// TODO(maros7) Handle bytes
		descriptor.FieldDescriptorProto_TYPE_UINT32:   "Uint32Slice",
		descriptor.FieldDescriptorProto_TYPE_ENUM:     "EnumSlice",
		descriptor.FieldDescriptorProto_TYPE_SFIXED32: "Int32Slice",
		descriptor.FieldDescriptorProto_TYPE_SFIXED64: "Int64Slice",
		descriptor.FieldDescriptorProto_TYPE_SINT32:   "Int32Slice",
		descriptor.FieldDescriptorProto_TYPE_SINT64:   "Int64Slice",
	}

	wellKnownTypeConv = map[string]string{
		".google.protobuf.Timestamp":   "Timestamp",
		".google.protobuf.Duration":    "Duration",
		".google.protobuf.StringValue": "StringValue",
		".google.protobuf.FloatValue":  "FloatValue",
		".google.protobuf.DoubleValue": "DoubleValue",
		".google.protobuf.BoolValue":   "BoolValue",
		".google.protobuf.BytesValue":  "BytesValue",
		".google.protobuf.Int32Value":  "Int32Value",
		".google.protobuf.UInt32Value": "UInt32Value",
		".google.protobuf.Int64Value":  "Int64Value",
		".google.protobuf.UInt64Value": "UInt64Value",
	}
)
