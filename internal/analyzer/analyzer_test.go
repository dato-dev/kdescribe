package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzeFindsAndRanksCriticalSignals(t *testing.T) {
	input := `Name: demo
State: Waiting
  Reason: ImagePullBackOff
Last State: Terminated
  Reason: OOMKilled
Events:
  Warning FailedScheduling 0/3 nodes are available: insufficient cpu.`

	findings := Analyze(input)

	if len(findings) < 3 {
		t.Fatalf("expected at least 3 findings, got %d", len(findings))
	}

	if findings[0].Message != "Container was OOMKilled" {
		t.Fatalf("expected OOMKilled to be ranked first, got %q", findings[0].Message)
	}

	if findings[0].LineNo != 5 {
		t.Fatalf("expected OOMKilled on line 5, got %d", findings[0].LineNo)
	}

	if RiskLevel(findings) != "HIGH" {
		t.Fatalf("expected HIGH risk, got %s", RiskLevel(findings))
	}
}

func TestAnalyzeDeduplicatesSameRuleAndLine(t *testing.T) {
	input := `Events:
  Warning BackOff Back-off restarting failed container api
  Warning BackOff Back-off restarting failed container api`

	findings := Analyze(input)

	if len(findings) != 1 {
		t.Fatalf("expected duplicate event line to collapse into 1 finding, got %d", len(findings))
	}

	if findings[0].Severity != Critical {
		t.Fatalf("expected critical finding, got %s", findings[0].Severity)
	}
}

func TestRiskLevelLowForNoFindings(t *testing.T) {
	if RiskLevel(nil) != "LOW" {
		t.Fatalf("expected LOW risk without findings")
	}
}

func TestAnalyzeFixtureCrashLoop(t *testing.T) {
	input := readFixture(t, "crashloop-pod.txt")
	analysis := AnalyzeDocument(input)

	if analysis.Risk.Level != "HIGH" {
		t.Fatalf("expected HIGH risk, got %s", analysis.Risk.Level)
	}

	assertHasFinding(t, analysis.Findings, "crash-loop")
	assertHasFinding(t, analysis.Findings, "high-restart-count")
	assertHasFinding(t, analysis.Findings, "pod-ready-false")

	crashLoop := findByID(t, analysis.Findings, "crash-loop")
	if crashLoop.Workload != "default/api-7f9f9d9c8b-2lq4m" {
		t.Fatalf("expected workload label default/api-7f9f9d9c8b-2lq4m, got %q", crashLoop.Workload)
	}
}

func TestAnalyzeFixtureScheduling(t *testing.T) {
	input := readFixture(t, "pending-scheduling.txt")
	analysis := AnalyzeDocument(input)

	assertHasFinding(t, analysis.Findings, "failed-scheduling")

	if analysis.Risk.Level != "MEDIUM" {
		t.Fatalf("expected MEDIUM risk, got %s", analysis.Risk.Level)
	}
}

func TestParseExtractsSectionsAndContainers(t *testing.T) {
	input := readFixture(t, "crashloop-pod.txt")
	doc := Parse(input)

	if doc.SectionForLine(15) != "Containers" {
		t.Fatalf("expected line 15 to be in Containers section, got %q", doc.SectionForLine(15))
	}

	if len(doc.Containers) != 1 {
		t.Fatalf("expected 1 container, got %d", len(doc.Containers))
	}

	container := doc.Containers[0]
	if container.RestartCount != 12 {
		t.Fatalf("expected restart count 12, got %d", container.RestartCount)
	}

	if container.Ready != "False" {
		t.Fatalf("expected container ready false, got %q", container.Ready)
	}
}

func TestParseExtractsWorkloads(t *testing.T) {
	input := readFixture(t, "pending-scheduling.txt")
	doc := Parse(input)

	if len(doc.Workloads) != 1 {
		t.Fatalf("expected 1 workload, got %d", len(doc.Workloads))
	}

	workload := doc.WorkloadForLine(18)
	if workload.Name != "worker-0" || workload.Namespace != "default" {
		t.Fatalf("unexpected workload parsed: %#v", workload)
	}
}

func readFixture(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}

	return string(data)
}

func assertHasFinding(t *testing.T, findings []Finding, id string) {
	t.Helper()

	for _, finding := range findings {
		if finding.ID == id {
			return
		}
	}

	t.Fatalf("expected finding %q, got %#v", id, findings)
}

func findByID(t *testing.T, findings []Finding, id string) Finding {
	t.Helper()

	for _, finding := range findings {
		if finding.ID == id {
			return finding
		}
	}

	t.Fatalf("expected finding %q", id)
	return Finding{}
}
