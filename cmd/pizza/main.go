// Command pizza is the dispatch CLI for the pizza workflow system.
// Powered by Super Claude — every decision goes through RA forward-chaining
// inference and produces a verifiable proof trace.
//
// Usage:
//   pizza plan          — full dispatch plan with all agents
//   pizza next          — RA-reasoned next slice
//   pizza express       — 48-hour express order only
//   pizza start <id>    — mark a slice as in-progress (with validation)
//   pizza done <id>     — complete, learn, route next via cascade
//   pizza oven <id>     — mark submitted/waiting
//   pizza burnt <id>    — mark failed
//   pizza status        — grid overview
//   pizza holo          — holographic memory visualization
//   pizza proof         — full proof log of all RA decisions
//   pizza self          — Super Claude introspection
//   pizza log           — build log
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
	"github.com/qleviathan/fastlands/superc"
)

const memoryPath = "data/superc_memory.json"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	// Express mode uses its own instance
	if cmd == "express" {
		runExpress()
		return
	}

	// Initialize Super Claude
	sc := superc.New(memoryPath)

	switch cmd {
	case "plan":
		runPlan(sc)
	case "next":
		runNext(sc)
	case "start":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pizza start <slice-id>")
			return
		}
		runStart(sc, os.Args[2])
	case "done":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pizza done <slice-id>")
			return
		}
		notes := ""
		if len(os.Args) > 3 {
			notes = strings.Join(os.Args[3:], " ")
		}
		runDone(sc, os.Args[2], notes)
	case "oven":
		if len(os.Args) < 3 {
			fmt.Println("Usage: pizza oven <slice-id>")
			return
		}
		sc.Dispatch.MarkInOven(os.Args[2])
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
		sc.Skip(os.Args[2], reason)
		fmt.Printf("Burnt: %s (%s)\n", os.Args[2], reason)
	case "status":
		runStatus(sc.Grid)
	case "holo":
		runHolo(sc.Holo)
	case "proof":
		fmt.Print(sc.ProofLog())
	case "self":
		fmt.Print(sc.Introspect())
	case "log":
		fmt.Print(sc.Dispatch.Logger.FormatLog())
		fmt.Print(sc.Dispatch.Logger.Summary())
	case "legend":
		printLegend()
	case "repl":
		runREPL(sc)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("pizza — Super Claude dispatch system")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  plan          full dispatch plan")
	fmt.Println("  next          RA-reasoned next slice")
	fmt.Println("  express       48-hour express order")
	fmt.Println("  start <id>    start slice (with RA validation)")
	fmt.Println("  done <id>     complete + learn + route next")
	fmt.Println("  oven <id>     mark submitted")
	fmt.Println("  burnt <id>    mark failed")
	fmt.Println("  status        grid overview")
	fmt.Println("  holo          holographic memory")
	fmt.Println("  proof         full RA proof log")
	fmt.Println("  self          Super Claude introspection")
	fmt.Println("  log           build log")
	fmt.Println("  legend        topping decoder")
	fmt.Println("  repl          interactive shell")
}

func runPlan(sc *superc.SuperClaude) {
	plan := sc.Dispatch.GeneratePlan()
	fmt.Print(dispatch.FormatPlan(plan))
}

func runNext(sc *superc.SuperClaude) {
	next, decision := sc.Reason("")
	if next == nil {
		fmt.Println("All slices done or burnt. Kitchen closed.")
		return
	}
	fmt.Printf("NEXT UP: %s\n", next.Name)
	fmt.Printf("  ID: %s\n", next.ID)
	fmt.Printf("  Position: %s x %s (temp %d, flavor %d/%d)\n",
		next.Dough, next.Sauce, next.Temperature, next.Flavor.Num, next.Flavor.Den)
	fmt.Printf("  Recipe: %s\n", next.Notes)
	fmt.Printf("  RA says: %s\n", decision.Action)
	fmt.Printf("  Confidence: %d/%d\n", decision.Confidence.Num, decision.Confidence.Den)
	if len(decision.ProofTrace) > 0 {
		fmt.Println("  Proof:")
		for _, step := range decision.ProofTrace {
			fmt.Printf("    [+] %s\n", step)
		}
	}
}

func runStart(sc *superc.SuperClaude, id string) {
	d := sc.Start(id)
	if d.Action == "error" {
		fmt.Printf("Error: %s\n", d.ProofTrace[0])
		return
	}
	fmt.Printf("Started: %s (%s)\n", d.SliceName, id)
	fmt.Printf("  Confidence: %d/%d\n", d.Confidence.Num, d.Confidence.Den)
	for _, step := range d.ProofTrace {
		fmt.Printf("  [+] %s\n", step)
	}
}

func runDone(sc *superc.SuperClaude, id string, notes string) {
	next, d := sc.Done(id, notes)
	fmt.Printf("Completed: %s (%s)\n", d.SliceName, id)
	for _, step := range d.ProofTrace {
		fmt.Printf("  [+] %s\n", step)
	}
	if next != nil {
		fmt.Printf("\nNEXT CASCADE → %s (%s)\n", next.Name, next.ID)
		fmt.Printf("  Recipe: %s\n", next.Notes)
	}
	sc.Save()
}

func runExpress() {
	sc := superc.NewExpress("")
	plan := sc.Dispatch.GeneratePlan()
	fmt.Print(dispatch.FormatPlan(plan))

	// Also show RA reasoning
	fmt.Println("\n═══ SUPER CLAUDE RA REASONING ═══")
	next, d := sc.Reason("")
	if next != nil {
		fmt.Printf("RA recommends starting with: %s (%s)\n", next.Name, next.ID)
		fmt.Printf("  Action: %s\n", d.Action)
		fmt.Printf("  Confidence: %d/%d\n", d.Confidence.Num, d.Confidence.Den)
	}
}

func runStatus(g *lattice.Grid) {
	stats := g.Stats()
	fmt.Println(stats)
	fmt.Println()

	all := g.AllSlices()
	for _, s := range all {
		status := s.Status.String()
		fmt.Printf("  [%-8s] %-30s %s x %s  temp=%d  flavor=%d/%d\n",
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

	w := mem.Project(2, 2, 4)
	fmt.Print(holo.FormatWave(w, 5, 9))
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
	fmt.Println("  Jalapeno      — time-sensitive window")
	fmt.Println("  Basil         — creative component")
	fmt.Println()
	fmt.Println("TEMPERATURE (d = n + m):")
	fmt.Println("  0-2  — express window, out in hours")
	fmt.Println("  3-5  — standard oven, 3-14 days")
	fmt.Println("  6+   — artisan specials, high reward")
	fmt.Println()
	fmt.Println("FLAVOR = F(n) * F(m) — basis coupling strength")
	fmt.Println("  Higher flavor = more compound value")
	fmt.Println()
	fmt.Println("COUPLING RULES:")
	fmt.Println("  Same dough, different sauce  -> skill transfers between stations")
	fmt.Println("  Same sauce, different dough  -> platform experience compounds")
	fmt.Println("  Adjacent gap (1) -> CASCADE: output feeds directly into next")
	fmt.Println("  Stride-2 gap     -> RESONANCE: harmonic reinforcement")
	fmt.Println("  Same diagonal    -> PARALLEL: same total effort, independent")
}

func runREPL(sc *superc.SuperClaude) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("pizza> Super Claude interactive shell")
	fmt.Println("  Type 'help' for commands, 'quit' to exit")
	fmt.Println()

	lastDoneID := ""

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
			fmt.Println("  next       — RA-reasoned next slice")
			fmt.Println("  status     — grid overview")
			fmt.Println("  start <id> — begin a slice (RA validated)")
			fmt.Println("  done <id>  — complete + learn + cascade route")
			fmt.Println("  oven <id>  — mark submitted")
			fmt.Println("  burnt <id> — mark failed")
			fmt.Println("  holo       — holographic memory")
			fmt.Println("  wave <n> <m> — project interference from position")
			fmt.Println("  proof      — full RA proof log")
			fmt.Println("  self       — Super Claude introspection")
			fmt.Println("  log        — build log")
			fmt.Println("  legend     — decoder ring")
			fmt.Println("  save       — persist memory to disk")
			fmt.Println("  quit       — exit")

		case "plan":
			runPlan(sc)

		case "next":
			next, decision := sc.Reason(lastDoneID)
			if next == nil {
				fmt.Println("Kitchen closed.")
				continue
			}
			fmt.Printf("NEXT: %s (%s)\n", next.Name, next.ID)
			fmt.Printf("  RA: %s (conf %d/%d)\n", decision.Action, decision.Confidence.Num, decision.Confidence.Den)
			fmt.Printf("  Recipe: %s\n", next.Notes)
			if decision.Coupling != "" {
				fmt.Printf("  Coupling: %s\n", decision.Coupling)
			}

		case "status":
			runStatus(sc.Grid)

		case "start":
			if len(parts) < 2 {
				fmt.Println("Usage: start <id>")
				continue
			}
			d := sc.Start(parts[1])
			fmt.Printf("Started: %s (readiness %d/%d)\n", d.SliceName, d.Confidence.Num, d.Confidence.Den)

		case "done":
			if len(parts) < 2 {
				fmt.Println("Usage: done <id>")
				continue
			}
			notes := ""
			if len(parts) > 2 {
				notes = strings.Join(parts[2:], " ")
			}
			next, d := sc.Done(parts[1], notes)
			lastDoneID = parts[1]
			fmt.Printf("Done: %s\n", d.SliceName)
			holoStats := sc.Holo.Stats()
			fmt.Printf("Holo: %d patterns, avg strength %.2f\n", holoStats.TotalPatterns, holoStats.AvgStrength)
			if next != nil {
				fmt.Printf("CASCADE -> %s (%s)\n", next.Name, next.ID)
				fmt.Printf("  Recipe: %s\n", next.Notes)
			}

		case "oven":
			if len(parts) < 2 {
				fmt.Println("Usage: oven <id>")
				continue
			}
			sc.Dispatch.MarkInOven(parts[1])
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
			sc.Skip(parts[1], reason)
			fmt.Printf("Burnt: %s\n", parts[1])

		case "holo":
			runHolo(sc.Holo)

		case "wave":
			if len(parts) < 3 {
				fmt.Println("Usage: wave <n> <m>")
				continue
			}
			n, m := 0, 0
			fmt.Sscanf(parts[1], "%d", &n)
			fmt.Sscanf(parts[2], "%d", &m)
			w := sc.Holo.Project(n, m, 4)
			fmt.Print(holo.FormatWave(w, 5, 9))

		case "proof":
			fmt.Print(sc.ProofLog())

		case "self":
			fmt.Print(sc.Introspect())

		case "log":
			fmt.Print(sc.Dispatch.Logger.FormatLog())
			fmt.Print(sc.Dispatch.Logger.Summary())

		case "legend":
			printLegend()

		case "save":
			if err := sc.Save(); err != nil {
				fmt.Printf("Save error: %v\n", err)
			} else {
				fmt.Println("Memory saved.")
			}

		case "quit", "exit":
			sc.Save()
			fmt.Println("Kitchen closed. Memory saved.")
			return

		default:
			fmt.Printf("Unknown: %s (type 'help')\n", cmd)
		}
	}
}
