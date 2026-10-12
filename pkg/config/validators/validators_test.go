package validators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidator(t *testing.T) {
	assert.NotNil(t, GetValidator("NonExistentKind"))
}

func TestDefaultValidator(t *testing.T) {
	// The baseline permits ordinary values of any shape...
	assert.Nil(t, defaultValidator("test"))
	assert.Nil(t, defaultValidator("./Dockerfile"))
	assert.Nil(t, defaultValidator("replicas=2,image.tag=latest"))
	assert.Nil(t, defaultValidator(`{"KEY": "value"}`))
	assert.Nil(t, defaultValidator("8080"))

	// ...but rejects the bytes that let a value escape the context it is
	// interpolated into.
	assert.NotNil(t, defaultValidator("line\nRUN touch pwned"))
	assert.NotNil(t, defaultValidator("carriage\rreturn"))
	assert.NotNil(t, defaultValidator("null\x00byte"))
	assert.NotNil(t, defaultValidator("bell\x07"))
	assert.NotNil(t, defaultValidator("escape\x1b[0m"))
	assert.NotNil(t, defaultValidator("line\u2028separator"))
	assert.NotNil(t, defaultValidator("paragraph\u2029separator"))
	assert.NotNil(t, defaultValidator("invalid\xc3\x28utf8"))
}

// TestFileNameValidator covers the ENTRYPOINT value in the python Dockerfile
// template, which is derived from the contents of an untrusted repository.
func TestFileNameValidator(t *testing.T) {
	valid := []string{
		"app.py",
		"main.py",
		"my_app.py",
		"my-app.py",
		"App2.py",
		"Dockerfile",
		".env",
	}
	for _, input := range valid {
		assert.Nil(t, fileNameValidator(input), "expected %q to be valid", input)
	}

	invalid := map[string]string{
		"empty":                 "",
		"dot":                   ".",
		"dotdot":                "..",
		"embedded quote":        `my"app.py`,
		"embedded single quote": "my'app.py",
		"embedded backslash":    `my\app.py`,
		"embedded newline":      "my\napp.py",
		"embedded space":        "my app.py",
		"embedded tab":          "my\tapp.py",
		"open bracket":          "my[app.py",
		"close bracket":         "my]app.py",
		"path separator":        "dir/app.py",
		"absolute path":         "/etc/passwd",
		"relative prefix":       "./app.py",
		"traversal":             "../app.py",
		"shell substitution":    "$(touch pwned).py",
		"shell backtick":        "`touch pwned`.py",
		"shell separator":       "app.py;touch pwned",
		"argument injection":    `main.py", "extra.py`,
		"instruction injection": "app.py\"]\nRUN touch pwned\n# .py",
		"too long":              strings.Repeat("a", maxFileNameLength+1),
	}
	for name, input := range invalid {
		assert.NotNil(t, fileNameValidator(input), "expected %q (%s) to be rejected", input, name)
	}
}

// TestFilePathValidator covers the ENTRYPOINT and DOCKERFILE values. Absolute
// paths and '..' segments are accepted: they are resolved by a build tool, not
// by draft, and constraining them closes no injection vector.
func TestFilePathValidator(t *testing.T) {
	valid := []string{
		"./Dockerfile",
		"Dockerfile",
		"./src/Dockerfile",
		"src/api/Dockerfile.prod",
		"app.py",
		"/abs/path/Dockerfile",
		"../Dockerfile",
		"../../shared/Dockerfile",
		"./src/../Dockerfile",
	}
	for _, input := range valid {
		assert.Nil(t, filePathValidator(input), "expected %q to be valid", input)
	}

	invalid := map[string]string{
		"empty":                "",
		"backslash":            `src\Dockerfile`,
		"windows absolute":     `C:\Windows\System32`,
		"windows drive slash":  "C:/Windows",
		"empty segment":        "src//Dockerfile",
		"trailing slash":       "src/",
		"embedded quote":       `./Docker"file`,
		"embedded newline":     "./Dockerfile\nRUN touch pwned",
		"embedded space":       "./my Dockerfile",
		"command substitution": "./$(touch pwned)",
		"dot only":             ".",
		"dotdot only":          "..",
		"dot slash only":       "./",
		"root only":            "/",
		"leading dash":         "-rf",
	}
	for name, input := range invalid {
		assert.NotNil(t, filePathValidator(input), "expected %q (%s) to be rejected", input, name)
	}
}

// TestPortValidator covers the PORT value, which the csharp template
// interpolates into a shell-form ENTRYPOINT. A port is either a number or an
// IANA service name, which Kubernetes accepts for a Service's targetPort.
func TestPortValidator(t *testing.T) {
	for _, input := range []string{"1", "80", "8080", "65535"} {
		assert.Nil(t, portValidator(input), "expected numeric port %q to be valid", input)
	}

	invalid := map[string]string{
		"empty":                "",
		"zero":                 "0",
		"above range":          "65536",
		"negative":             "-1",
		"signed":               "+80",
		"trailing space":       "80 ",
		"decimal":              "80.5",
		"overflow":             "99999999999999999999",
		"uppercase name":       "HTTP",
		"digits only name":     "8080x8080x8080x8080",
		"name too long":        "a-very-long-port-name",
		"leading hyphen":       "-http",
		"trailing hyphen":      "http-",
		"consecutive hyphens":  "my--port",
		"no letter":            "80-80",
		"underscore":           "my_port",
		"shell injection":      `8080"; touch pwned; echo "`,
		"command substitution": "$(touch pwned)",
		"newline injection":    "80\nRUN touch pwned",
		"ampersand":            "80 && touch pwned",
		"named port":           "http",
	}
	for name, input := range invalid {
		assert.NotNil(t, portValidator(input), "expected %q (%s) to be rejected", input, name)
	}
}

func TestDirPathValidator(t *testing.T) {
	for _, input := range []string{".", "./src", "src/my app", "資料/服務"} {
		assert.Nil(t, GetValidator("dirPath")(input), "expected %q to be valid", input)
	}

	for _, input := range []string{"", ".; touch /tmp/pwned #", "$(touch pwned)", "src\nother", "-rf", "--file=attacker/context"} {
		assert.NotNil(t, GetValidator("dirPath")(input), "expected %q to be rejected", input)
	}
}

func TestRepositoryBranchValidator(t *testing.T) {
	for _, input := range []string{"main", "feature/security-fix", "release-1.2"} {
		assert.Nil(t, GetValidator("repositoryBranch")(input), "expected %q to be valid", input)
	}

	for _, input := range []string{"", "main, attacker", "[main, attacker]", "feature branch", "../main", "main..next", "main#comment", "!main", "release+"} {
		assert.NotNil(t, GetValidator("repositoryBranch")(input), "expected %q to be rejected", input)
	}
}

func TestPythonEntryPointValidator(t *testing.T) {
	for _, input := range []string{"app.py", "./app.py", "./src/app.py", "src/my app.py", "資料/服務.py", "pkg/-internal/app.py"} {
		assert.Nil(t, GetValidator("pythonEntryPoint")(input), "expected %q to be valid", input)
	}

	for _, input := range []string{"", "/app.py", "../app.py", "src/../app.py", "-c", "app.py\nRUN touch pwned"} {
		assert.NotNil(t, GetValidator("pythonEntryPoint")(input), "expected %q to be rejected", input)
	}
}

func TestCINameValidatorsRejectShellSyntax(t *testing.T) {
	tests := map[string][]string{
		"azureContainerRegistry": {"registry;command", "registry$(command)", "registry --option value"},
		"azureManagedCluster":    {"cluster;command", "cluster$(command)", "cluster --option value"},
		"azureResourceGroup":     {"group;command", "group$(command)", "group --option value", "production west"},
		"containerImageName":     {"image;command", "image$(command)", "image --option value"},
		"kubernetesNamespace":    {"default;command", "default$(command)"},
	}
	for kind, inputs := range tests {
		for _, input := range inputs {
			assert.NotNil(t, GetValidator(kind)(input), "expected kind %s to reject %q", kind, input)
		}
	}

	assert.Nil(t, GetValidator("azureResourceGroup")("production-west"))
	assert.Nil(t, GetValidator("azureContainerRegistry")("myacr.azurecr.io"))
	assert.Nil(t, GetValidator("azureManagedCluster")("cluster-1"))
	assert.Nil(t, GetValidator("containerImageName")("team/my_app"))
	assert.Nil(t, GetValidator("kubernetesNamespace")("team-a"))
}

func TestKubernetesResourceNameValidator(t *testing.T) {
	for _, input := range []string{"app", "my-app", "app2"} {
		assert.Nil(t, GetValidator("kubernetesResourceName")(input), "expected %q to be valid", input)
	}
	for _, input := range []string{"App", "app_name", `app" }}`, strings.Repeat("a", 64)} {
		assert.NotNil(t, GetValidator("kubernetesResourceName")(input), "expected %q to be rejected", input)
	}
}

func TestHelmChartOverridesValidator(t *testing.T) {
	for _, input := range []string{"replicas=2", "image.tag=latest,replicas=2", "feature.enabled=true"} {
		assert.Nil(t, GetValidator("helmChartOverrides")(input), "expected %q to be valid", input)
	}
	for _, input := range []string{"replicas:2", "replicas=2;command", "replicas=$(command)", "replicas=two words"} {
		assert.NotNil(t, GetValidator("helmChartOverrides")(input), "expected %q to be rejected", input)
	}
}

// TestContainerImageVersionValidator covers the VERSION value interpolated into
// Dockerfile FROM instructions. The valid cases are the defaults and example
// values actually shipped in the templates.
func TestContainerImageVersionValidator(t *testing.T) {
	valid := []string{
		"3.13", "3.9", "latest", "8-jdk-alpine", "21-jre", "jdk21",
		"27.0-alpine", "7.1-apache", "3.1.2", "1.88.0", "5.0", "1_0",
	}
	for _, input := range valid {
		assert.Nil(t, containerImageVersionValidator(input), "expected %q to be valid", input)
	}

	invalid := map[string]string{
		"empty":             "",
		"leading dot":       ".3.13",
		"leading dash":      "-3.13",
		"slash":             "library/python",
		"colon":             "3.13:latest",
		"space":             "3.13 AS builder",
		"quote":             `3.13"`,
		"newline injection": "3.13\nRUN touch pwned",
		"command sub":       "$(touch pwned)",
		"too long":          strings.Repeat("1", maxImageTagLength+1),
	}
	for name, input := range invalid {
		assert.NotNil(t, containerImageVersionValidator(input), "expected %q (%s) to be rejected", input, name)
	}
}

// TestGetValidatorWiring asserts the kinds used by the vulnerable templates
// resolve to a real validator rather than falling through to the baseline.
// The payloads deliberately contain no control characters, so these assertions
// fail if a kind is unwired, rather than being rescued by defaultValidator.
func TestGetValidatorWiring(t *testing.T) {
	assert.NotNil(t, GetValidator("fileName")(`my"app.py`))
	assert.Nil(t, GetValidator("fileName")("app.py"))

	assert.NotNil(t, GetValidator("filePath")(`./Docker"file`))
	assert.Nil(t, GetValidator("filePath")("./Dockerfile"))

	assert.NotNil(t, GetValidator("port")(`8080"; touch pwned; echo "`))
	assert.Nil(t, GetValidator("port")("80"))

	assert.NotNil(t, GetValidator("containerImageVersion")("3.13 AS builder"))
	assert.Nil(t, GetValidator("containerImageVersion")("3.13"))

	assert.NotNil(t, GetValidator("envVarMap")(`{"A":"x\n  INJECTED: pwned"}`))
	assert.Nil(t, GetValidator("envVarMap")(`{"A":"x"}`))
}

// TestFilePathValidatorBlocksAdvisoryPayloads asserts that the ENTRYPOINT kind,
// which must stay permissive enough for nested paths such as "src/app.py",
// still rejects every payload from the Dockerfile instruction injection
// advisory.
func TestFilePathValidatorBlocksAdvisoryPayloads(t *testing.T) {
	payloads := []string{
		`my"app.py`,
		`my\app.py`,
		"my\napp.py",
		`main.py", "extra.py`,
		"app.py\"]\nRUN touch pwned\n# .py",
		"my app.py",
		"$(touch pwned).py",
	}
	for _, payload := range payloads {
		assert.NotNil(t, filePathValidator(payload), "expected %q to be rejected", payload)
	}

	// ...while still accepting the layouts the python extractor produces.
	for _, valid := range []string{"app.py", "src/app.py", "./src/server.py", "a/b/c/main.py"} {
		assert.Nil(t, filePathValidator(valid), "expected %q to be valid", valid)
	}
}

// TestKeyValueMapValidatorRejectsInjection covers ENVVARS, whose parsed keys and
// values are rendered as raw YAML scalars in the generated ConfigMap.
func TestKeyValueMapValidatorRejectsInjection(t *testing.T) {
	assert.Nil(t, keyValueMapValidator(`{"KEY": "value"}`))
	assert.Nil(t, keyValueMapValidator(`{"A": "b", "C": "d"}`))

	// A JSON \n escape survives parsing and would inject a new ConfigMap key.
	assert.NotNil(t, keyValueMapValidator(`{"A":"x\n  INJECTED: pwned"}`))
	assert.NotNil(t, keyValueMapValidator(`{"A\nINJECTED": "x"}`))
	assert.NotNil(t, keyValueMapValidator(`{"A":"x\u2028y"}`))
	assert.NotNil(t, keyValueMapValidator(`{"A":"x\u0000y"}`))
	assert.NotNil(t, keyValueMapValidator(`{"key": "value"`))
}

func TestRejectUnsafeBytesRejectsFormatCharacters(t *testing.T) {
	// Bidirectional overrides and zero-width characters make a generated file
	// render differently from how it is interpreted.
	assert.NotNil(t, rejectUnsafeBytes("evil\u202ereversed"))
	assert.NotNil(t, rejectUnsafeBytes("zero\u200bwidth"))
	assert.NotNil(t, rejectUnsafeBytes("\ufeffbom"))
	assert.Nil(t, rejectUnsafeBytes("ordinary value"))
}

func TestFileNameValidatorRejectsLeadingDash(t *testing.T) {
	// A leading dash is parsed as an option by the command the file name is
	// passed to, turning CMD ["-mpkg"] into "python -m pkg".
	assert.NotNil(t, fileNameValidator("-mpkg.mod"))
	assert.NotNil(t, fileNameValidator("-c"))
	assert.NotNil(t, filePathValidator("src/-mpkg.mod"))
	assert.Nil(t, fileNameValidator("my-app.py"))
}

// TestValidatorBoundaries covers the exact length limits.
func TestValidatorBoundaries(t *testing.T) {
	assert.Nil(t, fileNameValidator(strings.Repeat("a", maxFileNameLength)))
	assert.NotNil(t, fileNameValidator(strings.Repeat("a", maxFileNameLength+1)))

	assert.Nil(t, containerImageVersionValidator(strings.Repeat("1", maxImageTagLength)))
	assert.NotNil(t, containerImageVersionValidator(strings.Repeat("1", maxImageTagLength+1)))

	assert.Nil(t, portValidator("1"))
	assert.Nil(t, portValidator("65535"))
	assert.NotNil(t, portValidator("0"))
	assert.NotNil(t, portValidator("65536"))
}

func TestKubernetesProbeTypeValidator(t *testing.T) {
	assert.Nil(t, kubernetesProbeTypeValidator("httpGet"))
	assert.Nil(t, kubernetesProbeTypeValidator("tcpSocket"))
	assert.NotNil(t, kubernetesProbeTypeValidator("exec"))
}

func TestKeyValueMapValidator(t *testing.T) {
	assert.Nil(t, keyValueMapValidator(`{"key": "value"}`))
	assert.NotNil(t, keyValueMapValidator(`{"key": "value"`))
}

func TestScalingResourceTypeValidator(t *testing.T) {
	assert.Nil(t, scalingResourceTypeValidator("cpu"))
	assert.Nil(t, scalingResourceTypeValidator("memory"))
	assert.NotNil(t, scalingResourceTypeValidator("disk"))
}

func TestImagePullPolicyValidator(t *testing.T) {
	assert.Nil(t, imagePullPolicyValidator("Always"))
	assert.Nil(t, imagePullPolicyValidator("IfNotPresent"))
	assert.Nil(t, imagePullPolicyValidator("Never"))
	assert.NotNil(t, imagePullPolicyValidator("Sometimes"))
}
