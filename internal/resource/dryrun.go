package resource

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"

	"github.com/slok/kahoy/internal/model"
)

// pfunc is a helper alias to be less vebose on func declarations.
type pfunc = func(format string, a ...interface{}) string

type dryRunManager struct {
	out io.Writer

	redSprintf       pfunc
	yellowSprintf    pfunc
	whiteBoldSprintf pfunc
	cyanSprintf      pfunc
	greenSprintf     pfunc
}

// NewDryRunManager returns a resource manager that dry runs the changes
// without the need of an apiserver.
func NewDryRunManager(disableColor bool, out io.Writer) Manager {
	if out == nil {
		out = os.Stdout
	}

	redSprintf := fmt.Sprintf
	yellowSprintf := fmt.Sprintf
	whiteBoldSprintf := fmt.Sprintf
	cyanSprintf := fmt.Sprintf
	greenSprintf := fmt.Sprintf
	if !disableColor {
		color.NoColor = false // This is required because Color infers and uses globals, in our case we manage with explicit flag and force this.
		redSprintf = color.New(color.FgRed).Sprintf
		yellowSprintf = color.New(color.FgYellow).Sprintf
		whiteBoldSprintf = color.New(color.FgWhite, color.Bold).Sprintf
		cyanSprintf = color.New(color.FgCyan).Sprintf
		greenSprintf = color.New(color.FgGreen).Sprintf
	}

	return dryRunManager{
		out: out,

		redSprintf:       redSprintf,
		yellowSprintf:    yellowSprintf,
		whiteBoldSprintf: whiteBoldSprintf,
		cyanSprintf:      cyanSprintf,
		greenSprintf:     greenSprintf,
	}
}

func (d dryRunManager) Apply(ctx context.Context, resources []model.Resource) error {
	d.printTree("Apply", resources, d.greenSprintf)
	return nil
}

func (d dryRunManager) Delete(ctx context.Context, resources []model.Resource) error {
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

	c := 0
	d.printf("\n⯈ %s\n", d.whiteBoldSprintf(title))
	for groupID, ress := range resByGroup {
		// Print groups.
		joinSymbol := `├── `
		groupSymbol := `│`
		if c+1 >= len(resByGroup) {
			joinSymbol = `└── `
			groupSymbol = ` `
		}
		d.printf("%s⯈ %s\n", joinSymbol, d.yellowSprintf(groupID))

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
}

func (d dryRunManager) printf(format string, a ...interface{}) {
	fmt.Fprintf(d.out, format, a...)
}
