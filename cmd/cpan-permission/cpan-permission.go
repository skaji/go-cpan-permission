package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/skaji/go-cpan-permission"
	"github.com/skaji/go-table"
)

func main() {
	os.Exit(run(os.Args))
}

func run(args []string) int {
	if len(args) < 2 || args[1] == "-h" || args[1] == "--help" {
		fmt.Println("Usage: cpan-permission MODULE")
		return 1
	}

	module := args[1]
	p := permission.New()
	distfile, result, err := p.Get(module)
	if err != nil {
		fmt.Println(err)
		return 1
	}

	t := table.New()
	for _, r := range result {
		owner := r.Owner
		if owner == "" {
			owner = "N/A"
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
