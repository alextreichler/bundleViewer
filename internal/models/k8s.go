package models

import "time"

type K8sList struct {
	Kind  string        `json:"kind"`
	Items []K8sResource `json:"items"`
}

type K8sResource struct {
	Metadata       K8sMetadata       `json:"metadata"`
	Status         K8sStatus         `json:"status,omitempty"`
	Spec           K8sSpec           `json:"spec,omitempty"`
	InvolvedObject K8sInvolvedObject `json:"involvedObject,omitempty"`
	Reason         string            `json:"reason,omitempty"`
	Message        string            `json:"message,omitempty"`
	Source         K8sSource         `json:"source,omitempty"`
	FirstTimestamp time.Time         `json:"firstTimestamp,omitempty"`
	LastTimestamp  time.Time         `json:"lastTimestamp,omitempty"`
	Count          int               `json:"count,omitempty"`
	Type           string            `json:"type,omitempty"`
	Data           map[string]string `json:"data,omitempty"`
}

type K8sSpec struct {
	NodeName         string                 `json:"nodeName,omitempty"`
	Containers       []K8sContainer         `json:"containers,omitempty"`
	StorageClassName string                 `json:"storageClassName,omitempty"`
	AccessModes      []string               `json:"accessModes,omitempty"`
	Resources        K8sResourceRequirements `json:"resources,omitempty"`
	Type             string                 `json:"type,omitempty"`      // Service type
	ClusterIP        string                 `json:"clusterIP,omitempty"` // Service IP
	Ports            []K8sServicePort       `json:"ports,omitempty"`     // Service ports
	Replicas         int                    `json:"replicas,omitempty"`  // Controller replicas
}

type K8sContainer struct {
	Name           string                  `json:"name"`
	Image          string                  `json:"image"`
	Resources      K8sResourceRequirements `json:"resources,omitempty"`
	LivenessProbe  *K8sProbe               `json:"livenessProbe,omitempty"`
	ReadinessProbe *K8sProbe               `json:"readinessProbe,omitempty"`
	StartupProbe   *K8sProbe               `json:"startupProbe,omitempty"`
}

type K8sProbe struct {
	InitialDelaySeconds int            `json:"initialDelaySeconds,omitempty"`
	TimeoutSeconds      int            `json:"timeoutSeconds,omitempty"`
	PeriodSeconds       int            `json:"periodSeconds,omitempty"`
	SuccessThreshold    int            `json:"successThreshold,omitempty"`
	FailureThreshold    int            `json:"failureThreshold,omitempty"`
	HTTPGet             *K8sHTTPGet    `json:"httpGet,omitempty"`
	Exec                *K8sExecAction `json:"exec,omitempty"`
	TCPSocket           *K8sTCPSocket  `json:"tcpSocket,omitempty"`
}

type K8sHTTPGet struct {
	Path string      `json:"path,omitempty"`
	Port interface{} `json:"port"`
}

type K8sExecAction struct {
	Command []string `json:"command,omitempty"`
}

type K8sTCPSocket struct {
	Port interface{} `json:"port"`
}

type K8sResourceRequirements struct {
	Limits   map[string]string `json:"limits,omitempty"`
	Requests map[string]string `json:"requests,omitempty"`
}

type K8sServicePort struct {
	Name       string      `json:"name,omitempty"`
	Port       int         `json:"port"`
	TargetPort interface{} `json:"targetPort,omitempty"`
	Protocol   string      `json:"protocol,omitempty"`
}

type K8sStatus struct {
	Phase             string               `json:"phase,omitempty"`
	PodIP             string               `json:"podIP,omitempty"`
	ContainerStatuses []K8sContainerStatus `json:"containerStatuses,omitempty"`
	Conditions        []K8sCondition       `json:"conditions,omitempty"`
	Capacity          map[string]string    `json:"capacity,omitempty"`    // PVC/Node Capacity
	Allocatable       map[string]string    `json:"allocatable,omitempty"` // Node Allocatable
	AccessModes       []string             `json:"accessModes,omitempty"`
	Replicas          int                  `json:"replicas,omitempty"`      // Controller current replicas
	ReadyReplicas     int                  `json:"readyReplicas,omitempty"` // Controller ready replicas
}

type K8sMetadata struct {
	Name              string               `json:"name"`
	Namespace         string               `json:"namespace"`
	CreationTimestamp time.Time            `json:"creationTimestamp"`
	Labels            map[string]string    `json:"labels"`
	OwnerReferences   []K8sOwnerReference  `json:"ownerReferences,omitempty"`
}

type K8sOwnerReference struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type K8sContainerStatus struct {
	Name         string                    `json:"name"`
	State        map[string]K8sStateDetail `json:"state"`
	LastState    map[string]K8sStateDetail `json:"lastState"`
	Ready        bool                      `json:"ready"`
	RestartCount int                       `json:"restartCount"`
	Image        string                    `json:"image"`
}

type K8sStateDetail struct {
	Reason     string    `json:"reason,omitempty"`
	Message    string    `json:"message,omitempty"`
	StartedAt  time.Time `json:"startedAt,omitempty"`
	FinishedAt time.Time `json:"finishedAt,omitempty"`
	ExitCode   int       `json:"exitCode,omitempty"`
}

type K8sCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type K8sInvolvedObject struct {
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type K8sSource struct {
	Component string `json:"component"`
	Host      string `json:"host,omitempty"`
}

type K8sStore struct {
	Pods                   []K8sResource
	Services               []K8sResource
	ConfigMaps             []K8sResource
	Endpoints              []K8sResource
	Events                 []K8sResource
	LimitRanges            []K8sResource
	PersistentVolumeClaims []K8sResource
	ReplicationControllers []K8sResource
	ResourceQuotas         []K8sResource
	ServiceAccounts        []K8sResource
	StatefulSets           []K8sResource
	Deployments            []K8sResource
	Nodes                  []K8sResource
}
