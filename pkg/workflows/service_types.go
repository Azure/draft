package workflows

import (
	"errors"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8s.io/client-go/kubernetes/scheme"
)

type ServiceManifest interface {
	LoadFromFile(string) error
	WriteToFile(string) error
	SetAnnotations(map[string]string)
	SetServiceType(string)
	GetServiceName() string
}

type HelmProductionYaml struct {
	Image   image   `yaml:"image"`
	Service service `yaml:"service"`
}

type service struct {
	Annotations map[string]string `yaml:"annotations"`
	ServiceType string            `yaml:"type"`
	Port        string            `yaml:"port"`
}

type image struct {
	Repository string `yaml:"repository"`
	PullPolicy string `yaml:"pullPolicy"`
	Tag        string `yaml:"tag"`
}

func (hpy *HelmProductionYaml) SetAnnotations(annotations map[string]string) {
	hpy.Service.Annotations = annotations
}

func (hpy *HelmProductionYaml) SetServiceType(serviceType string) {
	hpy.Service.ServiceType = serviceType
}

func (hpy *HelmProductionYaml) GetServiceName() string {
	return `{{ include "{{APPNAME}}.fullname" . }}`
}

func (hpy *HelmProductionYaml) LoadFromFile(filePath string) error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(file, &hpy)
}

func (hpy *HelmProductionYaml) WriteToFile(filePath string) error {
	currYaml, err := yaml.Marshal(hpy)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, currYaml, 0755)
}

type ServiceYaml struct {
	*corev1.Service
}

func (sy *ServiceYaml) SetAnnotations(annotations map[string]string) {
	sy.Annotations = annotations
}

func (sy *ServiceYaml) SetServiceType(serviceType string) {
	sy.Spec.Type = corev1.ServiceType(serviceType)
}

func (sy *ServiceYaml) GetServiceName() string {
	return sy.Name
}

func (sy *ServiceYaml) LoadFromFile(filePath string) error {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	file, err := ioutil.ReadFile(filePath)
	k8sObj, _, err := decode(file, nil, nil)
	if err != nil {
		return err
	}

	svc, ok := k8sObj.(*corev1.Service)
	if !ok {
		return errors.New("could not load file into ServiceYaml")
	}

	sy.Service = svc

	return nil
}

func (hpy *ServiceYaml) WriteToFile(filePath string) error {
	printer := printers.YAMLPrinter{}

	out, err := os.OpenFile(filePath, os.O_RDWR, 0755)
	if err != nil {
		return nil
	}
	defer out.Close()

	return printer.PrintObj(hpy.Service, out)
}
