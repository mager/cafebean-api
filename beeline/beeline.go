package beeline

import (
	beeline "github.com/honeycombio/beeline-go"
)

// ProvideBeeline provides a Honeycomb client
func ProvideBeeline() beeline.Config  {
	return beeline.Config{
        WriteKey: "6d3b823188e2f0107f230b7b30cdc1ef",
        Dataset: "cafebean",
    }
}

var Options = ProvideBeeline
