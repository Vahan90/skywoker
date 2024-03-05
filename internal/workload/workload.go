package workload

type Requests struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

type Limits struct {
	Memory string `json:"memory"`
	CPU    string `json:"cpu"`
}

type Resource struct {
	Requests Requests `json:"requests"`
	Limits   Limits   `json:"limits"`
}

type Container struct {
	Name              string `json:"name"`
	ReadinessProbeSet bool   `json:"readinessProbeSet"`
	LivenessProbeSet  bool   `json:"livenessProbeSet"`
	Resource          `json:"resource"`
}

type Pod struct {
	PodLabels  map[string]string `json:"labels"`
	Containers []Container       `json:"containers"`
	QoS        string            `json:"qos"`
}

type Cronjob struct {
	Schedule                   string `json:"schedule"`
	ConcurrencyPolicy          string `json:"concurrencyPolicy"`
	Suspended                  bool   `json:"suspended"`
	SuccessfulJobsHistoryLimit int32  `json:"successfulJobsHistoryLimit"`
	FailedJobsHistoryLimit     int32  `json:"failedJobsHistoryLimit"`
	Parallelism                int32  `json:"parallelism"`
	Completions                int32  `json:"completions"`
	RestartPolicy              string `json:"restartPolicy"`
	BackoffLimit               int32  `json:"backoffLimit"`
	ActiveDeadlineSeconds      int64  `json:"activeDeadlineSeconds"`
	Pod                        `json:"pod"`
}

type Statefulset struct {
	Replicas int32 `json:"replicas"`
	HPASet   bool  `json:"hpaSet"`
	VPASet   bool  `json:"vpaSet"`
	PDBSet   bool  `json:"pdbSet"`
	Pod      `json:"pod"`
}

type Deployment struct {
	Replicas int32 `json:"replicas"`
	HPASet   bool  `json:"hpaSet"`
	VPASet   bool  `json:"vpaSet"`
	PDBSet   bool  `json:"pdbSet"`
	Pod      `json:"pod"`
}

type Workload struct {
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Labels      map[string]string `json:"labels"`
	Deployment  *Deployment       `json:"deployment,omitempty"`
	Statefulset *Statefulset      `json:"statefulset,omitempty"`
	Cronjob     *Cronjob          `json:"cronjob,omitempty"`
}
