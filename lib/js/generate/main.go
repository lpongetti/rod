// Package main ...
package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-rod/rod/lib/utils"
	"github.com/ysmood/gson"
)

func main() {
	list := getList()
	out := "// Package js generated by \"lib/js/generate\"\npackage js\n\n"

	for _, fn := range list.Arr() {
		name := fn.Get("name").Str()
		def := fn.Get("definition").Str()
		out += utils.S(`
			// {{.Name}} ...
			var {{.Name}} = &Function{
				Name: "{{.name}}",
				Definition:   {{.definition}},
				Dependencies: {{.dependencies}},
			}
		`,
			"Name", fnName(name),
			"name", name,
			"definition", utils.EscapeGoString(def),
			"dependencies", getDeps(def),
		)
	}

	utils.E(utils.OutputFile("lib/js/helper.go", out))

	utils.Exec("gofumpt -w lib/js/helper.go")
}

var regDeps = regexp.MustCompile(`\Wfunctions.(\w+)`)

func getDeps(fn string) string {
	ms := regDeps.FindAllStringSubmatch(fn, -1)

	list := []string{}

	for _, m := range ms {
		list = append(list, fnName(m[1]))
	}

	return "[]*Function{" + strings.Join(list, ",") + "}"
}

func fnName(name string) string {
	return strings.ToUpper(name[0:1]) + name[1:]
}

func getList() gson.JSON {
	code := utils.ExecLine(false, "npx -ys -- uglify-js@3.17.4 -c -m -- lib/js/helper.js")

	script := fmt.Sprintf(`
		%s

		const list = []

		for (const name in functions) {
			const reg = new RegExp('^(async )?' + name)
			const definition = functions[name].toString().replace(reg, '$1function')
			list.push({name, definition})
		}

		console.log(JSON.stringify(list))
	`, code)

	tmp := "tmp/helper.js"

	utils.E(utils.OutputFile(tmp, script))

	return gson.NewFrom(utils.ExecLine(false, "node", tmp))
}
