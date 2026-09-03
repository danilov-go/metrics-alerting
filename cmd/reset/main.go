package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	nameComment = "// generate:reset"
	nameFile    = "reset.gen.go"
)

const tmplt = `package {{.Name}}

{{range .Data}}
func (rs *{{.Name}}) Reset() {
	if rs == nil {
		return
	}
	{{- range .Fields}}
	{{if .Slice}}rs.{{.Name}} = rs.{{.Name}}[:0]
	{{- else if .Map}}clear(rs.{{.Name}})
	{{- else if .Pointer}}
	{{- if .String}}if rs.{{.Name}} != nil { *rs.{{.Name}} = "" }
	{{- else if .Bool}}if rs.{{.Name}} != nil { *rs.{{.Name}} = false }
	{{- else if .Num}}if rs.{{.Name}} != nil { *rs.{{.Name}} = 0 }
	{{- else}}if resetter, ok := interface{}(rs.{{.Name}}).(interface{ Reset() }); ok && rs.{{.Name}} != nil { 
		resetter.Reset() }
	{{- end}}
	{{- else}}
	{{- if .String}}rs.{{.Name}} = ""
	{{- else if .Bool}}rs.{{.Name}} = false
	{{- else if .Num}}rs.{{.Name}} = 0
	{{- else}}if resetter, ok := interface{}(&rs.{{.Name}}).(interface{ Reset() }); ok { 
		resetter.Reset() }
	{{- end}}
	{{- end}}
	{{- end}}
}
{{end}}`

type pkgData struct {
	Name string
	Data []structData
}

type structData struct {
	Name   string
	Fields []fieldData
}

type fieldData struct {
	Name    string
	Slice   bool
	Map     bool
	Pointer bool
	String  bool
	Bool    bool
	Num     bool
}

func main() {
	paths, err := getPaths(".")
	if err != nil {
		panic(err)
	}
	fset := token.NewFileSet()
	pathData := make(map[string][]structData)
	for _, p := range paths {
		f, err := parser.ParseFile(fset, p, nil, parser.ParseComments)
		if err != nil {
			fmt.Println(err)
			continue
		}
		data := inspectFile(f)
		if len(data) > 0 {
			path := filepath.Dir(p)
			pathData[path] = append(pathData[path], data...)
		}
	}
	writeFiles(pathData)
}

func writeFiles(pathData map[string][]structData) {
	t, err := template.New("reset").Parse(tmplt)
	if err != nil {
		panic(err)
	}
	for path, d := range pathData {
		pkgName := filepath.Base(path)
		data := pkgData{
			Name: pkgName,
			Data: d,
		}
		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			fmt.Println(err)
			continue
		}
		genPath := filepath.Join(path, nameFile)
		err = os.WriteFile(genPath, buf.Bytes(), 0666)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getType(fName string, expr ast.Expr) fieldData {
	data := fieldData{Name: fName}
	if star, ok := expr.(*ast.StarExpr); ok {
		data.Pointer = true
		expr = star.X
	}
	switch t := expr.(type) {
	case *ast.ArrayType:
		data.Slice = true
	case *ast.MapType:
		data.Map = true
	case *ast.Ident:
		switch t.Name {
		case "string":
			data.String = true
		case "bool":
			data.Bool = true
		case "int", "int64", "float64":
			data.Num = true
		}
	}
	return data
}

func getPaths(startPath string) ([]string, error) {
	var paths []string
	err := filepath.WalkDir(startPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".go") {
			paths = append(paths, path)
		}
		return nil
	})
	return paths, err
}

func inspectFile(f *ast.File) []structData {
	var genData []structData
	ast.Inspect(f, func(node ast.Node) bool {
		genDecl, ok := node.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE || genDecl.Doc == nil {
			return true
		}
		comment := false
		for _, c := range genDecl.Doc.List {
			if strings.TrimSpace(c.Text) == nameComment {
				comment = true
				break
			}
		}
		if !comment {
			return true
		}
		for _, specs := range genDecl.Specs {
			typeSpec, ok := specs.(*ast.TypeSpec)
			if !ok {
				continue
			}
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}
			var fields []fieldData
			for _, field := range structType.Fields.List {
				for _, nameIdent := range field.Names {
					data := getType(nameIdent.Name, field.Type)
					fields = append(fields, data)
				}
			}
			genData = append(genData, structData{
				Name:   typeSpec.Name.Name,
				Fields: fields,
			})
		}
		return false
	})
	return genData
}
