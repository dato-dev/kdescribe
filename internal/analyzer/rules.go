package analyzer

type Severity string

const (
	Critical Severity = "CRITICAL"
	Warning  Severity = "WARNING"
	Info     Severity = "INFO"
)

type Rule struct {
	Pattern  string
	Severity Severity
	Message  string
}

var Rules = []Rule{
	{
		Pattern:  "OOMKilled",
		Severity: Critical,
		Message:  "Container was OOMKilled",
	},
	{
		Pattern:  "CrashLoopBackOff",
		Severity: Critical,
		Message:  "Container in CrashLoopBackOff",
	},
	{
		Pattern:  "ImagePullBackOff",
		Severity: Critical,
		Message:  "Image pull failed",
	},
	{
		Pattern:  "ErrImagePull",
		Severity: Critical,
		Message:  "Image pull error",
	},
	{
		Pattern:  "FailedMount",
		Severity: Warning,
		Message:  "Volume mount failed",
	},
	{
		Pattern:  "FailedScheduling",
		Severity: Warning,
		Message:  "Pod scheduling failed",
	},
	{
		Pattern:  "Readiness probe failed",
		Severity: Warning,
		Message:  "Readiness probe failing",
	},
	{
		Pattern:  "Liveness probe failed",
		Severity: Warning,
		Message:  "Liveness probe failing",
	},
}
