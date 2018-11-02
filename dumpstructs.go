package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/urfave/cli"
)

func joinFieldNames(Names []*ast.Ident) string {
	var ret string
	for i, n := range Names {
		if i > 0 {
			ret += " "
		}
		ret += n.Name
	}
	return strings.TrimSpace(ret)
}

// handleFieldTypes is a simple recursive function to decipher and convert the
// field type to a human-readable string.
func handleFieldTypes(r string, typ ast.Expr) string {
	switch typ.(type) {
	case *ast.FuncType:
		r += "func("
		for i, f := range genFieldList(typ.(*ast.FuncType).Params.List) {
			if i > 0 {
				r += ", "
			}
			r += f.Types
		}
		r += ") "
		if typ.(*ast.FuncType).Results == nil {
			return r
		} else {
			r += "("
		}
		for i, f := range genFieldList(typ.(*ast.FuncType).Results.List) {
			if i > 0 {
				r += ", "
			}
			r += f.Types
		}
		r += ")"
		return r
	case *ast.MapType:
		return handleFieldTypes(r+fmt.Sprintf("map[%s]", typ.(*ast.MapType).Key), typ.(*ast.MapType).Value)
	case *ast.ArrayType:
		return handleFieldTypes(r+"[]", typ.(*ast.ArrayType).Elt)
	case *ast.InterfaceType:
		return r + "interface{}"
	case *ast.SelectorExpr:
		return r + fmt.Sprintf("%s", typ.(*ast.SelectorExpr).X) + "." + typ.(*ast.SelectorExpr).Sel.Name
	case *ast.StarExpr:
		return handleFieldTypes(r+"*", typ.(*ast.StarExpr).X)
	case ast.Expr:
		return r + fmt.Sprintf("%s", typ)
	default:
		return "TODO"
	}
}

func dumpStructs(fset *token.FileSet, file string) {
	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}
	ast.Inspect(node, func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		fn, err := filepath.Abs(file)
		if err != nil {
			log.Fatal(err)
		}
		loc := fmt.Sprintf("// %s - line: %d", fn, fset.Position(t.Pos()).Line)

		if t.Type == nil {
			return true
		}

		s, ok := t.Type.(*ast.StructType)
		if !ok {
			return true
		}

		fmt.Printf("%s\n", loc)
		fmt.Printf("type %s struct {\n", t.Name.Name)
		for _, field := range genFieldList(s.Fields.List) {
			fmt.Println("\t" + fmt.Sprintf("%s %s %s %s",
				field.Names, field.Types, field.Tags, field.Comments))
		}
		fmt.Println("}")
		return true
	})
}

func dump(userpath string, exclude string) {
	fset := token.NewFileSet()
	var err error
	err = filepath.Walk(userpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if exclude != "" {
			skip, _ := regexp.MatchString(exclude, path)
			if skip {
				return nil
			}
		}
		if strings.HasSuffix(path, ".go") {
			dumpStructs(fset, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("Error walking the path: %v\n", err)
		return
	}
}

type fieldListStr struct {
	Names    string
	Types    string
	Tags     string
	Comments string
}

func genFieldList(fields []*ast.Field) []fieldListStr {
	ret := make([]fieldListStr, 0)
	for _, field := range fields {
		tag := ""
		if field.Tag != nil {
			tag = field.Tag.Value
		}
		cmmt := ""
		if field.Comment != nil {
			cmmt = "/* " + strings.TrimSpace(field.Comment.Text()) + " */"
		}
		ret = append(ret, fieldListStr{
			Names:    joinFieldNames(field.Names),
			Types:    handleFieldTypes("", field.Type),
			Tags:     tag,
			Comments: cmmt,
		})
	}
	return ret
}

func main() {
	app := cli.NewApp()
	app.Name = "dumpstructs"
	app.Usage = ""
	app.Version = "0.0.1"
	app.UsageText = fmt.Sprintf("%s [options]", app.Name)

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "path, p",
			Value: ".",
			Usage: "Path to traverse to discover Go files. Optional.",
		},
		cli.StringFlag{
			Name:  "exclude, x",
			Value: "",
			Usage: "Regex pattern to use to exclude paths from list. Optional.",
		},
	}

	app.Action = func(c *cli.Context) error {
		dump(c.String("path"), c.String("exclude"))
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
