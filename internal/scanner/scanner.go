package scanner

import (
	"context"
	"encoding/json"
	"os"

	"github.com/vahan90/skywoker/internal/logger"
	v1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	core "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpaApiv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpaApi "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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

func getDeployments(namespace string, clientset *kubernetes.Clientset) (*v1.DeploymentList, error) {
	deploymentsClient := clientset.AppsV1().Deployments(namespace)
	list, err := deploymentsClient.List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, d := range list.Items {
		logger.Debugf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}

	return list, err
}

func getPodsForDeployment(namespace string, clientset *kubernetes.Clientset, matchLabels map[string]string) (*core.PodList, error) {
	podsClient := clientset.CoreV1().Pods(namespace)

	list, err := podsClient.List(context.Background(), metav1.ListOptions{LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: matchLabels})})

	if err != nil {
		return nil, err
	}

	for _, d := range list.Items {
		logger.Debugf(" * %s\n", d.Name)
	}

	return list, err
}

func getHPAs(namespace string, clientset *kubernetes.Clientset) (*v2.HorizontalPodAutoscalerList, error) {
	hpaClient := clientset.AutoscalingV2().HorizontalPodAutoscalers(namespace)
	list, err := hpaClient.List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, d := range list.Items {
		logger.Debugf(" * %s\n", d.Name)
	}

	return list, err
}

func getPDBs(namespace string, clientset *kubernetes.Clientset) (*policy.PodDisruptionBudgetList, error) {
	pdbClient := clientset.PolicyV1().PodDisruptionBudgets(namespace)
	list, err := pdbClient.List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, d := range list.Items {
		logger.Debugf(" * %s\n", d.Name)
	}

	return list, err
}

func getVPAs(namespace string, clientset *vpaApi.Clientset) (*vpaApiv1.VerticalPodAutoscalerList, error) {
	vpaClient := clientset.AutoscalingV1().VerticalPodAutoscalers(namespace)
	list, err := vpaClient.List(context.Background(), metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, d := range list.Items {
		logger.Debugf(" * %s\n", d.Name)
	}

	return list, err
}

func processDeployment(d v1.Deployment, c *kubernetes.Clientset, v *vpaApi.Clientset) {
	logger.Debugf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)

	// loop to get container information
	var containers []Container
	for _, container := range d.Spec.Template.Spec.Containers {
		var c Container
		c.Name = container.Name
		c.Resource.Requests.Memory = container.Resources.Requests.Memory().String()
		c.Resource.Requests.CPU = container.Resources.Requests.Cpu().String()
		c.Resource.Limits.Memory = container.Resources.Limits.Memory().String()
		c.Resource.Limits.CPU = container.Resources.Limits.Cpu().String()

		// loop to get readiness and liveness probe information
		readinessProbeSet := false
		livenessProbeSet := false
		if container.ReadinessProbe != nil {
			readinessProbeSet = true
		}
		if container.LivenessProbe != nil {
			livenessProbeSet = true
		}

		c.ReadinessProbeSet = readinessProbeSet
		c.LivenessProbeSet = livenessProbeSet
		containers = append(containers, c)
	}

	pods, err := getPodsForDeployment(d.Namespace, c, d.ObjectMeta.Labels)

	if err != nil {
		logger.Errorf(err.Error())
	}

	qos := "0 pods active for this workload"

	if len(pods.Items) > 0 {
		firstPod := pods.Items[0]
		qos = string(firstPod.Status.QOSClass)
	}

	pod := Pod{
		PodLabels:  d.Spec.Template.ObjectMeta.Labels,
		Containers: containers,
		QoS:        qos,
	}

	pdbs, err := getPDBs(d.Namespace, c)

	if err != nil {
		logger.Errorf(err.Error())
	}

	var matchingPDBs []string
	for _, pdb := range pdbs.Items {
		// Check if the PDB matches the workload selector
		if labelsMatchAny(pdb.Spec.Selector, d.ObjectMeta.Labels) {
			matchingPDBs = append(matchingPDBs, pdb.Name)
		}
	}

	PDBSet := false

	if len(matchingPDBs) > 0 {
		PDBSet = true
	}

	hpaSet := false
	hpas, err := getHPAs(d.Namespace, c)

	if err != nil {
		logger.Errorf(err.Error())
	}

	var matchingHPAs []string
	for _, hpa := range hpas.Items {
		if hpa.Spec.ScaleTargetRef.Name == d.Name {
			matchingHPAs = append(matchingHPAs, hpa.Name)
		}
	}

	if len(matchingHPAs) > 0 {
		hpaSet = true
	}

	vpaSet := false
	vpas, err := getVPAs(d.Namespace, v)

	if err != nil {
		logger.Errorf(err.Error())
	}

	var matchingVPAs []string
	for _, vpa := range vpas.Items {
		if vpa.Spec.TargetRef.Name == d.Name {
			matchingVPAs = append(matchingVPAs, vpa.Name)
		}
	}

	if len(matchingVPAs) > 0 {
		vpaSet = true
	}

	deployment := Deployment{
		Replicas: *d.Spec.Replicas,
		HPASet:   hpaSet,
		VPASet:   vpaSet,
		PDBSet:   PDBSet,
		Pod:      pod,
	}

	// get results
	results := Workload{
		Type:       "Deployment",
		Labels:     d.ObjectMeta.Labels,
		Name:       d.Name,
		Deployment: &deployment,
	}

	logger.Infof("%#v\n", results)
	b, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		logger.Warningf("error:", err)
	}
	logger.Infof(string(b))
}

func ScanCluster(namespace string, workloadType string) {
	kubeconfig := os.Getenv("KUBECONFIG")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		logger.Errorf(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		logger.Errorf(err.Error())
	}

	vpaClientset, err := vpaApi.NewForConfig(config)

	if err != nil {
		logger.Errorf(err.Error())
	}

	deploymentList, err := getDeployments(namespace, clientset)

	if err != nil {
		logger.Errorf(err.Error())
	}

	for _, d := range deploymentList.Items {
		processDeployment(d, clientset, vpaClientset)
	}
}

// labelsMatchAny checks if a selector matches any of the provided labels.
func labelsMatchAny(selector *metav1.LabelSelector, labels map[string]string) bool {
	if selector == nil {
		return false
	}
	for key, value := range labels {
		if selector.MatchLabels[key] == value {
			return true
		}
	}
	return false
}
