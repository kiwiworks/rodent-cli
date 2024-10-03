package oas3

import (
	"github.com/dave/jennifer/jen"
	"github.com/pb33f/libopenapi/datamodel/high/v3"
)

const (
	sdkPackage = "github.com/kiwiworks/rodent/web/sdk"
	optPackage = "github.com/kiwiworks/rodent/system/opt"
	errPackage = "github.com/kiwiworks/rodent/errors"
)

func (g *Generator) generateClient(document v3.Document) error {
	f := g.generateFile("", "client")

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
