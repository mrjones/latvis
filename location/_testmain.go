package main

import "github.com/mrjones/latvis/location"
import "testing"
import __regexp__ "regexp"

var tests = []testing.InternalTest{
	{"location.TestContains", location.TestContains},
}
var benchmarks = []testing.InternalBenchmark{ //
}

func main() {
	testing.Main(__regexp__.MatchString, tests)
	testing.RunBenchmarks(__regexp__.MatchString, benchmarks)
}
