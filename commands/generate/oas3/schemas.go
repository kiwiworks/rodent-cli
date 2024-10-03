package oas3

import (
	"path"
	"strings"

	"github.com/chanced/caps"
	"github.com/dave/jennifer/jen"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"

	"github.com/kiwiworks/rodent/errors"
	"github.com/kiwiworks/rodent/slices"
)

func (g *Generator) generateFile(packageName, name string) *jen.File {
	f := jen.NewFilePathName(packageName, name)
	if !strings.HasPrefix(name, ".go") {
		name += ".go"
	}
	g.files[path.Join(packageName, name)] = f
	f.HeaderComment("Code generated by github.com/kiwiworks/rodent-cli")
	f.HeaderComment("DO NOT EDIT.")

	return f
}

func oas3StringFormatToGoType(stmt *jen.Statement, format string) error {
	switch format {
	case "", "phone", "email":
		stmt.String()
	case "integer":
		stmt.Int64()
	case "uuid":
		stmt.Qual("github.com/google/uuid", "UUID")
	case "ulid":
		stmt.Qual("github.com/oklog/ulid", "ULID")
	case "date", "date-time":
		stmt.Qual("time", "Time")
	case "binary", "byte":
		stmt.Index().Byte()
	case "uri", "url":
		stmt.Qual("net/url", "URL")
	case "ipv4", "ipv6":
		stmt.Qual("net", "IP")
	case "hostname":
		stmt.Qual("net", "IP")
	default:
		return errors.Newf("unsupported `string` format %s", format)
	}
	return nil
}

func oas3NumberFormatToGoType(stmt *jen.Statement, format string) error {
	switch format {
	case "":
		stmt.Float64()
	case "float":
		stmt.Float32()
	case "double":
		stmt.Float64()
	case "int32":
		stmt.Int32()
	case "int64":
		stmt.Int64()
	default:
		return errors.Newf("unsupported `number` format %s", format)
	}
	return nil
}

func oas3ObjectToGoType(stmt *jen.Statement, proxy *base.SchemaProxy, _ *base.Schema) error {
	stmt.Id(path.Base(proxy.GetReference()))
	return nil
}

func oas3TypeToGoType(stmt *jen.Statement, proxy *base.SchemaProxy, schema *base.Schema) error {
	if schema.Type == nil {
		stmt.Any()
		return nil
	}
	kind := schema.Type[0]
	switch kind {
	case "string":
		return oas3StringFormatToGoType(stmt, schema.Format)
	case "number", "integer":
		return oas3NumberFormatToGoType(stmt, schema.Format)
	case "boolean":
		stmt.Bool()
	case "object":
		return oas3ObjectToGoType(stmt, proxy, schema)
	case "array":
		stmt.Index()
		if schema.Items != nil && schema.Items.IsA() {
			proxy = schema.Items.A
			itemSchema, err := proxy.BuildSchema()
			if err != nil {
				return errors.Wrapf(err, "invalid array item schema %s", proxy.GetReference())
			}
			schema = itemSchema
		}
		return oas3TypeToGoType(stmt, proxy, schema)
	case "":
		stmt.Any()
	default:
		return errors.Newf("unsupported type %s", kind)
	}
	return nil
}

func (g *Generator) generateSchemas(schemaProxies *orderedmap.Map[string, *base.SchemaProxy]) error {
	f := g.generateFile("dtos", "dtos")
	for key := range schemaProxies.KeysFromNewest() {
		schemaProxy := schemaProxies.Value(key)
		schema, err := schemaProxy.BuildSchema()
		if err != nil {
			return errors.Wrapf(err, "invalid schema %s", key)
		}
		fields := make([]jen.Code, 0)
		for prop := range schema.Properties.KeysFromNewest() {
			if prop == "$schema" {
				continue
			}
			propSchemaProxy := schema.Properties.Value(prop)
			propSchema, err := propSchemaProxy.BuildSchema()
			if err != nil {
				return errors.Wrapf(err, "invalid property schema %s.%s", key, prop)
			}
			goPropName := caps.ToCamel(prop)
			stmt := jen.Id(goPropName)
			if err = oas3TypeToGoType(stmt, propSchemaProxy, propSchema); err != nil {
				return errors.Wrapf(err, "invalid property type %s.%s (%s)", key, prop, propSchema.Type)
			}
			jsonProps := slices.Of(prop)
			if !slices.Contains(schema.Required, prop) {
				jsonProps = append(jsonProps, "omitempty")
			}
			stmt.Tag(map[string]string{"json": strings.Join(jsonProps, ",")})
			//fmt.Printf("%s.%s -> %s(%s)\n", key, prop, propSchema.Type, propSchema.Format)
			fields = append(fields, stmt)
		}
		f.Type().Id(key).Struct(fields...)
		if err != nil {
			return err
		}
	}
	return nil
}
