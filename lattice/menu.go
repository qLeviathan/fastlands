// Package lattice — menu.go contains the full pizza menu.
// Every slice on the menu is a real task with real coordinates.
// The menu is the legend. Only the chef knows what each topping means.
package lattice

// LoadFullMenu populates the grid with every actionable slice.
// The menu is organized by oven temperature (effort shell).
// Low-temp slices come out fastest. High-temp slices are the artisan specials.
func LoadFullMenu(g *Grid) {
	// ═══════════════════════════════════════════════════════════
	// TEMPERATURE 0-2: Express window slices (fastest out the oven)
	// ═══════════════════════════════════════════════════════════

	g.Add(&Slice{
		ID: "tc-mar-001", Name: "Classic Margherita",
		Dough: ThinCrust, Sauce: Marinara,
		Toppings: []Topping{Pepperoni, Jalapeno},
		Status: Raw,
		Notes: "Sign up, pass starter assessment, begin same day. Pays every 3 days.",
	})

	g.Add(&Slice{
		ID: "tc-alf-001", Name: "White Pizza",
		Dough: ThinCrust, Sauce: Alfredo,
		Toppings: []Topping{Pepperoni},
		Status: Raw,
		Notes: "Google account signup, ID + resume upload, domain assessments. Weekly pay.",
	})

	g.Add(&Slice{
		ID: "tc-tru-001", Name: "Truffle Flatbread",
		Dough: ThinCrust, Sauce: Truffle,
		Toppings: []Topping{Pepperoni, Mushroom},
		Status: Raw,
		Notes: "Identity verify, AI interview with Zara, domain assessments. Weekly via Deel. One-shot assessments — no retakes.",
	})

	g.Add(&Slice{
		ID: "tc-buf-001", Name: "Buffalo Thin",
		Dough: ThinCrust, Sauce: Buffalo,
		Toppings: []Topping{Olive},
		Status: Raw,
		Notes: "5-min apply, no prior experience needed. $15-100/hr. Bi-weekly.",
	})

	// ═══════════════════════════════════════════════════════════
	// TEMPERATURE 3-5: Standard menu (3-14 day oven time)
	// ═══════════════════════════════════════════════════════════

	g.Add(&Slice{
		ID: "ht-pes-001", Name: "Pesto Artisan",
		Dough: HandTossed, Sauce: Pesto,
		Toppings: []Topping{Pepperoni, Jalapeno},
		Status: Raw,
		Notes: "Free signup, build profile emphasizing rare combo. 10+ personalized proposals daily. Apply within 15-20 min of new posts. First gig 3-7 days. 10% house fee.",
	})

	g.Add(&Slice{
		ID: "ht-gb-001", Name: "Garlic Knot Special",
		Dough: HandTossed, Sauce: GarlicButter,
		Toppings: []Topping{Jalapeno},
		Status: Raw,
		Notes: "Bids within 60 seconds of posting. 8 free bids/month. Target payment-verified. First project 1-3 days.",
	})

	g.Add(&Slice{
		ID: "ht-bbq-001", Name: "BBQ Chicken",
		Dough: HandTossed, Sauce: BBQ,
		Toppings: []Topping{Olive, Basil},
		Status: Raw,
		Notes: "Create 3-4 fixed-price listings. #1 category. +18,347% search growth on agent gigs. First organic order 2-4 weeks. 20% house fee. Accelerate via social sharing.",
	})

	g.Add(&Slice{
		ID: "ht-buf-001", Name: "Buffalo Ranch Toss",
		Dough: HandTossed, Sauce: Buffalo,
		Toppings: []Topping{Basil, Olive},
		Status: Raw,
		Notes: "$5/mo membership, connect payment processor, publish immediately. $15-30 per 1K views. Write 8-15 min reads. Scale to $100-500/mo in 3-6 months.",
	})

	g.Add(&Slice{
		ID: "dd-pes-001", Name: "Deep Dish Pesto",
		Dough: DeepDish, Sauce: Pesto,
		Toppings: []Topping{Mushroom, Pepperoni},
		Status: Raw,
		Notes: "7,567+ open gigs for technical docs. $30-150/hr. Tools: ReadMe, Mintlify, GitBook, Swagger. Premium niche for hardware specs + MLOps.",
	})

	g.Add(&Slice{
		ID: "dd-mar-001", Name: "Deep Dish Classic",
		Dough: DeepDish, Sauce: Marinara,
		Toppings: []Topping{Mushroom},
		Status: Raw,
		Notes: "Expert track for domain evaluation. $40-50/hr tier. Economics + quantitative analysis track.",
	})

	g.Add(&Slice{
		ID: "dd-alf-001", Name: "Deep Dish Alfredo",
		Dough: DeepDish, Sauce: Alfredo,
		Toppings: []Topping{Mushroom},
		Status: Raw,
		Notes: "Prompt engineering $32/hr, SW engineering $43/hr, specialists $50-65/hr. Domain: economics reasoning.",
	})

	g.Add(&Slice{
		ID: "sc-pes-001", Name: "Stuffed Crust Pesto",
		Dough: StuffedCrust, Sauce: Pesto,
		Toppings: []Topping{Pepperoni, Mushroom},
		Status: Raw,
		Notes: "RAG agent optimization, AI SaaS audits, voice AI setups, quick coding tasks. $35-200/hr for AI/ML engineering.",
	})

	g.Add(&Slice{
		ID: "sc-bbq-001", Name: "Stuffed BBQ",
		Dough: StuffedCrust, Sauce: BBQ,
		Toppings: []Topping{Mushroom, Basil},
		Status: Raw,
		Notes: "AI consulting, prompt engineering, agent building gigs. $30-125/project starting.",
	})

	g.Add(&Slice{
		ID: "sc-mar-001", Name: "Stuffed Marinara",
		Dough: StuffedCrust, Sauce: Marinara,
		Toppings: []Topping{Mushroom},
		Status: Raw,
		Notes: "Coding evaluation expert track. Highest-paying tier. Combined with domain expertise.",
	})

	// ═══════════════════════════════════════════════════════════
	// TEMPERATURE 6+: Artisan specials (high reward, longer oven)
	// ═══════════════════════════════════════════════════════════

	g.Add(&Slice{
		ID: "neo-vod-001", Name: "Neapolitan Vodka",
		Dough: Neapolitan, Sauce: Vodka,
		Toppings: []Topping{Mushroom},
		Status: Raw,
		Notes: "Register as council member. Set rate $200-300/hr. Phone consultations. $150-500/hr paid per minute. Register with all 3 major networks simultaneously. First engagement 1-4 weeks.",
	})

	g.Add(&Slice{
		ID: "neo-ran-001", Name: "Neapolitan Ranch",
		Dough: Neapolitan, Sauce: Ranch,
		Toppings: []Topping{Mushroom, Basil},
		Status: Raw,
		Notes: "$100K prize pool ($50K first place). Predicting excess returns using ML + proprietary signals. Scored on modified Sharpe ratio. Deadline June 16, 2026. Download data today.",
	})

	g.Add(&Slice{
		ID: "neo-hh-001", Name: "Hot Honey Neapolitan",
		Dough: Neapolitan, Sauce: HotHoney,
		Toppings: []Topping{Basil, Olive},
		Status: Raw,
		Notes: "Free plan supports 2,500 subs with unlimited sends. Built-in boost marketplace — one creator earned $25K/month from boosts alone. Niche: hardware + economics crossover.",
	})

	g.Add(&Slice{
		ID: "dd-buf-001", Name: "Deep Dish Buffalo",
		Dough: DeepDish, Sauce: Buffalo,
		Toppings: []Topping{Basil, Mushroom},
		Status: Raw,
		Notes: "Up to $500/article for ML/data tutorials (Neptune). $900/article for data engineering (Airbyte). $500+50% bonus for high views (Semaphore). Apply with writing samples. Pay in 1-2 weeks.",
	})

	g.Add(&Slice{
		ID: "sc-alf-001", Name: "Stuffed Alfredo",
		Dough: StuffedCrust, Sauce: Alfredo,
		Toppings: []Topping{Mushroom},
		Status: Raw,
		Notes: "Cognitive labeling, red teaming, domain-specific prompt writing. $25-100/hr, up to $150/hr for complex reasoning tasks.",
	})
}

// LoadExpressOrder returns just the first 48-hour priority slices.
// These are the ones that come out of the oven fastest.
func LoadExpressOrder(g *Grid) {
	// Hour 1-2: Immediate oven
	g.Add(&Slice{
		ID: "express-001", Name: "Speed Margherita",
		Dough: ThinCrust, Sauce: Marinara,
		Toppings: []Topping{Pepperoni, Jalapeno},
		Status: Raw,
		Notes: "HOUR 1: Sign up, pass 30-60 min assessment, begin earning. PayPal every 3 days.",
	})
	g.Add(&Slice{
		ID: "express-002", Name: "Speed White",
		Dough: ThinCrust, Sauce: Alfredo,
		Toppings: []Topping{Pepperoni},
		Status: Raw,
		Notes: "HOUR 1: Google signup, upload ID + resume, take domain assessments. Weekly PayPal.",
	})
	g.Add(&Slice{
		ID: "express-003", Name: "Speed Pesto",
		Dough: HandTossed, Sauce: Pesto,
		Toppings: []Topping{Pepperoni, Jalapeno},
		Status: Raw,
		Notes: "HOUR 2: Build profile. Bid 10+ urgent gigs immediately. First gig 3-7 days.",
	})

	// Hour 3-4: Second batch
	g.Add(&Slice{
		ID: "express-004", Name: "Speed Garlic",
		Dough: HandTossed, Sauce: GarlicButter,
		Toppings: []Topping{Jalapeno},
		Status: Raw,
		Notes: "HOUR 3: Profile + aggressive bids on quick-turnaround projects.",
	})
	g.Add(&Slice{
		ID: "express-005", Name: "Speed BBQ",
		Dough: HandTossed, Sauce: BBQ,
		Toppings: []Topping{Basil},
		Status: Raw,
		Notes: "HOUR 3: Create 3-4 fixed-price listings in top category.",
	})
	g.Add(&Slice{
		ID: "express-006", Name: "Speed Truffle",
		Dough: ThinCrust, Sauce: Truffle,
		Toppings: []Topping{Pepperoni, Mushroom},
		Status: Raw,
		Notes: "HOUR 4: Identity verify, AI interview, one-shot domain assessment. Weekly via Deel.",
	})
	g.Add(&Slice{
		ID: "express-007", Name: "Speed Buffalo Thin",
		Dough: ThinCrust, Sauce: Buffalo,
		Toppings: []Topping{Olive},
		Status: Raw,
		Notes: "HOUR 4: 5-min apply. No experience needed. Bi-weekly pay.",
	})

	// Hour 5-6: Expert registrations + content
	g.Add(&Slice{
		ID: "express-008", Name: "Speed Vodka",
		Dough: Neapolitan, Sauce: Vodka,
		Toppings: []Topping{Mushroom},
		Status: Raw,
		Notes: "HOUR 5: Register with all 3 networks. Set rate $200-300/hr. Optimize LinkedIn.",
	})
	g.Add(&Slice{
		ID: "express-009", Name: "Speed Ranch",
		Dough: Neapolitan, Sauce: Ranch,
		Toppings: []Topping{Mushroom, Basil},
		Status: Raw,
		Notes: "HOUR 5: Sign up, download competition data.",
	})
	g.Add(&Slice{
		ID: "express-010", Name: "Speed Buffalo Toss",
		Dough: HandTossed, Sauce: Buffalo,
		Toppings: []Topping{Basil},
		Status: Raw,
		Notes: "HOUR 6: $5/mo membership, connect payment, publish first article on familiar AI topic.",
	})
}
