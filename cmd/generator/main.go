package main

import (
	"bytes"
	"fmt"
	"github.com/netbox-community/go-netbox/v4"
	"go/format"
	"go/types"
	"golang.org/x/exp/slices"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

//go:generate go run github.com/e-breuninger/terraform-provider-netbox/cmd/generator github.com/e-breuninger/terraform-provider-netbox/internal/provider/datasource_webhook/WebhookModel
func main() {
	// 1. Handle arguments to command
	if len(os.Args) != 2 {
		failErr(fmt.Errorf("expected exactly one argument: <source type>"))
	}
	sourceType := os.Args[1]
	sourceTypePackage, sourceTypeName := splitSourceType(sourceType)

	// 2. Inspect package and use type checker to infer imported types
	pkg := loadPackage(sourceTypePackage)

	// 3. Lookup the given source type name in the package declarations
	obj := pkg.Types.Scope().Lookup(sourceTypeName)
	outputDir := strings.Split(sourceTypePackage, "github.com/e-breuninger/terraform-provider-netbox/")[1]
	if obj == nil {
		failErr(fmt.Errorf("%s not found in declared types of %s",
			sourceTypeName, pkg))
	}

	// 4. We check if it is a declared type
	if _, ok := obj.(*types.TypeName); !ok {
		failErr(fmt.Errorf("%v is not a named type", obj))
	}
	// 5. We expect the underlying type to be a struct
	structType, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		failErr(fmt.Errorf("type %v is not a struct", obj))
	}

	// 6. Now we can iterate through fields and access tags
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		tagValue := structType.Tag(i)
		fmt.Println(field.Name(), tagValue, field.Type())
	}

	err := generate(sourceTypeName, structType, outputDir)
	if err != nil {
		failErr(err)
	}
}

type NetboxMethods interface {
}

func generate(sourceTypeName string, structType *types.Struct, outputDir string) error {
	m := map[string]NetboxMethods{
		"test": netbox.Webhook{},
	}
	val := reflect.ValueOf(m["test"])
	var tags []string
	var nullableTags []string
	var patchedHttpMethod []string
	for i := 0; i < val.Type().NumField(); i++ {
		t := val.Type().Field(i)

		fieldName := t.Name
		fieldType := t.Type
		print(fieldType)
		if fieldType.AssignableTo(reflect.TypeOf((*netbox.NullableString)(nil)).Elem()) {
			print("wow")
			nullableTags = append(nullableTags, fieldName)
		}
		if fieldType.AssignableTo(reflect.TypeOf((*netbox.PatchedWebhookRequestHttpMethod)(nil)).Elem()) {
			print("wow")
			patchedHttpMethod = append(patchedHttpMethod, fieldName)
		}

		tag := t.Tag.Get("json")

		if strings.Contains(tag, "omitempty") {
			print(fieldName + " is omitempty, adding to list\n")
			tags = append(tags, fieldName)
		}
		//print(t.Tag.Get("json"))
	}

	//Start of the code generator
	var buf bytes.Buffer

	//Load the header
	tmpl, err := template.New("tags").Parse(headerTemplate)
	if err != nil {
		failErr(err)
	}
	templateData := struct {
		PackageName string
	}{
		PackageName: "datasource_webhook",
	}
	err = tmpl.Execute(&buf, templateData)
	if err != nil {
		failErr(err)
	}

	//Load the ReadAPI method
	tmpl, err = template.New("readApiHeader").Parse(funcReadApiStart)
	if err != nil {
		failErr(err)
	}

	//print("test" + value.Type().Name())

	// 4. Iterate over struct fields
	var bufVariables bytes.Buffer
	for i := 0; i < structType.NumFields(); i++ {
		field := structType.Field(i)
		if field.Name() == "Tags" {
			tmpl, err := template.New("tags").Parse(tagsTemplate)
			if err != nil {
				return err
			}
			err = tmpl.Execute(&bufVariables, nil)
			if err != nil {
				return err
			}

		} else if field.Name() == "CustomFields" {
			tmpl, err := template.New("customfields").Parse(customFieldsTemplate)
			if err != nil {
				return err
			}
			err = tmpl.Execute(&bufVariables, nil)
			if err != nil {
				return err
			}
		} else {

			// Generate code for each changeset field
			switch v := field.Type().(type) {
			case *types.Named:
				typeName := v.Obj()
				if slices.Contains(nullableTags, field.Name()) {
					tmpl, err := template.New("nullable_string").Parse(nullableStringTemplate)
					if err != nil {
						return err
					}
					templateData := struct {
						Variable string
					}{
						Variable: field.Name(),
					}
					err = tmpl.Execute(&bufVariables, templateData)
					if err != nil {
						return err
					}
				} else {
					tmpl, err := template.New("variable").Parse(variableTemplate)
					if err != nil {
						return err
					}
					templateData := struct {
						VariableName string
						Type         string
					}{
						VariableName: field.Name(),
						Type:         typeName.Name(),
					}

					if slices.Contains(tags, field.Name()) {
						templateData.Type = "StringPointerValue"
					}
					err = tmpl.Execute(&bufVariables, templateData)
					if err != nil {
						return err
					}
				}

			default:
				return fmt.Errorf("struct field type not hanled: %T", v)
			}

		}

	}

	readapiHeader := struct {
		DataModel string
		ApiMethod string
		Variables string
	}{
		DataModel: sourceTypeName,
		ApiMethod: val.Type().Name(),
		Variables: bufVariables.String(),
	}
	err = tmpl.Execute(&buf, readapiHeader)
	if err != nil {
		failErr(err)
	}

	//Format the source data
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		failErr(err)
	}
	print(formatted)

	//Create the file
	f, err := os.Create(filepath.Join(outputDir, fmt.Sprintf("%s_gen2.go", sourceTypeName)))
	if err != nil {
		return err
	}
	_, err = f.Write(formatted)
	if err != nil {
		return err
	}

	// 5. Generate changeset type

	// 6. Build the target file name

	// 7. Write generated file
	return nil
}

func loadPackage(path string) *packages.Package {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedImports}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		failErr(fmt.Errorf("loading packages for inspection: %v", err))
	}
	if packages.PrintErrors(pkgs) > 0 {
		os.Exit(1)
	}

	return pkgs[0]
}

func splitSourceType(sourceType string) (string, string) {
	idx := strings.LastIndexByte(sourceType, '.')
	if idx == -1 {
		failErr(fmt.Errorf(`expected qualified type as "pkg/path.MyType"`))
	}
	sourceTypePackage := sourceType[0:idx]
	sourceTypeName := sourceType[idx+1:]
	return sourceTypePackage, sourceTypeName
}

func failErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
