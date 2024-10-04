package oas3

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/chanced/caps"
	"github.com/dave/jennifer/jen"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/datamodel/high/v3"

	"github.com/kiwiworks/rodent/errors"
	"github.com/kiwiworks/rodent/logger"
	"github.com/kiwiworks/rodent/logger/props"
	"github.com/kiwiworks/rodent/slices"
)

const (
	sdkPackage = "github.com/kiwiworks/rodent/web/sdk"
	optPackage = "github.com/kiwiworks/rodent/system/opt"
	errPackage = "github.com/kiwiworks/rodent/errors"
)

func (g *Generator) generateClient(document v3.Document, f *jen.File) error {
	f.Type().Id("Client").StructFunc(func(group *jen.Group) {
		group.Op("*").Qual(sdkPackage, "Client")
	})

	g.imports = append(g.imports, Import{
		Package: "github.com/kiwiworks/rodent",
		Version: "latest",
	})

	f.Comment("NewClient creates a new client from the given string endpoint, and optional options.")
	f.Func().
		Id("NewClient").
		Params(
			jen.Id("endpoint").String(),
			jen.Id("opts").Op("...").
				Qual(optPackage, "Option").
				Index(jen.Qual(sdkPackage, "Config")),
		).
		Parens(jen.List(jen.Op("*").Id("Client"), jen.Error())).
		BlockFunc(func(group *jen.Group) {
			if document.Servers != nil || len(document.Servers) > 0 && document.Servers[0].URL != "" {
				group.If(jen.Id("endpoint").Op("==").Lit("")).BlockFunc(func(group *jen.Group) {
					group.Id("endpoint").Op("=").Lit(document.Servers[0].URL)
				})
			}
			group.List(jen.Id("c"), jen.Id("err")).Op(":=").
				Qual(sdkPackage, "New").
				Call(jen.Id("endpoint"), jen.Id("opts").Op("..."))
			group.If(jen.Err().Op("!=").Nil()).
				BlockFunc(func(group *jen.Group) {
					group.Return(jen.Nil(), jen.Qual(errPackage, "Wrapf").Call(
						jen.Id("err"),
						jen.Lit("failed to create client from endpoint %s"),
						jen.Id("endpoint"),
					))
				})
			group.ReturnFunc(func(group *jen.Group) {
				group.Op("&").Id("Client").Values(jen.Dict{
					jen.Id("Client"): jen.Id("c"),
				})
				group.Nil()
			})
		})
	return nil
}

func methodNameFromOperation(method string, operation *v3.Operation) string {
	operationId := operation.OperationId
	if operationId == "" && len(operation.Tags) > 0 {
		operationId = caps.ToCamel(method) + caps.ToCamel(operation.Tags[0])
	}
	fragments := strings.FieldsFunc(operationId, func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
	fragments = slices.Map(fragments, func(in string) string {
		return caps.ToCamel(in)
	})
	return strings.Join(fragments, "")
}

func schemaByResponseContent(response *v3.Response, content string) (*base.SchemaProxy, error) {
	if mediaType := response.Content.Value(content); mediaType != nil {
		return mediaType.Schema, nil
	}
	return nil, nil
}

func operationResponse(operation *v3.Operation, content string) (*base.SchemaProxy, error) {
	if operation.Responses == nil {
		return nil, nil
	}
	if operation.Responses.Codes == nil {
		return nil, errors.Newf("no response codes found for operation %s", operation.OperationId)
	}
	if response := operation.Responses.FindResponseByCode(200); response != nil {
		return schemaByResponseContent(response, content)
	}
	if response := operation.Responses.FindResponseByCode(201); response != nil {
		return schemaByResponseContent(response, content)
	}
	if response := operation.Responses.FindResponseByCode(202); response != nil {
		return schemaByResponseContent(response, content)
	}
	if operation.Responses.Default != nil {
		return schemaByResponseContent(operation.Responses.Default, content)
	}
	return nil, errors.Newf("no valid response found for operation %s with content type %s from any of the successful http status code", operation.OperationId, content)
}

func (g *Generator) generateClientMethod(f *jen.File, method, apiPath string, operation *v3.Operation) error {
	if operation == nil {
		return nil
	}

	log := logger.New().With(props.HttpMethod(method), props.HttpPath(apiPath))
	methodName := methodNameFromOperation(method, operation)

	params := slices.Of[jen.Code](jen.Id("ctx").Qual("context", "Context"))

	dtoPackage := path.Join(g.moduleName, "dtos")
	if operation.RequestBody != nil {
		request := operation.RequestBody.Content.Value("application/json")
		if request == nil && slices.Contains(slices.Of("POST", "PUT", "PATCH"), strings.ToUpper(method)) {
			log.Warn("no request body found for operation with content type application/json")
			return nil
		}
		params = append(params, jen.Id("req").Op("*").Qual(dtoPackage, filepath.Base(request.Schema.GetReference())))
	}

	response, err := operationResponse(operation, "application/json")
	if err != nil {
		return errors.Wrapf(err, "failed to generate client method for %s", apiPath)
	}
	if response == nil {
		log.Warn("no default response found for operation with content type application/json")
		return nil
	}

	var result jen.Code
	var executeResult jen.Code
	responseType := filepath.Base(response.GetReference())
	schema, err := response.BuildSchema()
	if err != nil {
		return errors.Wrapf(err, "failed to build response schema for %s", apiPath)
	}
	if schema.Items != nil && schema.Items.IsA() {
		result = jen.List(jen.Index().Qual(dtoPackage, filepath.Base(schema.Items.A.GetReference())))
		executeResult = result
	} else if responseType != "" && schema.Type[0] == "object" {
		result = jen.Op("*").Qual(dtoPackage, responseType)
		executeResult = jen.Qual(dtoPackage, responseType)
	} else {
		stmt := jen.Op("")
		if err := oas3TypeToGoType(stmt, response, schema); err != nil {
			return errors.Wrapf(err, "failed to generate client method for %s, invalid response type", apiPath)
		}
		result = stmt
		executeResult = stmt
	}

	params = append(params, jen.Id("opts").Op("...").Qual(optPackage, "Option").Index(jen.Qual(sdkPackage, "Request")))

	f.Commentf("%s performs the %s %s operation.\n%s", methodName, method, apiPath, operation.Description)
	f.Func().Params(jen.Id("c").Op("*").Id("Client")).Id(methodName).
		Params(params...).
		Parens(jen.List(result, jen.Error())).
		BlockFunc(func(group *jen.Group) {
			group.Id("request").Op(":=").Id("c").Dot("Request").CallFunc(func(group *jen.Group) {
				group.Lit(method)
				group.Lit(apiPath)
				group.Id("opts").Op("...")
			})
			group.List(jen.Id("response"), jen.Err()).Op(":=").Qual(sdkPackage, "Execute").Index(executeResult).
				CallFunc(func(group *jen.Group) {
					group.Id("ctx")
					group.Op("*").Id("c")
					group.Id("request")
				})
			group.If(jen.Err().Op("!=").Nil()).BlockFunc(func(group *jen.Group) {
				group.Return(jen.Nil(), jen.Qual(errPackage, "Wrapf").CallFunc(func(group *jen.Group) {
					group.Id("err")
					group.Lit(fmt.Sprintf("failed to execute %s %s operation", method, apiPath))
				}))
			})
			group.Return(jen.Id("response"), jen.Nil())
		})
	return nil
}

func (g *Generator) generateClientPackage(document v3.Document) error {
	f := g.generateFile("", "client")
	if err := g.generateClient(document, f); err != nil {
		return errors.Wrapf(err, "failed to generate client")
	}

	for apiPath := range document.Paths.PathItems.KeysFromNewest() {
		pathItem := document.Paths.PathItems.Value(apiPath)
		if err := g.generateClientMethod(f, "GET", apiPath, pathItem.Get); err != nil {
			return errors.Wrapf(err, "failed to generate client GET method for %s", apiPath)
		}
		if err := g.generateClientMethod(f, "POST", apiPath, pathItem.Post); err != nil {
			return errors.Wrapf(err, "failed to generate client POST method for %s", apiPath)
		}
		if err := g.generateClientMethod(f, "PUT", apiPath, pathItem.Put); err != nil {
			return errors.Wrapf(err, "failed to generate client PUT method for %s", apiPath)
		}
		if err := g.generateClientMethod(f, "PATCH", apiPath, pathItem.Patch); err != nil {
			return errors.Wrapf(err, "failed to generate client PATCH method for %s", apiPath)
		}
		if err := g.generateClientMethod(f, "DELETE", apiPath, pathItem.Delete); err != nil {
			return errors.Wrapf(err, "failed to generate client DELETE method for %s", apiPath)
		}
		if err := g.generateClientMethod(f, "HEAD", apiPath, pathItem.Head); err != nil {
			return errors.Wrapf(err, "failed to generate client HEAD method for %s", apiPath)
		}
		if err := g.generateClientMethod(f, "OPTIONS", apiPath, pathItem.Options); err != nil {
			return errors.Wrapf(err, "failed to generate client OPTIONS method for %s", apiPath)
		}
		if err := g.generateClientMethod(f, "TRACE", apiPath, pathItem.Trace); err != nil {
			return errors.Wrapf(err, "failed to generate client TRACE method for %s", apiPath)
		}
	}

	return nil
}
