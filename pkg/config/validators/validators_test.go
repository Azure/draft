package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidator(t *testing.T) {
	assert.NotNil(t, GetValidator("NonExistentKind"))
}

func TestDefaultValidator(t *testing.T) {
	assert.Nil(t, defaultValidator("test"))
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
