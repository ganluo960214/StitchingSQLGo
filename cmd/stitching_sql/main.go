package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/ganluo960214/StringCase"
	"github.com/go-playground/validator/v10"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

/*
env
*/
var (
	GOFILE    = os.Getenv("GOFILE")
	GOPACKAGE = os.Getenv("GOPACKAGE")
)

/*
validator
*/
var (
	validate = validator.New()
)

/*
flags
*/
var (
	flags = struct {
		Type        string `validate:"required"`
		Table       string
		FileName    string
		IsAddImport bool
	}{}
)

func init() {
	flag.StringVar(&flags.Type, "type", "", "type name; must be set")
	flag.StringVar(&flags.Table, "table", "", "table name in database; default as \"-type\"")
	flag.StringVar(&flags.FileName, "file-name", "", "newly generated file name; default as \"-type_stitching_sql.go\"") // flag -file-name-type
	flag.BoolVar(&flags.IsAddImport, "is-add-import", true, "only for develop, ignore it.")
}

/*
init log and flag
*/
func init() {
	log.SetFlags(0)
	flag.Parse()
}

/*
file template
*/
type FileTemplateContent struct {
	Type        string
	Package     string
	Table       string
	FieldMapper map[string]string // map[struct field]table field
	IsAddImport bool
}

const FileTemplate = `// Code generated by "gml -type={{.Type}}"; DO NOT EDIT.

package {{.Package}}

import(
	"database/sql"
	"errors"
	{{if .IsAddImport}}"github.com/ganLuo960214/StitchingSQLGo"{{end}}
)

// --- table ---
func ({{.Type}}) Table(s *{{if $.IsAddImport}}StitchingSQLGo.{{end}}SqlBuilder) error {
	if s == nil {
		return {{if $.IsAddImport}}StitchingSQLGo.{{end}}ErrorNilSQL
	}

	s.WriteString(" {{.Table}}")
	return nil
}

// --- field ---
type generate{{.Type}}field struct {
	{{range $k,$v:= .FieldMapper}}
		{{$k}} {{if $.IsAddImport}}StitchingSQLGo.{{end}}Field{{end}}
}

var (
	{{.Type}}_F = generate{{.Type}}field { {{range $k,$v:= .FieldMapper}}
		{{$k}}: generate{{$.Type}}field{{$k}}{},{{end}}
	}
)

{{range $k,$v:= .FieldMapper}}
// --- generate {{$.Table}} table {{$v}} Field
type generate{{$.Type}}field{{$k}} struct{}
func (generate{{$.Type}}field{{$k}}) Field(s *{{if $.IsAddImport}}StitchingSQLGo.{{end}}SqlBuilder,isRefTable bool) error {
	if s == nil {
		return {{if $.IsAddImport}}StitchingSQLGo.{{end}}ErrorNilSQL
	}

	if isRefTable {
		if err := ({{$.Type}}{}).Table(s); err != nil{
			return err
		}
		s.WriteByte('.')
	} else {
		s.WriteByte(' ')
    }

	s.WriteString("{{$v}}")

	return nil
}
{{end}}


// --- basic sql method ---
var (
	Error{{.Type}}FieldsToStructFieldPointer = errors.New("sql fields can't find match {{stringCaseToLowerCamelCase .Type}} struct field")
)

func ({{.Type}}) FieldsToStructFieldPointer(fs {{if $.IsAddImport}}StitchingSQLGo.{{end}}Fields, {{stringCaseToLowerCamelCase .Type}} *{{.Type}}) ([]interface{}, error) {
	fields := make([]interface{}, 0, len(fs))

	for _, f := range fs {
		var prt interface{}

		switch f { {{range $k,$v:= .FieldMapper}}
		case {{$.Type}}_F.{{$k}}:
			prt = &{{stringCaseToLowerCamelCase $.Type}}.{{$k}}{{end}}
		default:
			return nil, Error{{.Type}}FieldsToStructFieldPointer
		}
		fields = append(fields, prt)
	}
	return fields, nil
}

func ({{.Type}}) QueryRow(db *sql.DB, s {{if .IsAddImport}}StitchingSQLGo.{{end}}Query) ({{.Type}}, error) {
    if db == nil {
        return {{.Type}}{},ErrorDBIsNil
    }

	{{stringCaseToLowerCamelCase .Type}} := {{.Type}}{}
	fields, err := {{stringCaseToLowerCamelCase .Type}}.FieldsToStructFieldPointer(s.Fields(), &{{stringCaseToLowerCamelCase .Type}})
	if err != nil {
		return {{.Type}}{}, err
	}

	if err := QueryRowScan(db,s,fields...); err != nil {
		return {{.Type}}{}, err
	}

	return {{stringCaseToLowerCamelCase .Type}}, nil
}

func ({{.Type}}) Query(db *sql.DB, s {{if .IsAddImport}}StitchingSQLGo.{{end}}Query) ([]{{.Type}}, error) {
    if db == nil {
        return nil,ErrorDBIsNil
    }

	{{stringCaseToLowerCamelCase .Type}}Slice := make([]{{.Type}}, 0)

	sql, args, err := s.Query()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		{{stringCaseToLowerCamelCase .Type}} := {{.Type}}{}
		fields, err := {{stringCaseToLowerCamelCase .Type}}.FieldsToStructFieldPointer(s.Fields(), &{{stringCaseToLowerCamelCase .Type}})
		if err != nil {
			return nil, err
		}

		if err := rows.Scan(fields...); err != nil {
			return nil, err
		}
		{{stringCaseToLowerCamelCase .Type}}Slice = append({{stringCaseToLowerCamelCase .Type}}Slice, {{stringCaseToLowerCamelCase .Type}})
	}

	return {{stringCaseToLowerCamelCase .Type}}Slice, nil
}

func ({{.Type}}) Exec(db *sql.DB, e {{if .IsAddImport}}StitchingSQLGo.{{end}}Exec) (sql.Result, error) {
    if db == nil {
        return nil,ErrorDBIsNil
    }

	sql, args, err := e.Exec()
	if err != nil {
		return nil, err
	}

	return db.Exec(sql, args...)
}

func ({{.Type}}) ExecWithReturning(db *sql.DB, r {{if .IsAddImport}}StitchingSQLGo.{{end}}ExecWithReturning) ({{.Type}}, error) {
    if db == nil {
        return {{.Type}}{},ErrorDBIsNil
    }

	sql, args, err := r.Exec()
	if err != nil {
		return {{.Type}}{}, err
	}

	{{stringCaseToLowerCamelCase .Type}} := {{.Type}}{}
	fields, err := {{stringCaseToLowerCamelCase .Type}}.FieldsToStructFieldPointer({{if .IsAddImport}}StitchingSQLGo.{{end}}Fields(r.ExecWithReturning()), &{{stringCaseToLowerCamelCase .Type}})
	if err != nil {
		return {{.Type}}{}, err
	}

	if err := db.QueryRow(sql, args...).Scan(fields...); err != nil {
		return {{.Type}}{}, err
	}

	return {{stringCaseToLowerCamelCase .Type}}, nil
}

func ({{.Type}}) TxQueryRow(tx *sql.Tx, s {{if .IsAddImport}}StitchingSQLGo.{{end}}Query) ({{.Type}}, error) {
    if tx == nil {
        return {{.Type}}{},ErrorTxIsNil
    }

	{{stringCaseToLowerCamelCase .Type}} := {{.Type}}{}
	fields, err := {{stringCaseToLowerCamelCase .Type}}.FieldsToStructFieldPointer(s.Fields(), &{{stringCaseToLowerCamelCase .Type}})
	if err != nil {
		return {{.Type}}{}, err
	}

	if err := TxQueryRowScan(tx, s,fields...); err != nil {
		return {{.Type}}{}, err
	}

	return {{stringCaseToLowerCamelCase .Type}}, nil
}

func ({{.Type}}) TxQuery(tx *sql.Tx, s {{if .IsAddImport}}StitchingSQLGo.{{end}}Query) ([]{{.Type}}, error) {
    if tx == nil {
        return nil,ErrorTxIsNil
    }

	{{stringCaseToLowerCamelCase .Type}}Slice := make([]{{.Type}}, 0)

	sql, args, err := s.Query()
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(sql, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		{{stringCaseToLowerCamelCase .Type}} := {{.Type}}{}
		fields, err := {{stringCaseToLowerCamelCase .Type}}.FieldsToStructFieldPointer(s.Fields(), &{{stringCaseToLowerCamelCase .Type}})
		if err != nil {
			return nil, err
		}

		if err := rows.Scan(fields...); err != nil {
			return nil, err
		}
	}

	return {{stringCaseToLowerCamelCase .Type}}Slice, nil
}

func ({{.Type}}) TxExec(tx *sql.Tx, e {{if .IsAddImport}}StitchingSQLGo.{{end}}Exec) (sql.Result, error) {
    if tx == nil {
        return nil,ErrorTxIsNil
    }

	sql, args, err := e.Exec()
	if err != nil {
		return nil, err
	}

	return tx.Exec(sql, args...)
}

func ({{.Type}}) TxExecWithReturning(tx *sql.Tx, r {{if .IsAddImport}}StitchingSQLGo.{{end}}ExecWithReturning) ({{.Type}}, error) {
    if tx == nil {
        return {{.Type}}{},ErrorTxIsNil
    }

	sql, args, err := r.Exec()
	if err != nil {
		return {{.Type}}{}, err
	}

	{{stringCaseToLowerCamelCase .Type}} := {{.Type}}{}
	fields, err := {{stringCaseToLowerCamelCase .Type}}.FieldsToStructFieldPointer({{if .IsAddImport}}StitchingSQLGo.{{end}}Fields(r.ExecWithReturning()), &{{stringCaseToLowerCamelCase .Type}})
	if err != nil {
		return {{.Type}}{}, err
	}

	if err := tx.QueryRow(sql, args...).Scan(fields...); err != nil {
		return {{.Type}}{}, err
	}

	return {{stringCaseToLowerCamelCase .Type}}, nil
}
`

func main() {
	// flags validator
	if err := validate.Struct(&flags); err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(GOFILE); os.IsNotExist(err) {
		log.Fatal(err)
	}

	// ast analysis
	fset := token.NewFileSet()
	astF, err := parser.ParseFile(fset, GOFILE, nil, parser.ParseComments)
	if err != nil {
		log.Fatal(err)
	}

	// file template content
	ftc := FileTemplateContent{
		Type:        flags.Type,
		Package:     GOPACKAGE,
		Table:       flags.Table,
		FieldMapper: map[string]string{},
		IsAddImport: flags.IsAddImport,
	}

	if flags.Table == "" {
		ftc.Table = StringCase.ToSnakeCase(flags.Type)
	}

	isFindType := false
	ast.Inspect(astF, func(n ast.Node) bool {
		t, ok := n.(*ast.TypeSpec)
		if ok == false {
			return true
		}

		s, ok := t.Type.(*ast.StructType)
		if ok == false {
			return true
		}

		if t.Name.Name != flags.Type {
			return false
		}
		isFindType = true
		for _, f := range s.Fields.List {
			n := f.Names[0].Name
			ftc.FieldMapper[n] = StringCase.ToSnakeCase(n)
		}

		return false
	})

	if !isFindType {
		log.Fatal("-type error, not find type as struct")
	}

	// file content container
	b := bytes.NewBuffer([]byte{})

	// new file template
	t, err := template.New("").Funcs(template.FuncMap{
		"stringCaseToLowerCamelCase": StringCase.ToLowerCamelCase,
	}).Parse(FileTemplate)
	if err != nil {
		log.Fatal(err)
	}
	// generate file content write to file content container
	if err := t.Execute(b, ftc); err != nil {
		log.Fatal(err)
	}

	// file name
	fileName := fmt.Sprintf("%s_stitching_sql.go", flags.Type)
	if flags.FileName != "" {
		fileName = flags.FileName
		if fileName[len(fileName)-3:] != ".go" {
			fileName += ".go"
		}
	}

	content, err := format.Source(b.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(fileName, content, 0644); err != nil {
		log.Fatal(err)
	}

	// file content container
	c := TemplateStitchingSQLContent{
		Package:     GOPACKAGE,
		IsAddImport: flags.IsAddImport,
	}.generateContent()

	cFormatted, err := format.Source(c)
	if err != nil {
		log.Fatal(err)
	}

	// write template
	if err := ioutil.WriteFile("reference_stitching_sql.go", cFormatted, 0644); err != nil {
		log.Fatal(err)
	}
}
