package staticlint

import (
	"github.com/gostaticanalysis/nakedreturn"
	"github.com/gostaticanalysis/nilerr"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
	"strings"
)

func main() {

	var analyzers []*analysis.Analyzer

	analyzers = append(analyzers, exitAnalyzer)
	analyzers = append(analyzers, printf.Analyzer)
	analyzers = append(analyzers, shadow.Analyzer)
	analyzers = append(analyzers, structtag.Analyzer)
	analyzers = append(analyzers, structtag.Analyzer)

	analyzers = append(analyzers, nakedreturn.Analyzer)
	analyzers = append(analyzers, nilerr.Analyzer)

	for _, v := range staticcheck.Analyzers {
		if strings.Contains(v.Name, "SA") {
			analyzers = append(analyzers, v)
		}
	}

	for _, v := range stylecheck.Analyzers {
		if v.Name == "ST1000" || v.Name == "ST1001" {
			analyzers = append(analyzers, v)
		}
	}

	multichecker.Main(
		analyzers...,
	)
}
