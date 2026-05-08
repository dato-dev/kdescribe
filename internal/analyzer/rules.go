package analyzer

import "regexp"

type Severity string

type Domain string

type Confidence string

const (
	Critical Severity = "CRITICAL"
	Warning  Severity = "WARNING"
	Info     Severity = "INFO"
)

const (
	DomainImages     Domain = "images"
	DomainNetworking Domain = "networking"
	DomainProbes     Domain = "probes"
	DomainResources  Domain = "resources"
	DomainRuntime    Domain = "runtime"
	DomainScheduling Domain = "scheduling"
	DomainStorage    Domain = "storage"
)

const (
	RuleBased  Confidence = "rule-based"
	Structured Confidence = "structured"
)

type Rule struct {
	ID       string
	Domain   Domain
	Pattern  *regexp.Regexp
	Severity Severity
	Score    int
	Message  string
	Hint     string
}

var Rules = append(
	append(
		append(
			append(imageRules, schedulingRules...),
			probeRules...,
		),
		storageRules...,
	),
	append(resourceRules, runtimeRules...)...,
)

var imageRules = []Rule{
	{
		ID:       "image-pull-backoff",
		Domain:   DomainImages,
		Pattern:  regexp.MustCompile(`\bImagePullBackOff\b`),
		Severity: Critical,
		Score:    90,
		Message:  "Image pull failed",
		Hint:     "Verify image name, tag, registry access, and imagePullSecrets.",
	},
	{
		ID:       "err-image-pull",
		Domain:   DomainImages,
		Pattern:  regexp.MustCompile(`\bErrImagePull\b`),
		Severity: Critical,
		Score:    90,
		Message:  "Image pull error",
		Hint:     "Verify image name, tag, registry access, and imagePullSecrets.",
	},
}

var runtimeRules = []Rule{
	{
		ID:       "oom-killed",
		Domain:   DomainResources,
		Pattern:  regexp.MustCompile(`\bOOMKilled\b`),
		Severity: Critical,
		Score:    100,
		Message:  "Container was OOMKilled",
		Hint:     "Check memory limits/requests and recent memory usage.",
	},
	{
		ID:       "crash-loop",
		Domain:   DomainRuntime,
		Pattern:  regexp.MustCompile(`\bCrashLoopBackOff\b|Back-off restarting failed container`),
		Severity: Critical,
		Score:    100,
		Message:  "Container in CrashLoopBackOff",
		Hint:     "Inspect previous logs, exit code, probes, and startup command.",
	},
	{
		ID:       "failed-pod-sandbox",
		Domain:   DomainNetworking,
		Pattern:  regexp.MustCompile(`\bFailedCreatePodSandBox\b|failed to create pod sandbox`),
		Severity: Critical,
		Score:    85,
		Message:  "Pod sandbox creation failed",
		Hint:     "Check CNI/runtime health on the node.",
	},
}

var storageRules = []Rule{
	{
		ID:       "failed-mount",
		Domain:   DomainStorage,
		Pattern:  regexp.MustCompile(`\bFailedMount\b|MountVolume\..*failed`),
		Severity: Warning,
		Score:    80,
		Message:  "Volume mount failed",
		Hint:     "Check PVC/PV binding, secrets/configmaps, and storage backend events.",
	},
	{
		ID:       "failed-attach-volume",
		Domain:   DomainStorage,
		Pattern:  regexp.MustCompile(`\bFailedAttachVolume\b|AttachVolume\..*failed`),
		Severity: Warning,
		Score:    80,
		Message:  "Volume attach failed",
		Hint:     "Check node/storage attachment limits and CSI controller events.",
	},
}

var schedulingRules = []Rule{
	{
		ID:       "failed-scheduling",
		Domain:   DomainScheduling,
		Pattern:  regexp.MustCompile(`\bFailedScheduling\b`),
		Severity: Warning,
		Score:    70,
		Message:  "Pod scheduling failed",
		Hint:     "Check resource requests, taints/tolerations, affinity, and node capacity.",
	},
}

var probeRules = []Rule{
	{
		ID:       "unhealthy",
		Domain:   DomainProbes,
		Pattern:  regexp.MustCompile(`\bUnhealthy\b`),
		Severity: Warning,
		Score:    60,
		Message:  "Kubelet reported container unhealthy",
		Hint:     "Probe failures often point to slow startup, wrong port/path, or app errors.",
	},
	{
		ID:       "readiness-probe",
		Domain:   DomainProbes,
		Pattern:  regexp.MustCompile(`Readiness probe failed`),
		Severity: Warning,
		Score:    60,
		Message:  "Readiness probe failing",
		Hint:     "Check probe path/port and whether the app is actually ready.",
	},
	{
		ID:       "liveness-probe",
		Domain:   DomainProbes,
		Pattern:  regexp.MustCompile(`Liveness probe failed`),
		Severity: Warning,
		Score:    60,
		Message:  "Liveness probe failing",
		Hint:     "A failing liveness probe can cause restart loops.",
	},
}

var resourceRules = []Rule{
	{
		ID:       "evicted",
		Domain:   DomainResources,
		Pattern:  regexp.MustCompile(`\bEvicted\b|The node was low on resource`),
		Severity: Warning,
		Score:    75,
		Message:  "Pod was evicted",
		Hint:     "Check node pressure, requests, limits, and QoS class.",
	},
}
