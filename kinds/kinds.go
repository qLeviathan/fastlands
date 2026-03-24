package kinds

// Kind represents a reasoning technique
type Kind struct {
	ID     string
	Name   string
	Parent string
}

var (
	KindLogicPrograms = Kind{ID: "ar-lp", Name: "Logic Programs"}
	KindChainEval     = Kind{ID: "ar-ce", Name: "Chain Evaluation"}
	KindPatternMatch  = Kind{ID: "ar-pm", Name: "Pattern Matching"}
)
