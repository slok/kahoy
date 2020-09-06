package dryrun

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/fatih/color"

	"github.com/slok/kahoy/internal/model"
	"github.com/slok/kahoy/internal/resource/manage"
)

// pfunc is a helper alias to be less vebose on func declarations.
type pfunc = func(format string, a ...interface{}) string

type dryRunManager struct {
	out io.Writer

	redSprintf        pfunc
	yellowBoldSprintf pfunc
	whiteBoldSprintf  pfunc
	cyanSprintf       pfunc
	greenSprintf      pfunc
	blueSprintf       pfunc
}

// NewManager returns a resource manager that dry runs the changes
// without the need of an apiserver.
func NewManager(disableColor bool, out io.Writer) manage.ResourceManager {
	if out == nil {
		out = os.Stdout
	}

	redSprintf := fmt.Sprintf
	yellowBoldSprintf := fmt.Sprintf
	whiteBoldSprintf := fmt.Sprintf
	cyanSprintf := fmt.Sprintf
	greenSprintf := fmt.Sprintf
	blueSprintf := fmt.Sprintf
	if !disableColor {
		color.NoColor = false // This is required because Color infers and uses globals, in our case we manage with explicit flag and force this.
		redSprintf = color.New(color.FgRed).Sprintf
		yellowBoldSprintf = color.New(color.FgYellow, color.Bold).Sprintf
		whiteBoldSprintf = color.New(color.FgWhite, color.Bold).Sprintf
		cyanSprintf = color.New(color.FgCyan).Sprintf
		greenSprintf = color.New(color.FgGreen).Sprintf
		blueSprintf = color.New(color.FgBlue).Sprintf
	}

	return dryRunManager{
		out: out,

		redSprintf:        redSprintf,
		yellowBoldSprintf: yellowBoldSprintf,
		whiteBoldSprintf:  whiteBoldSprintf,
		cyanSprintf:       cyanSprintf,
		greenSprintf:      greenSprintf,
		blueSprintf:       blueSprintf,
	}
}

func (d dryRunManager) Apply(ctx context.Context, resources []model.Resource) error {
	d.sort(resources)
	d.printTree("Apply", resources, d.greenSprintf)
	return nil
}

func (d dryRunManager) Delete(ctx context.Context, resources []model.Resource) error {
	d.sort(resources)
	d.printTree("Delete", resources, d.redSprintf)
	return nil
}

func (d dryRunManager) printTree(title string, resources []model.Resource, printColor pfunc) {
	if len(resources) == 0 {
		return
	}

	// Group by groups.
	resByGroup := map[string][]model.Resource{}
	for _, res := range resources {
		resByGroup[res.GroupID] = append(resByGroup[res.GroupID], res)
	}
	// Sort groups so we print in order.
	orderedGroups := make([]string, 0, len(resByGroup))
	for groupID := range resByGroup {
		orderedGroups = append(orderedGroups, groupID)
	}
	sort.Slice(orderedGroups, func(i, j int) bool { return orderedGroups[i] < orderedGroups[j] })

	c := 0
	d.printf("\n⯈ %s %s\n", d.whiteBoldSprintf(title), d.blueSprintf("(%d resources)", len(resources)))
	for _, groupID := range orderedGroups {
		ress := resByGroup[groupID] // We got the these group IDs from this map, should exist.

		// Print groups.
		joinSymbol := `├── `
		groupSymbol := `│`
		if c+1 >= len(resByGroup) {
			joinSymbol = `└── `
			groupSymbol = ` `
		}
		d.printf("%s⯈ %s %s\n", joinSymbol, d.yellowBoldSprintf(groupID), d.blueSprintf("(%d resources)", len(ress)))

		// Print resources.
		for i, res := range ress {
			joinSymbol := groupSymbol + `   ├── `
			if i+1 >= len(ress) {
				joinSymbol = groupSymbol + `   └── `
			}

			d.printf(joinSymbol + printColor(res.ID) + d.cyanSprintf(" (%s)", res.ManifestPath) + "\n")
		}

		c++
	}
	fmt.Println()
}

func (d dryRunManager) printf(format string, a ...interface{}) {
	fmt.Fprintf(d.out, format, a...)
}

func (d dryRunManager) sort(rs []model.Resource) {
	sort.SliceStable(rs, func(i, j int) bool {
		ri, rj := rs[i], rs[j]
		return ri.GroupID+ri.ID < rj.GroupID+rj.ID
	})
}
