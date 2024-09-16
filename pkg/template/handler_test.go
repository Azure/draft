package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type FakeTemplateWriter struct {
	filesWritten map[string][]byte
}

func (m *FakeTemplateWriter) EnsureDirectory(dirPath string) error {
	return nil
}

func (m *FakeTemplateWriter) WriteFile(filePath string, content []byte) error {
	if m.filesWritten == nil {
		m.filesWritten = make(map[string][]byte)
	}
	m.filesWritten[filePath] = content
	return nil
}

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		templateName string
		destination  string
		tmplWriter   *FakeTemplateWriter
		expectedErr  bool
	}{
		{
			name:         "successful 0.0.1 copy",
			version:      "0.0.1",
			templateName: "Deployment",
			destination:  "./manifests",
			tmplWriter:   &FakeTemplateWriter{},
			expectedErr:  false,
		},
		{
			name:         "invalid template name",
			version:      "0.0.1",
			templateName: "InvalidTemplate",
			destination:  "./manifests",
			tmplWriter:   &FakeTemplateWriter{},
			expectedErr:  true,
		},
		{
			name:         "invalid template version",
			version:      "1.0.0",
			templateName: "Deployment",
			destination:  "./manifests",
			tmplWriter:   &FakeTemplateWriter{},
			expectedErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl, err := GetTemplate(tt.templateName, tt.version, tt.destination, tt.tmplWriter)
			assert.Equal(t, tt.expectedErr, err != nil)
			if !tt.expectedErr {
				assert.Equal(t, tmpl.dest, tt.destination)
				assert.Equal(t, tmpl.version, tt.version)
				assert.Equal(t, tmpl.Config.TemplateName, tt.templateName)
			}
		})
	}

}

func TestSuccessfulVersions(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		cpuIn          string
		memIn          string
		tmplWriter     *FakeTemplateWriter
		expectedOutput string
	}{
		{
			name:           "successful 0.0.1 copy",
			version:        "0.0.1",
			tmplWriter:     &FakeTemplateWriter{},
			expectedOutput: "apiVersion: apps/v1 \nkind: Deployment\nmetadata:\n name: test\n namespace: test\n labels:\n  app.kubernetes.io/name: test\n  app.kubernetes.io/part-of: test\n  kubernetes.azure.com/generator: draft\nspec:\n replicas: 2\n selector:\n  matchLabels:\n   app: test\n template:\n  metadata:\n   labels:\n    app: test\n  spec:\n   containers:\n    - name: test\n      image: test:latest\n      imagePullPolicy: IfNotPresent\n      ports:\n      - containerPort: 80\n      securityContext:\n        privileged: false\n        allowPrivilegeEscalation: false\n        runAsUser: 1000\n        runAsGroup: 3000\n        runAsNonRoot: true\n        readOnlyRootFilesystem: true\n        capabilities:\n        drop:\n          - all\n        add:\n          - SETPCAP\n          - MKNOD\n          - AUDIT_WRITE\n          - CHOWN\n          - DAC_OVERRIDE\n          - FOWNER\n          - FSETID\n          - KILL\n          - SETGID\n          - SETUID\n          - NET_BIND_SERVICE\n          - SYS_CHROOT\n          - SETFCAP\n          - SYS_PTRACE\n  affinity:\n    podAntiAffinity:\n      preferredDuringSchedulingIgnoredDuringExecution:\n      - weight: 100\n        podAffinityTerm:\n          topologyKey: kubernetes.io/hostname\n          labelSelector:\n            matchLabels:\n            app: test\n    topologySpreadConstraints:\n      - maxSkew: 1\n        topologyKey: topology.kubernetes.io/zone\n        whenUnsatisfiable: ScheduleAnyway\n        labelSelector:\n          matchLabels:\n          app: test\n    hostNetwork: false\n    hostIPC: false\n    securityContext:\n      seccompProfile:\n      type: RuntimeDefault",
		},
		{
			name:           "successful 0.0.2 copy",
			version:        "0.0.2",
			cpuIn:          "1",
			memIn:          "1Gi",
			tmplWriter:     &FakeTemplateWriter{},
			expectedOutput: "apiVersion: apps/v1 \nkind: Deployment\nmetadata:\n name: test\n namespace: test\n labels:\n  app.kubernetes.io/name: test\n  app.kubernetes.io/part-of: test\n  kubernetes.azure.com/generator: draft\nspec:\n replicas: 2\n selector:\n  matchLabels:\n   app: test\n template:\n  metadata:\n   labels:\n    app: test\n  spec:\n   containers:\n    - name: test\n      image: test:latest\n      imagePullPolicy: IfNotPresent\n      ports:\n      - containerPort: 80\n      resources:\n        limits:\n          cpu: 1\n          memory: 1Gi\n        requests:\n          cpu: 1\n          memory: 1Gi\n      securityContext:\n        privileged: false\n        allowPrivilegeEscalation: false\n        runAsUser: 1000\n        runAsGroup: 3000\n        runAsNonRoot: true\n        readOnlyRootFilesystem: true\n        capabilities:\n        drop:\n          - all\n        add:\n          - SETPCAP\n          - MKNOD\n          - AUDIT_WRITE\n          - CHOWN\n          - DAC_OVERRIDE\n          - FOWNER\n          - FSETID\n          - KILL\n          - SETGID\n          - SETUID\n          - NET_BIND_SERVICE\n          - SYS_CHROOT\n          - SETFCAP\n          - SYS_PTRACE\n  affinity:\n    podAntiAffinity:\n      preferredDuringSchedulingIgnoredDuringExecution:\n      - weight: 100\n        podAffinityTerm:\n          topologyKey: kubernetes.io/hostname\n          labelSelector:\n            matchLabels:\n            app: test\n    topologySpreadConstraints:\n      - maxSkew: 1\n        topologyKey: topology.kubernetes.io/zone\n        whenUnsatisfiable: ScheduleAnyway\n        labelSelector:\n          matchLabels:\n          app: test\n    hostNetwork: false\n    hostIPC: false\n    securityContext:\n      seccompProfile:\n      type: RuntimeDefault",
		},
		{
			name:           "successful 0.0.3 copy",
			version:        "0.0.3",
			cpuIn:          "1",
			memIn:          "1Gi",
			tmplWriter:     &FakeTemplateWriter{},
			expectedOutput: "apiVersion: apps/v1 \nkind: Deployment\nmetadata:\n name: test\n namespace: test\n labels:\n  app.kubernetes.io/name: test\n  app.kubernetes.io/part-of: test\n  kubernetes.azure.com/generator: draft\nspec:\n replicas: 2\n selector:\n  matchLabels:\n   app: test\n template:\n  metadata:\n   labels:\n    app: test\n  spec:\n   containers:\n    - name: test\n      image: test:latest\n      imagePullPolicy: IfNotPresent\n      ports:\n      - containerPort: 80\n      resources:\n        limits:\n          cpu: 1\n          memory: 1Gi\n        requests:\n          cpu: 1\n          memory: 1Gi\n      livenessProbe:\n        tcpSocket:\n          port: 80\n      readinessProbe:\n        tcpSocket:\n          port: 80\n        periodSeconds: 1\n        timeoutSeconds: 5\n      securityContext:\n        privileged: false\n        allowPrivilegeEscalation: false\n        runAsUser: 1000\n        runAsGroup: 3000\n        runAsNonRoot: true\n        readOnlyRootFilesystem: true\n        capabilities:\n        drop:\n          - all\n        add:\n          - SETPCAP\n          - MKNOD\n          - AUDIT_WRITE\n          - CHOWN\n          - DAC_OVERRIDE\n          - FOWNER\n          - FSETID\n          - KILL\n          - SETGID\n          - SETUID\n          - NET_BIND_SERVICE\n          - SYS_CHROOT\n          - SETFCAP\n          - SYS_PTRACE\n  affinity:\n    podAntiAffinity:\n      preferredDuringSchedulingIgnoredDuringExecution:\n      - weight: 100\n        podAffinityTerm:\n          topologyKey: kubernetes.io/hostname\n          labelSelector:\n            matchLabels:\n            app: test\n    topologySpreadConstraints:\n      - maxSkew: 1\n        topologyKey: topology.kubernetes.io/zone\n        whenUnsatisfiable: ScheduleAnyway\n        labelSelector:\n          matchLabels:\n          app: test\n    hostNetwork: false\n    hostIPC: false\n    securityContext:\n      seccompProfile:\n      type: RuntimeDefault",
		},
		{
			name:           "successful 0.0.4 copy",
			version:        "0.0.4",
			cpuIn:          "1",
			memIn:          "1Gi",
			tmplWriter:     &FakeTemplateWriter{},
			expectedOutput: "apiVersion: apps/v1 \nkind: Deployment\nmetadata:\n name: test\n namespace: test\n labels:\n  app.kubernetes.io/name: test\n  app.kubernetes.io/part-of: test\n  kubernetes.azure.com/generator: draft\nspec:\n replicas: 2\n selector:\n  matchLabels:\n   app: test\n template:\n  metadata:\n   labels:\n    app: test\n  spec:\n   containers:\n    - name: test\n      image: test:latest\n      imagePullPolicy: IfNotPresent\n      ports:\n      - containerPort: 80\n      resources:\n        limits:\n          cpu: 1\n          memory: 1Gi\n        requests:\n          cpu: 1\n          memory: 1Gi\n      livenessProbe:\n        tcpSocket:\n          port: 80\n      readinessProbe:\n        tcpSocket:\n          port: 80\n        periodSeconds: 1\n        timeoutSeconds: 5\n        initialDelaySeconds: 3\n        successThreshold: 1 \n        failureThreshold: 1\n      securityContext:\n        privileged: false\n        allowPrivilegeEscalation: false\n        runAsUser: 1000\n        runAsGroup: 3000\n        runAsNonRoot: true\n        readOnlyRootFilesystem: true\n        capabilities:\n        drop:\n          - all\n        add:\n          - SETPCAP\n          - MKNOD\n          - AUDIT_WRITE\n          - CHOWN\n          - DAC_OVERRIDE\n          - FOWNER\n          - FSETID\n          - KILL\n          - SETGID\n          - SETUID\n          - NET_BIND_SERVICE\n          - SYS_CHROOT\n          - SETFCAP\n          - SYS_PTRACE\n  affinity:\n    podAntiAffinity:\n      preferredDuringSchedulingIgnoredDuringExecution:\n      - weight: 100\n        podAffinityTerm:\n          topologyKey: kubernetes.io/hostname\n          labelSelector:\n            matchLabels:\n            app: test\n    topologySpreadConstraints:\n      - maxSkew: 1\n        topologyKey: topology.kubernetes.io/zone\n        whenUnsatisfiable: ScheduleAnyway\n        labelSelector:\n          matchLabels:\n          app: test\n    hostNetwork: false\n    hostIPC: false\n    securityContext:\n      seccompProfile:\n      type: RuntimeDefault",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deploymentTemplate, err := GetTemplate("Deployment", tt.version, "./manifests", tt.tmplWriter)
			assert.NoError(t, err)

			deploymentTemplate.Config.SetVariable("APPNAME", "test")
			deploymentTemplate.Config.SetVariable("NAMESPACE", "test")
			deploymentTemplate.Config.SetVariable("CPULIMIT", tt.cpuIn)
			deploymentTemplate.Config.SetVariable("MEMLIMIT", tt.memIn)
			deploymentTemplate.Config.SetVariable("CPUREQ", tt.cpuIn)
			deploymentTemplate.Config.SetVariable("MEMREQ", tt.memIn)

			err = deploymentTemplate.CreateTemplates()
			assert.NoError(t, err)

			assert.Equal(t, string(tt.tmplWriter.filesWritten["./manifests/deployment.yaml"]), tt.expectedOutput)
		})
	}
}
