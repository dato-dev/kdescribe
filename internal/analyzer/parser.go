package analyzer

import (
	"strconv"
	"strings"
)

type Document struct {
	Lines      []Line
	Sections   []Section
	Containers []Container
	Conditions []Condition
	Workloads  []Workload
}

type Line struct {
	No   int
	Text string
}

type Section struct {
	Name      string
	StartLine int
	EndLine   int
}

type Container struct {
	Name         string
	StateReason  string
	StateLine    Line
	RestartCount int
	RestartLine  Line
	Ready        string
	ReadyLine    Line
}

type Condition struct {
	Type   string
	Status string
	Line   Line
}

type Workload struct {
	Kind      string
	Namespace string
	Name      string
	StartLine int
	EndLine   int
}

func Parse(input string) Document {
	rawLines := strings.Split(input, "\n")
	doc := Document{
		Lines: make([]Line, 0, len(rawLines)),
	}

	for index, text := range rawLines {
		doc.Lines = append(doc.Lines, Line{No: index + 1, Text: text})
	}

	doc.Sections = parseSections(doc.Lines)
	doc.Containers = parseContainers(doc)
	doc.Conditions = parseConditions(doc)
	doc.Workloads = parseWorkloads(doc.Lines)

	return doc
}

func (d Document) SectionForLine(lineNo int) string {
	for _, section := range d.Sections {
		if lineNo >= section.StartLine && lineNo <= section.EndLine {
			return section.Name
		}
	}

	return ""
}

func (d Document) EvidenceAround(lineNo int, radius int) []Evidence {
	if lineNo <= 0 {
		return nil
	}

	start := lineNo - radius
	if start < 1 {
		start = 1
	}

	end := lineNo + radius
	if end > len(d.Lines) {
		end = len(d.Lines)
	}

	evidence := make([]Evidence, 0, end-start+1)
	for line := start; line <= end; line++ {
		text := strings.TrimSpace(d.Lines[line-1].Text)
		if text == "" {
			continue
		}

		evidence = append(evidence, Evidence{
			LineNo: line,
			Text:   text,
		})
	}

	return evidence
}

func (d Document) WorkloadForLine(lineNo int) Workload {
	for _, workload := range d.Workloads {
		if lineNo >= workload.StartLine && lineNo <= workload.EndLine {
			return workload
		}
	}

	return Workload{}
}

func parseSections(lines []Line) []Section {
	var sections []Section

	for _, line := range lines {
		text := strings.TrimSpace(line.Text)
		if text == "" || strings.HasPrefix(line.Text, " ") || !strings.HasSuffix(text, ":") {
			continue
		}

		if len(sections) > 0 {
			sections[len(sections)-1].EndLine = line.No - 1
		}

		sections = append(sections, Section{
			Name:      strings.TrimSuffix(text, ":"),
			StartLine: line.No,
			EndLine:   len(lines),
		})
	}

	return sections
}

func parseContainers(doc Document) []Container {
	section := findSection(doc, "Containers")
	if section == nil {
		return nil
	}

	var containers []Container
	var current *Container
	lastField := ""

	for _, line := range doc.Lines[section.StartLine:section.EndLine] {
		text := strings.TrimSpace(line.Text)
		if text == "" {
			continue
		}

		indent := leadingSpaces(line.Text)
		if indent == 2 && strings.HasSuffix(text, ":") {
			containers = append(containers, Container{Name: strings.TrimSuffix(text, ":")})
			current = &containers[len(containers)-1]
			lastField = ""
			continue
		}

		if current == nil {
			continue
		}

		switch {
		case strings.HasPrefix(text, "State:"):
			lastField = "State"
		case strings.HasPrefix(text, "Ready:"):
			current.Ready = strings.TrimSpace(strings.TrimPrefix(text, "Ready:"))
			current.ReadyLine = line
			lastField = "Ready"
		case strings.HasPrefix(text, "Restart Count:"):
			value := strings.TrimSpace(strings.TrimPrefix(text, "Restart Count:"))
			if restartCount, err := strconv.Atoi(value); err == nil {
				current.RestartCount = restartCount
				current.RestartLine = line
			}
			lastField = "Restart Count"
		case strings.HasPrefix(text, "Reason:") && lastField == "State":
			current.StateReason = strings.TrimSpace(strings.TrimPrefix(text, "Reason:"))
			current.StateLine = line
		}
	}

	return containers
}

func parseConditions(doc Document) []Condition {
	section := findSection(doc, "Conditions")
	if section == nil {
		return nil
	}

	var conditions []Condition
	for _, line := range doc.Lines[section.StartLine:section.EndLine] {
		fields := strings.Fields(line.Text)
		if len(fields) < 2 || fields[0] == "Type" {
			continue
		}

		conditions = append(conditions, Condition{
			Type:   fields[0],
			Status: fields[1],
			Line:   line,
		})
	}

	return conditions
}

func findSection(doc Document, name string) *Section {
	for index := range doc.Sections {
		if doc.Sections[index].Name == name {
			return &doc.Sections[index]
		}
	}

	return nil
}

func leadingSpaces(text string) int {
	return len(text) - len(strings.TrimLeft(text, " "))
}

func parseWorkloads(lines []Line) []Workload {
	var workloads []Workload
	currentIndex := -1

	for _, line := range lines {
		text := strings.TrimSpace(line.Text)
		if text == "" || strings.HasPrefix(line.Text, " ") {
			continue
		}

		switch {
		case strings.HasPrefix(text, "Name:"):
			if currentIndex >= 0 {
				workloads[currentIndex].EndLine = line.No - 1
			}

			workloads = append(workloads, Workload{
				Name:      strings.TrimSpace(strings.TrimPrefix(text, "Name:")),
				StartLine: line.No,
				EndLine:   len(lines),
			})
			currentIndex = len(workloads) - 1
		case strings.HasPrefix(text, "Namespace:") && currentIndex >= 0:
			workloads[currentIndex].Namespace = strings.TrimSpace(strings.TrimPrefix(text, "Namespace:"))
		case strings.HasPrefix(text, "Kind:") && currentIndex >= 0:
			workloads[currentIndex].Kind = strings.TrimSpace(strings.TrimPrefix(text, "Kind:"))
		}
	}

	return workloads
}
