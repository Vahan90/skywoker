package reporter

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rodaine/table"
	"github.com/vahan90/skywoker/internal/workload"
)

func GenerateDeploymentReport(w workload.Workload) {
	c := color.New(color.FgMagenta, color.Bold, color.Underline)
	fmt.Println()
	c.Println("Report for Deployment ", w.Name)
	fmt.Println()
	c.Println("Container-specific checks")
	fmt.Println()

	headerFmt := color.New(color.FgGreen, color.Bold).SprintfFunc()
	columnFmt := color.New(color.FgYellow, color.Bold).SprintfFunc()

	containerTbl := table.New("Container Name", "Rule", "Compliance", "Reason")
	containerTbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithPadding(5)

	for _, container := range w.Deployment.Pod.Containers {
		if !container.LivenessProbeSet {
			containerTbl.AddRow(container.Name, "Liveness Probe Set", "✖️", "Liveness Probe not set, should be set. https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes")
		} else {
			containerTbl.AddRow(container.Name, "Liveness Probe Set", "✔️", "Liveness Probe set")
		}

		if !container.ReadinessProbeSet {
			containerTbl.AddRow(container.Name, "Readiness Probe Set", "✖️", "Readiness Probe not set, should be set. https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes")
		} else {
			containerTbl.AddRow(container.Name, "Readiness Probe Set", "✔️", "Readiness Probe set")
		}

		if container.Resource.Limits.CPU == "0" {
			containerTbl.AddRow(container.Name, "CPU Limits Not Set", "✔️", "CPU Limits not set.")
		} else {
			containerTbl.AddRow(container.Name, "CPU Limits Set", "✖️", "CPU Limits should not be set, see https://home.robusta.dev/blog/stop-using-cpu-limits in the Robusta blog for more information.")
		}

		if container.Resource.Limits.Memory == "0" {
			containerTbl.AddRow(container.Name, "Memory Limits Not Set", "✖️", "You should set a memory limit for your container. See https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-requests-and-limits-of-pod-and-container for more information.")
		} else {
			containerTbl.AddRow(container.Name, "Memory Limits Set", "✔️", "Memory Limits should be set.")
		}

		if container.Resource.Requests.CPU == "0" {
			containerTbl.AddRow(container.Name, "CPU Requests Not Set", "✖️", "You should set a CPU request for your container. See https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-requests-and-limits-of-pod-and-container for more information.")
		} else {
			containerTbl.AddRow(container.Name, "CPU Requests Set", "✔️", "CPU Requests should be set.")
		}

		if container.Resource.Requests.Memory == "0" {
			containerTbl.AddRow(container.Name, "Memory Requests Not Set", "✖️", "You should set a memory request for your container. See https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/#resource-requests-and-limits-of-pod-and-container for more information.")
		} else {
			containerTbl.AddRow(container.Name, "Memory Requests Set", "✔️", "Memory Requests should be set.")
		}
	}

	containerTbl.Print()
	c.Println("Pod-specific checks")
	fmt.Println()

	podTbl := table.New("Rule", "Compliance", "Reason")
	podTbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithPadding(5)

	if w.Deployment.Pod.QoS == "Guaranteed" {
		podTbl.AddRow("Pod QoS", "✔️", "Pod QoS is Guaranteed, this will get the highest priority with the scheduler, but it comes at the expense of having to set the CPU limit, which is not recommended.")
	} else {
		if w.Deployment.Pod.QoS == "Burstable" {
			podTbl.AddRow("Pod QoS", "✔️", "Pod QoS as Burstable is pretty reasonable. It means that the pod has a memory limit and a memory request, but no CPU limit.")
		} else {
			podTbl.AddRow("Pod QoS", "✖️", "Pod QoS as BestEffort is the lowest priority and should be avoided. See https://kubernetes.io/docs/tasks/configure-pod-container/quality-service-pod/ for more information")
		}
	}
	podTbl.Print()
	c.Println("Deployment-specific checks")
	fmt.Println()

	deploymentTbl := table.New("Rule", "Compliance", "Reason")
	deploymentTbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt).WithPadding(5)

	if w.Deployment.Replicas > 1 {
		deploymentTbl.AddRow("Replicas", "✔️", "You have more than one replica, which is good for high availability.")
	} else {
		deploymentTbl.AddRow("Replicas", "✖️", "You only have one replica, which is not recommended for high availability.")
	}

	if w.Deployment.HPASet {
		deploymentTbl.AddRow("HPA Set", "✔️", "You have an HPA set, which is good for autoscaling.")
	} else {
		deploymentTbl.AddRow("HPA Set", "✖️", "You do not have an HPA set, which is not recommended for autoscaling. We recommend to use Keda, please check the docs here: https://keda.sh/")
	}

	if w.Deployment.VPASet {
		deploymentTbl.AddRow("VPA Set", "✔️", "You have a VPA set, which is good for autoscaling.")
	} else {
		deploymentTbl.AddRow("VPA Set", "✖️", "You do not have a VPA set, which would be fine if you have any autoscaling set such as HPA's.")
	}

	if w.Deployment.PDBSet {
		deploymentTbl.AddRow("PDB Set", "✔️", "You have a PDB set, which is good for high availability.")
	} else {
		deploymentTbl.AddRow("PDB Set", "✖️", "You do not have a PDB set, which is not recommended for high availability. See https://kubernetes.io/docs/concepts/workloads/pods/disruptions/ for more information.")
	}

	deploymentTbl.Print()
}
