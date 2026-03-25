// Command pizza is the dispatch CLI for the pizza workflow system.
// It runs from anywhere — including a phone terminal — and gives you
// the next slice to work on, tracks progress, and learns from completions.
//
// Usage:
//   pizza plan          — full dispatch plan with all agents
//   pizza next          — what to do right now
//   pizza express       — 48-hour express order only
//   pizza start <id>    — mark a slice as in-progress
//   pizza done <id>     — mark a slice as completed
//   pizza oven <id>     — mark a slice as submitted/waiting
//   pizza burnt <id>    — mark a slice as failed
//   pizza status        — grid overview
//   pizza holo          — holographic memory visualization
//   pizza log           — full build log
//   pizza legend        — what each topping means
//   pizza repl          — interactive dispatch shell
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/qleviathan/fastlands/dispatch"
	"github.com/qleviathan/fastlands/holo"
	"github.com/qleviathan/fastlands/lattice"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// Initialize the system
	grid := lattice.NewGrid()
	lattice.LoadFullMenu(grid)
	disp := dispatch.NewDispatcher(grid)
	mem := holo.NewStore()

	cmd := os.Args[1]
	switch cmd {
	case "plan":
		runPlan(disp)
	case "next":
		runNext(disp)
	case "express":
		runExpress()
	case "start":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pizza start <slice-id>")
			return
		}
		disp.MarkPrepping(os.Args[2])
		fmt.Printf("Started: %s\n", os.Args[2])
		showNext(disp, grid.Get(os.Args[2]))
	case "done":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pizza done <slice-id>")
			return
		}
		s := grid.Get(os.Args[2])
		disp.MarkDone(os.Args[2])
		fmt.Printf("Completed: %s\n", os.Args[2])
		if s != nil {
			mem.LearnFromCompletion(s.ID, int(s.Dough), int(s.Sauce), "success", 0, "")
		}
		showNext(disp, s)
	case "oven":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pizza oven <slice-id>")
			return
		}
		disp.MarkInOven(os.Args[2])
		fmt.Printf("In oven (waiting): %s\n", os.Args[2])
	case "burnt":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pizza burnt <slice-id>")
			return
		}
		reason := ""
		if len(os.Args) > 3 {
			reason = strings.Join(os.Args[3:], " ")
		}
		disp.MarkBurnt(os.Args[2], reason)
		fmt.Printf("Burnt: %s (%s)\n", os.Args[2], reason)
	case "status":
		runStatus(grid)
	case "holo":
		runHolo(mem)
	case "log":
		fmt.Print(disp.Logger.FormatLog())
		fmt.Print(disp.Logger.Summary())
	case "legend":
		printLegend()
	case "repl":
		runREPL(grid, disp, mem)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("pizza — dispatch system")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  plan          full dispatch plan")
	fmt.Println("  next          what to do right now")
	fmt.Println("  express       48-hour express order")
	fmt.Println("  start <id>    mark slice in-progress")
	fmt.Println("  done <id>     mark slice completed")
	fmt.Println("  oven <id>     mark slice submitted")
	fmt.Println("  burnt <id>    mark slice failed")
	fmt.Println("  status        grid overview")
	fmt.Println("  holo          holographic memory")
	fmt.Println("  log           build log")
	fmt.Println("  legend        topping decoder")
	fmt.Println("  repl          interactive shell")
}

func runPlan(d *dispatch.Dispatcher) {
	plan := d.GeneratePlan()
	fmt.Print(dispatch.FormatPlan(plan))
}

func runNext(d *dispatch.Dispatcher) {
	next := d.Next(nil)
	if next == nil {
		fmt.Println("All slices are done or burnt. Kitchen is closed.")
		return
	}
	fmt.Printf("NEXT UP: %s\n", next.Name)
	fmt.Printf("  ID: %s\n", next.ID)
	fmt.Printf("  Position: %s × %s (temp %d)\n", next.Dough, next.Sauce, next.Temperature)
	fmt.Printf("  Recipe: %s\n", next.Notes)
}

func runExpress() {
	grid := lattice.NewGrid()
	lattice.LoadExpressOrder(grid)
	d := dispatch.NewDispatcher(grid)
	plan := d.GeneratePlan()
	fmt.Print(dispatch.FormatPlan(plan))
}

func runStatus(g *lattice.Grid) {
	stats := g.Stats()
	fmt.Println(stats)
	fmt.Println()

	all := g.AllSlices()
	for _, s := range all {
		status := s.Status.String()
		fmt.Printf("  [%-8s] %-30s %s×%s  temp=%d  flavor=%d/%d\n",
			status, s.Name, s.Dough, s.Sauce, s.Temperature, s.Flavor.Num, s.Flavor.Den)
	}
}

func runHolo(mem *holo.Store) {
	stats := mem.Stats()
	fmt.Println(stats)
	if stats.TotalPatterns == 0 {
		fmt.Println("No memories yet. Complete some slices to build holographic memory.")
		return
	}

	// Show interference pattern from center of grid
	w := mem.Project(2, 2, 4)
	fmt.Print(holo.FormatWave(w, 5, 9))
}

func showNext(d *dispatch.Dispatcher, justDone *lattice.Slice) {
	next := d.Next(justDone)
	if next != nil {
		fmt.Printf("\nNEXT → %s (%s)\n", next.Name, next.ID)
		fmt.Printf("  Recipe: %s\n", next.Notes)
	}
}

func printLegend() {
	fmt.Println("══════════════════════════════════════")
	fmt.Println("       PIZZA DECODER RING             ")
	fmt.Println("══════════════════════════════════════")
	fmt.Println()
	fmt.Println("DOUGH (n-axis — preparation style):")
	fmt.Println("  Thin Crust    (0) — lowest prep, fastest out")
	fmt.Println("  Hand Tossed   (1) — moderate craft")
	fmt.Println("  Deep Dish     (2) — analytical depth")
	fmt.Println("  Stuffed Crust (3) — engineering heavy")
	fmt.Println("  Neapolitan    (4) — artisan, expert-level")
	fmt.Println()
	fmt.Println("SAUCE (m-axis — kitchen station):")
	fmt.Println("  Marinara      (0) — fastest oven, pays every 3 days")
	fmt.Println("  Alfredo       (1) — weekly pay, prefers experienced chefs")
	fmt.Println("  Pesto         (2) — highest volume, 6K+ open orders")
	fmt.Println("  BBQ           (3) — create your own listings, 20% house cut")
	fmt.Println("  Buffalo       (4) — content kitchen, $5/mo entry, scales")
	fmt.Println("  Garlic Butter (5) — fast turnaround, 8 free bids")
	fmt.Println("  Vodka         (6) — premium phone consultations $150-500/hr")
	fmt.Println("  Ranch         (7) — competitions, $100K prize pools")
	fmt.Println("  Hot Honey     (8) — newsletter, free to 2.5K subs")
	fmt.Println("  Truffle       (9) — cognitive work, $25-150/hr, weekly via Deel")
	fmt.Println()
	fmt.Println("TOPPINGS (task character):")
	fmt.Println("  Pepperoni     — high urgency, do it now")
	fmt.Println("  Mushroom      — requires deep focus")
	fmt.Println("  Olive         — passive / async, can run in background")
	fmt.Println("  Jalapeño      — time-sensitive window")
	fmt.Println("  Basil         — creative component")
	fmt.Println()
	fmt.Println("TEMPERATURE (d = n + m):")
	fmt.Println("  0-2  — express window, out in hours")
	fmt.Println("  3-5  — standard oven, 3-14 days")
	fmt.Println("  6+   — artisan specials, high reward")
	fmt.Println()
	fmt.Println("FLAVOR = F(n) · F(m) — basis coupling strength")
	fmt.Println("  Higher flavor = more compound value")
	fmt.Println()
	fmt.Println("COUPLING RULES:")
	fmt.Println("  Same dough, different sauce → skill transfers between stations")
	fmt.Println("  Same sauce, different dough → platform experience compounds")
	fmt.Println("  Adjacent gap (1) → CASCADE: output feeds directly into next")
	fmt.Println("  Stride-2 gap   → RESONANCE: harmonic reinforcement")
	fmt.Println("  Same diagonal   → PARALLEL: same total effort, independent")
}

func runREPL(g *lattice.Grid, d *dispatch.Dispatcher, mem *holo.Store) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("pizza> interactive dispatch shell")
	fmt.Println("  Type 'help' for commands, 'quit' to exit")
	fmt.Println()

	for {
		fmt.Print("pizza> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]

		switch cmd {
		case "help":
			fmt.Println("  plan       — full dispatch plan")
			fmt.Println("  next       — what to do now")
			fmt.Println("  status     — grid overview")
			fmt.Println("  start <id> — begin a slice")
			fmt.Println("  done <id>  — complete a slice")
			fmt.Println("  oven <id>  — mark submitted")
			fmt.Println("  burnt <id> — mark failed")
			fmt.Println("  holo       — holographic memory")
			fmt.Println("  wave <n> <m> — project interference from position")
			fmt.Println("  log        — build log")
			fmt.Println("  legend     — decoder ring")
			fmt.Println("  quit       — exit")

		case "plan":
			plan := d.GeneratePlan()
			fmt.Print(dispatch.FormatPlan(plan))

		case "next":
			runNext(d)

		case "status":
			runStatus(g)

		case "start":
			if len(parts) < 2 {
				fmt.Println("Usage: start <id>")
				continue
			}
			d.MarkPrepping(parts[1])
			fmt.Printf("Started: %s\n", parts[1])

		case "done":
			if len(parts) < 2 {
				fmt.Println("Usage: done <id>")
				continue
			}
			s := g.Get(parts[1])
			d.MarkDone(parts[1])
			fmt.Printf("Done: %s\n", parts[1])
			if s != nil {
				mem.LearnFromCompletion(s.ID, int(s.Dough), int(s.Sauce), "success", 0, "")
				suggestion := mem.SuggestNext(int(s.Dough), int(s.Sauce))
				fmt.Printf("Holo: %s\n", suggestion)
			}
			showNext(d, s)

		case "oven":
			if len(parts) < 2 {
				fmt.Println("Usage: oven <id>")
				continue
			}
			d.MarkInOven(parts[1])
			fmt.Printf("In oven: %s\n", parts[1])

		case "burnt":
			if len(parts) < 2 {
				fmt.Println("Usage: burnt <id>")
				continue
			}
			reason := ""
			if len(parts) > 2 {
				reason = strings.Join(parts[2:], " ")
			}
			d.MarkBurnt(parts[1], reason)
			fmt.Printf("Burnt: %s\n", parts[1])

		case "holo":
			runHolo(mem)

		case "wave":
			if len(parts) < 3 {
				fmt.Println("Usage: wave <n> <m>")
				continue
			}
			n, m := 0, 0
			fmt.Sscanf(parts[1], "%d", &n)
			fmt.Sscanf(parts[2], "%d", &m)
			w := mem.Project(n, m, 4)
			fmt.Print(holo.FormatWave(w, 5, 9))

		case "log":
			fmt.Print(d.Logger.FormatLog())
			fmt.Print(d.Logger.Summary())

		case "legend":
			printLegend()

		case "quit", "exit":
			fmt.Println("Kitchen closed.")
			return

		default:
			fmt.Printf("Unknown: %s (type 'help')\n", cmd)
		}
	}
}
