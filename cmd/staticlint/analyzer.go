package staticlint

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var exitAnalyzer = &analysis.Analyzer{
	Name: "os exit analyzer",
	Doc:  "has to be no os exit in main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.File:
				if x.Name.Name != "main" {
					return false
				}
			case *ast.SelectorExpr:
				if x.Sel.Name == "Exit" {
					pass.Reportf(x.Pos(), "there is os exit in main")

				}
			}
			return true
		})
	}
	return nil, nil
}
