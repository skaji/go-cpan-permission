package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/skaji/go-cpan-permission"
	"github.com/skaji/go-table"
)

var (
	asjson = flag.Bool("j", false, "json output")
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: cpan-permission [-j] Module")
	}
	flag.Parse()
	module := flag.Arg(0)
	os.Exit(run(module))
}

type jsonRes struct {
	Distfile   string                        `json:"distfile"`
	Permission []permission.PermissionResult `json:"permission"`
}

func run(module string) int {
	if module == "" {
		flag.Usage()
		return 1
	}
	p := permission.New()
	distfile, result, err := p.Get(module)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	if *asjson {
		b, err := json.MarshalIndent(jsonRes{Distfile: distfile, Permission: result}, "", "  ")
		if err != nil {
			fmt.Println(err)
			return 1
		}
		fmt.Println(string(b))
		return 0
	}

	t := table.New()
	t.Add([]string{"module_name", "owner", "co_maintainers"})
	for _, r := range result {
		owner := "N/A"
		if owner != "" {
			owner = r.Owner
		}
		co := "N/A"
		if len(r.CoMaintainers) > 0 {
			co = strings.Join(r.CoMaintainers, ",")
		}
		t.Add([]string{r.ModuleName, owner, co})
	}
	fmt.Print(distfile, "\n\n")
	t.Render(os.Stdout)
	return 0
}
