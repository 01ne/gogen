package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/01ne/gogen/automock"
	"github.com/01ne/gogen/importer"
	"github.com/01ne/gogen/strconv"
)

var (
	out      = flag.String("o", "", "override the name of the generated code. Default value is by generated based on the name of the interface, e.g.: Reader -> reader_mock_test.go (use \"-\" to print to stdout)")
	mockName = flag.String("mock-name", "", "override the name for the mock struct")
	mockPkg  = flag.String("mock-pkg", "", "override the package name for the mock")
	pkg      = flag.String("pkg", ".", "override package to get the interface from. It can be specified in the interface name, e.g.: goautomock io.Reader")
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	iface := flag.Arg(0)

	if iface == "" {
		log.Fatal("need to specify an interface name")
	}

	parts := strings.Split(iface, ".")
	switch len(parts) {
	case 1:
	case 2:
		if *pkg != "." {
			log.Fatalf("unexpected -pkg value (%q), package is already defined in the interface name as %s", *pkg, parts[0])
		}
		*pkg = parts[0]
		iface = parts[1]
	default:
		log.Fatalf("invalid interface %q", iface)
	}

	gen, err := automock.NewGenerator(*pkg, iface)
	if err != nil {
		log.Fatal(err)
	}

	gen.SetTemplate(template)

	if *mockName != "" {
		gen.SetName(*mockName)
	}
	inPkg := *pkg == "." && path.Dir(*out) == "."
	gen.SetInternal(inPkg)
	if *mockPkg == "" && !inPkg {
		p, err := importer.Default().Import(".")
		if err != nil {
			log.Fatal(err)
		}
		*mockPkg = p.Name()
	}
	if *mockPkg != "" {
		gen.SetPackage(*mockPkg)
	}

	w := os.Stdout
	if *out == "" {
		*out = fmt.Sprintf("%s_test.go", gen.Name())
		if p := regexp.MustCompile(".*/").ReplaceAllString(*pkg, ""); !inPkg && p != "" && p != "." {
			*out = p + "_" + *out
		}
	}
	if *out != "-" {
		*out = strconv.SnakeCase(*out)
		log.Printf("Generating mock for %s in %s", iface, *out)
		w, err = os.OpenFile(*out, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}

	err = gen.Write(w)
	switch err := err.(type) {
	case automock.GenerationError:
		log.Println(err.CodeWithLineNumbers())
		log.Fatal(err)
	case error:
		log.Fatal(err)
	}
}

var template = `/*
* CODE GENERATED AUTOMATICALLY WITH github.com/01ne/gogen/cmd/gosimplemock
* THIS FILE SHOULD NOT BE EDITED BY HAND
*/

package {{.Package}}

import (
	"fmt"
{{range $path, $name := .Imports}}
	{{$name}} "{{$path}}"{{end}}
)

// {{.Name}} mock
type {{.Name}} struct {
	{{range .Methods}}
	{{.Name}}Func func({{range $index, $type := .ParamTypes}}{{if $index}}, {{end}}{{$type}}{{end}}) ({{range $index, $type := .ReturnTypes}}{{if $index}}, {{end}}{{$type}}{{end}})
	{{end}}
}

{{$gen := .}}
{{range .Methods}}
// {{.Name}} mocked method
func (m *{{$gen.Name}}) {{.Name}}({{range $index, $type := .ParamTypes}}{{if $index}}, {{end}}p{{$index}} {{$type}}{{end}}) ({{range $index, $type := .ReturnTypes}}{{if $index}}, {{end}}{{$type}}{{end}}) {
	if m.{{.Name}}Func == nil {
		panic("unexpected call to mocked method {{.Name}}")
	}
	{{if .ReturnTypes}}return {{end}} m.{{.Name}}Func({{range $index, $type := .ParamTypes}}{{if $index}}, {{end}}p{{$index}}{{end}})
}
{{end}}
`
