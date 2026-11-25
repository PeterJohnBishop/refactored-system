package models

type NodeRuntimeInfo struct {
	NodeName         string
	ContainerRuntime string
	KubeletVersion   string
	OSImage          string
	OperatingSystem  string
	Architecture     string
	KernelVersion    string
}
