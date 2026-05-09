package cmd

import "testing"

func TestParseArgsSeparatesKdescribeAndKubectlFlags(t *testing.T) {
	resetOptionsForTest()

	describeArgs, err := parseArgs([]string{
		"pod",
		"nginx",
		"-n",
		"default",
		"--output",
		"json",
		"--min-score=80",
		"--show-all",
	})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"pod", "nginx", "-n", "default"}
	if !equalStringSlices(describeArgs, expected) {
		t.Fatalf("expected describe args %#v, got %#v", expected, describeArgs)
	}

	if options.output != "json" {
		t.Fatalf("expected output json, got %q", options.output)
	}

	if options.minScore != 80 {
		t.Fatalf("expected min score 80, got %d", options.minScore)
	}

	if !options.showAll {
		t.Fatalf("expected showAll to be true")
	}
}

func TestParseArgsPassesEverythingAfterDoubleDash(t *testing.T) {
	resetOptionsForTest()

	describeArgs, err := parseArgs([]string{"--no-color", "--", "pod", "nginx", "--output=json"})
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{"pod", "nginx", "--output=json"}
	if !equalStringSlices(describeArgs, expected) {
		t.Fatalf("expected describe args %#v, got %#v", expected, describeArgs)
	}

	if !options.noColor {
		t.Fatalf("expected noColor to be true")
	}
}

func resetOptionsForTest() {
	options.output = "human"
	options.noColor = false
	options.minScore = 0
	options.exitCode = false
	options.showAll = false
	options.help = false
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}

	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}

	return true
}
