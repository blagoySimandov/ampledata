package prompts

import _ "embed"

//go:embed decision-maker.txt
var DecisionMaker string

//go:embed extraction.txt
var Extraction string

//go:embed query-pattern.txt
var QueryPattern string

//go:embed key-selector.txt
var KeySelector string

//go:embed source-name.txt
var SourceName string
