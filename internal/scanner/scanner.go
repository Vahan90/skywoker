package scanner

import (
	"context"
	"encoding/json"
	"os"

	"github.com/vahan90/skywoker/internal/logger"
	"github.com/vahan90/skywoker/internal/reporter"
	"github.com/vahan90/skywoker/internal/workload"
	v1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	policy "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	vpaApiv1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpaApi "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func calculateQos(d v1.Deployment) string {
	containers := d.Spec.Template.Spec.Containers
	for _, container := range containers {
		// Check if either memory or CPU requests or limits are not set
		if (container.Resources.Requests.Memory().IsZero() || container.Resources.Requests.Cpu().IsZero()) ||
			(container.Resources.Limits.Memory().IsZero() || container.Resources.Limits.Cpu().IsZero()) {
			return "Burstable"
		}

		// Check if both memory and CPU requests and limits are set
		if !container.Resources.Requests.Memory().IsZero() && !container.Resources.Requests.Cpu().IsZero() &&
			!container.Resources.Limits.Memory().IsZero() && !container.Resources.Limits.Cpu().IsZero() {
			return "Guaranteed"
		}
	}

	// If no containers match any of the above conditions, return BestEffort
	return "BestEffort"
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

func processDeployment(d v1.Deployment, c *kubernetes.Clientset, v *vpaApi.Clientset) workload.Workload {
	logger.Debugf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)

	// loop to get container information
	var containers []workload.Container
	for _, container := range d.Spec.Template.Spec.Containers {
		var c workload.Container
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

	qos := calculateQos(d)

	pod := workload.Pod{
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

	deployment := workload.Deployment{
		Replicas: *d.Spec.Replicas,
		HPASet:   hpaSet,
		VPASet:   vpaSet,
		PDBSet:   PDBSet,
		Pod:      pod,
	}

	// get results
	results := workload.Workload{
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

	return results
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
		w := processDeployment(d, clientset, vpaClientset)
		reporter.GenerateDeploymentReport(w)
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
