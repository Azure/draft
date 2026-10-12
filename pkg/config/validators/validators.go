package validators

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"k8s.io/apimachinery/pkg/util/validation"
)

const (
	// maxFileNameLength is the conventional per-component filename limit on
	// the filesystems draft targets.
	maxFileNameLength = 255
	// maxFilePathLength is the conventional total path length limit.
	maxFilePathLength = 4096
	// maxImageTagLength is the OCI limit on a container image tag.
	maxImageTagLength = 128
	// maxPort is the highest valid TCP/UDP port.
	maxPort = 65535

	// Docker reference name grammar (namePat), without tags or digests:
	// https://github.com/distribution/reference/blob/main/regexp.go
	imageDomainComponent = `(?:[a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9-]*[a-zA-Z0-9])`
	imageHost            = `(?:` + imageDomainComponent + `(?:\.` + imageDomainComponent + `)*|\[(?:[a-fA-F0-9:]+)\])`
	imagePathComponent   = `[a-z0-9]+(?:(?:[._]|__|[-]+)[a-z0-9]+)*`
	imageName            = `(?:` + imageHost + `(?::[0-9]+)?/)?` + imagePathComponent + `(?:/` + imagePathComponent + `)*`
)

var (
	ciNamePattern             = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._/-]*$`)
	containerImageNamePattern = regexp.MustCompile(`^` + imageName + `$`)
	helmChartOverridesPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+=[A-Za-z0-9_.:/+-]+(?:,[A-Za-z0-9_.-]+=[A-Za-z0-9_.:/+-]+)*$`)
)

// safeNameRunes reports whether r is allowed in a single file or directory
// name. The set is deliberately narrow: it excludes path separators, quotes,
// whitespace and every shell, YAML and JSON metacharacter, so a value that
// passes cannot terminate the syntactic context it is interpolated into.
func safeNameRunes(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z':
		return true
	case r >= 'A' && r <= 'Z':
		return true
	case r >= '0' && r <= '9':
		return true
	case r == '.' || r == '_' || r == '-':
		return true
	default:
		return false
	}
}

func GetValidator(variableKind string) func(string) error {
	switch variableKind {
	case "azureContainerRegistry", "azureManagedCluster", "azureResourceGroup":
		return ciNameValidator
	case "containerImageName":
		return containerImageNameValidator
	case "containerImageVersion":
		return containerImageVersionValidator
	case "dirPath":
		return dirPathValidator
	case "envVarMap":
		return keyValueMapValidator
	case "fileName":
		return fileNameValidator
	case "filePath":
		return filePathValidator
	case "imagePullPolicy":
		return imagePullPolicyValidator
	case "helmChartOverrides":
		return helmChartOverridesValidator
	case "kubernetesNamespace", "kubernetesResourceName":
		return kubernetesDNSLabelValidator
	case "kubernetesProbeType":
		return kubernetesProbeTypeValidator
	case "port":
		return portValidator
	case "pythonEntryPoint":
		return pythonEntryPointValidator
	case "repositoryBranch":
		return repositoryBranchValidator
	case "scalingResourceType":
		return scalingResourceTypeValidator
	default:
		return defaultValidator
	}
}

func ciNameValidator(input string) error {
	if !ciNamePattern.MatchString(input) {
		return fmt.Errorf("invalid CI resource value %q: contains unsupported characters", input)
	}
	return nil
}

// Image tags are supplied separately by containerImageVersion. Anchor Docker's
// name grammar so a valid substring cannot conceal trailing shell syntax.
func containerImageNameValidator(input string) error {
	if !containerImageNamePattern.MatchString(input) {
		return fmt.Errorf("invalid container image name %q: expected a repository name with an optional registry and port", input)
	}
	return nil
}

func kubernetesDNSLabelValidator(input string) error {
	if errs := validation.IsDNS1123Label(input); len(errs) > 0 {
		return fmt.Errorf("invalid Kubernetes name %q: %s", input, strings.Join(errs, "; "))
	}
	return nil
}

func helmChartOverridesValidator(input string) error {
	if !helmChartOverridesPattern.MatchString(input) {
		return fmt.Errorf("invalid Helm chart overrides %q: expected comma-separated key=value assignments", input)
	}
	return nil
}

func imagePullPolicyValidator(input string) error {
	switch input {
	case "Always", "IfNotPresent", "Never":
		return nil
	default:
		return fmt.Errorf("invalid image pull policy: %s. valid values: Always, IfNotPresent, Never", input)
	}
}

func scalingResourceTypeValidator(input string) error {
	switch input {
	case "cpu", "memory":
		return nil
	default:
		return fmt.Errorf("invalid scaling resource type: %s. valid values: cpu, memory", input)
	}
}

func kubernetesProbeTypeValidator(input string) error {
	switch input {
	case "httpGet", "tcpSocket":
		return nil
	default:
		return fmt.Errorf("invalid probe type: %s. valid values: httpGet, tcpSocket", input)
	}
}

// keyValueMapValidator validates a JSON object of string keys to string values,
// such as ENVVARS. The parsed keys and values are written into generated YAML as
// raw scalars, so they are checked individually: a JSON "\n" escape survives
// parsing intact and would otherwise inject additional YAML keys.
func keyValueMapValidator(input string) error {
	if err := rejectUnsafeBytes(input); err != nil {
		return fmt.Errorf("invalid key value map: %w", err)
	}

	parsed := map[string]string{}
	if err := json.Unmarshal([]byte(input), &parsed); err != nil {
		return fmt.Errorf("failed to unmarshal variable as map[string]string: %s", err)
	}

	for key, value := range parsed {
		if err := rejectUnsafeBytes(key); err != nil {
			return fmt.Errorf("invalid key %q: %w", key, err)
		}
		if err := rejectUnsafeBytes(value); err != nil {
			return fmt.Errorf("invalid value for key %q: %w", key, err)
		}
	}
	return nil
}

// portValidator validates a value used as a numeric network port.
//
// The port is also interpolated into shell-form Dockerfile instructions, for
// example the csharp template's ENTRYPOINT, where an unconstrained value could
// terminate the quoted argument and append shell commands. Both accepted forms
// are restricted to digits, lowercase letters and hyphens, so neither can.
func portValidator(input string) error {
	if input == "" {
		return fmt.Errorf("invalid port: must not be empty")
	}

	if !isAllDigits(input) {
		return fmt.Errorf("invalid port %q: must be a number between 1 and %d", input, maxPort)
	}

	port, err := strconv.Atoi(input)
	if err != nil {
		return fmt.Errorf("invalid port %q: %w", input, err)
	}
	if port < 1 || port > maxPort {
		return fmt.Errorf("invalid port %d: must be between 1 and %d", port, maxPort)
	}

	return nil
}

func dirPathValidator(input string) error {
	if input == "" {
		return fmt.Errorf("invalid directory path: must not be empty")
	}
	if err := rejectUnsafeBytes(input); err != nil {
		return fmt.Errorf("invalid directory path: %w", err)
	}
	if strings.ContainsAny(input, `;&|$`+"`"+`<>!(){}[]?'"\\`) {
		return fmt.Errorf("invalid directory path %q: contains shell metacharacters", input)
	}
	if strings.HasPrefix(input, "-") {
		return fmt.Errorf("invalid directory path %q: must not start with '-'", input)
	}
	return nil
}

func repositoryBranchValidator(input string) error {
	if input == "" {
		return fmt.Errorf("invalid repository branch: must not be empty")
	}
	if err := rejectUnsafeBytes(input); err != nil {
		return fmt.Errorf("invalid repository branch: %w", err)
	}
	if strings.ContainsAny(input, " ~^:?*[\\,#!+") || strings.Contains(input, "..") ||
		strings.HasPrefix(input, ".") || strings.HasPrefix(input, "/") || strings.HasSuffix(input, "/") ||
		strings.HasSuffix(input, ".") || strings.HasSuffix(input, ".lock") || strings.Contains(input, "//") ||
		strings.Contains(input, "@{") {
		return fmt.Errorf("invalid repository branch %q", input)
	}
	return nil
}

func pythonEntryPointValidator(input string) error {
	if input == "" {
		return fmt.Errorf("invalid Python entrypoint: must not be empty")
	}
	if err := rejectUnsafeBytes(input); err != nil {
		return fmt.Errorf("invalid Python entrypoint: %w", err)
	}
	if strings.HasPrefix(input, "/") || strings.ContainsRune(input, '\\') {
		return fmt.Errorf("invalid Python entrypoint %q: must be a relative slash-separated path", input)
	}
	pathInput := strings.TrimPrefix(input, "./")
	if pathInput == "" {
		return fmt.Errorf("invalid Python entrypoint %q: must name a file", input)
	}
	for i, segment := range strings.Split(pathInput, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("invalid Python entrypoint %q: invalid path component", input)
		}
		if i == 0 && strings.HasPrefix(segment, "-") {
			return fmt.Errorf("invalid Python entrypoint %q: must not start with '-'", input)
		}
	}
	return nil
}

func isAllDigits(input string) bool {
	for _, r := range input {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// containerImageVersionValidator validates a value used as a container image
// tag, for example the VERSION interpolated into a Dockerfile FROM instruction.
// This value can be derived from the contents of an untrusted repository, so it
// is restricted to the character set the OCI distribution spec allows for tags.
func containerImageVersionValidator(input string) error {
	if input == "" {
		return fmt.Errorf("invalid container image version: must not be empty")
	}
	if len(input) > maxImageTagLength {
		return fmt.Errorf("invalid container image version: must be at most %d characters, got %d", maxImageTagLength, len(input))
	}
	if i := strings.IndexFunc(input, func(r rune) bool { return !safeNameRunes(r) }); i >= 0 {
		r, _ := utf8.DecodeRuneInString(input[i:])
		return fmt.Errorf(
			"invalid container image version %q: disallowed character %q at byte %d, tags may only contain letters, digits, '.', '_' and '-'",
			input, r, i)
	}
	// The charset check above guarantees input is ASCII, so indexing is safe.
	if first := input[0]; first == '.' || first == '-' {
		return fmt.Errorf("invalid container image version %q: must start with a letter, digit or '_'", input)
	}
	return nil
}

// fileNameValidator validates a value that is written into the generated
// artifact as a single file name, such as the python template's ENTRYPOINT.
// Such values are frequently derived from the contents of an untrusted
// repository, so the allowed character set is restricted to a safe basename
// charset. This rejects the quotes, brackets, whitespace and newlines that
// would otherwise let a crafted filename close the surrounding Dockerfile
// instruction and append attacker-controlled instructions.
func fileNameValidator(input string) error {
	if input == "" {
		return fmt.Errorf("invalid file name: must not be empty")
	}
	if err := rejectUnsafeBytes(input); err != nil {
		return fmt.Errorf("invalid file name: %w", err)
	}
	if len(input) > maxFileNameLength {
		return fmt.Errorf("invalid file name: must be at most %d characters, got %d", maxFileNameLength, len(input))
	}
	if input == "." || input == ".." {
		return fmt.Errorf("invalid file name %q: must name a file", input)
	}
	// A leading '-' would be parsed as an option by the command the file name is
	// passed to, for example turning CMD ["-mpkg"] into "python -m pkg".
	if strings.HasPrefix(input, "-") {
		return fmt.Errorf("invalid file name %q: must not start with '-'", input)
	}
	if i := strings.IndexFunc(input, func(r rune) bool { return !safeNameRunes(r) }); i >= 0 {
		r, _ := utf8.DecodeRuneInString(input[i:])
		return fmt.Errorf(
			"invalid file name %q: disallowed character %q at byte %d, file names may only contain letters, digits, '.', '_' and '-'",
			input, r, i)
	}
	return nil
}

// filePathValidator validates a value that is written into the generated
// artifact as a path, such as the python template's ENTRYPOINT or the DOCKERFILE
// variable in the workflow and pipeline templates. Each path segment is held to
// the same character restrictions as fileNameValidator, which is what prevents a
// crafted value from terminating the surrounding instruction or YAML scalar.
//
// Absolute paths and '..' segments are deliberately permitted. The path is
// written out for a build tool to resolve later and is never read by draft, so
// constraining its shape would reject legitimate layouts without closing any
// injection vector.
func filePathValidator(input string) error {
	if input == "" {
		return fmt.Errorf("invalid file path: must not be empty")
	}
	if err := rejectUnsafeBytes(input); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	if len(input) > maxFilePathLength {
		return fmt.Errorf("invalid file path: must be at most %d characters, got %d", maxFilePathLength, len(input))
	}
	if strings.ContainsRune(input, '\\') {
		return fmt.Errorf("invalid file path %q: backslashes are not allowed, use '/' as the path separator", input)
	}

	hasName := false
	for _, segment := range strings.Split(strings.TrimPrefix(input, "/"), "/") {
		switch segment {
		case "":
			return fmt.Errorf("invalid file path %q: empty path segment", input)
		case ".", "..":
			continue
		}
		if err := fileNameValidator(segment); err != nil {
			return fmt.Errorf("invalid file path %q: %w", input, err)
		}
		hasName = true
	}

	if !hasName {
		return fmt.Errorf("invalid file path %q: must name a file", input)
	}
	return nil
}

// defaultValidator is the fallback used by every variable kind that does not
// declare a more specific validator. Note that kinds with their own validator
// do not reach this function, so those validators are responsible for calling
// rejectUnsafeBytes themselves.
//
// It makes no assumptions about the shape of the value; it only rejects bytes
// that allow a value to break out of the syntactic context it is interpolated
// into, namely newlines and other control characters, which would let a value
// start a new Dockerfile instruction, YAML key or workflow step.
func defaultValidator(input string) error {
	return rejectUnsafeBytes(input)
}

// rejectUnsafeBytes rejects values that are not valid UTF-8 or that contain
// control characters, Unicode format characters or Unicode line separators.
func rejectUnsafeBytes(input string) error {
	if !utf8.ValidString(input) {
		return fmt.Errorf("value is not valid UTF-8")
	}
	for i, r := range input {
		switch {
		case r == '\n':
			return fmt.Errorf("value contains a newline at byte %d", i)
		case r == '\r':
			return fmt.Errorf("value contains a carriage return at byte %d", i)
		case r == 0:
			return fmt.Errorf("value contains a null byte at byte %d", i)
		case unicode.IsControl(r):
			return fmt.Errorf("value contains the control character %q at byte %d", r, i)
		case r == '\u2028' || r == '\u2029':
			return fmt.Errorf("value contains the Unicode line separator %q at byte %d", r, i)
		// Format characters cover zero-width and bidirectional overrides, which
		// make a generated file render differently from how it is interpreted.
		case unicode.Is(unicode.Cf, r):
			return fmt.Errorf("value contains the Unicode format character %q at byte %d", r, i)
		}
	}
	return nil
}
