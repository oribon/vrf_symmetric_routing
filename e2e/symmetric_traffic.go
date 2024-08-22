package e2e

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	externalHostIP string
)

const (
	testNamespace = "the-namespace"

	localDeployment = "the-local-deployment"
	localService    = "the-local-service"
	localServiceLB  = "192.200.10.1"

	clusterDeployment = "the-cluster-deployment"
	clusterService    = "the-cluster-service"
	clusterServiceLB  = "192.200.10.2"

	servicesPort     = 9090
	externalHostPort = 9090
)

var _ = ginkgo.Describe("E2E", func() {
	var cs clientset.Interface
	externalHostDst := net.JoinHostPort(externalHostIP, fmt.Sprint(externalHostPort))
	localServiceDst := net.JoinHostPort(localServiceLB, fmt.Sprint(servicesPort))
	clusterServiceDst := net.JoinHostPort(clusterServiceLB, fmt.Sprint(servicesPort))

	ginkgo.BeforeEach(func() {
		cs = newClient()
		out, err := exec.Command("kubectl", "apply", "-f", "apps.yaml").CombinedOutput()
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), string(out))

		gomega.Eventually(func() error {
			err := deploymentReady(cs, localDeployment, testNamespace)
			if err != nil {
				return err
			}

			return deploymentReady(cs, clusterDeployment, testNamespace)
		}, 2*time.Minute, time.Second).ShouldNot(gomega.HaveOccurred())
	})

	ginkgo.AfterEach(func() {
		out, err := exec.Command("kubectl", "delete", "-f", "apps.yaml").CombinedOutput()
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), out)

		gomega.Eventually(func() bool {
			_, err := cs.CoreV1().Namespaces().Get(context.TODO(), testNamespace, metav1.GetOptions{})
			return k8serrors.IsNotFound(err)
		}, 2*time.Minute, time.Second).Should(gomega.BeTrue())
	})

	ginkgo.It("verifies symmetric routing works as expected", func() {
		clusterPods, err := podsForDeployment(cs, clusterDeployment, testNamespace)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		localPods, err := podsForDeployment(cs, localDeployment, testNamespace)
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		ginkgo.By("Verifying all of the pods reach the external host with their LB's IP")
		for _, p := range localPods {
			out, err := execOnPod(p, "curl", "-s", fmt.Sprintf("http://%s/clientip", externalHostDst))
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			ip, _, err := net.SplitHostPort(out)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(ip).To(gomega.BeEquivalentTo(localServiceLB))
		}

		for _, p := range clusterPods {
			out, err := execOnPod(p, "curl", "-s", fmt.Sprintf("http://%s/clientip", externalHostDst))
			gomega.Expect(err).ToNot(gomega.HaveOccurred())

			ip, _, err := net.SplitHostPort(out)
			gomega.Expect(err).ToNot(gomega.HaveOccurred())
			gomega.Expect(ip).To(gomega.BeEquivalentTo(clusterServiceLB))
		}

		ginkgo.By("Verifying the external host reaches the ETP=Local service with its IP")
		out, err := exec.Command("curl", "-s", fmt.Sprintf("http://%s/clientip", localServiceDst)).CombinedOutput()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		ip, _, err := net.SplitHostPort(string(out))
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
		gomega.Expect(ip).To(gomega.BeEquivalentTo(externalHostIP))

		ginkgo.By("Verifying the external host reaches the ETP=Cluster service")
		out, err = exec.Command("curl", "-s", fmt.Sprintf("http://%s/clientip", clusterServiceDst)).CombinedOutput()
		gomega.Expect(err).ToNot(gomega.HaveOccurred(), out)

	})

})

func newClient() *clientset.Clientset {
	config := ctrl.GetConfigOrDie()
	res, err := clientset.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return res
}

func deploymentReady(cs clientset.Interface, name, namespace string) error {
	deployment, err := cs.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	pods, err := podsForDeployment(cs, name, namespace)
	if err != nil {
		return err
	}

	if len(pods) != int(*deployment.Spec.Replicas) {
		return fmt.Errorf("deployment %s has %d replicas but %d pods created", name, *deployment.Spec.Replicas, len(pods))
	}

	for _, p := range pods {
		if p.Status.Phase != corev1.PodRunning {
			return fmt.Errorf("deployment %s pod %s is not running, in phase %s", name, p.Name, p.Status.Phase)
		}
	}

	return nil
}

func podsForDeployment(cs clientset.Interface, name, namespace string) ([]corev1.Pod, error) {
	deployment, err := cs.AppsV1().Deployments(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	pods, err := cs.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=" + deployment.Spec.Selector.MatchLabels["app"],
	})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) != int(*deployment.Spec.Replicas) {
		return nil, fmt.Errorf("deployment %s has %d replicas but %d pods created", name, *deployment.Spec.Replicas, len(pods.Items))
	}

	return pods.Items, nil
}

func execOnPod(pod corev1.Pod, cmd string, args ...string) (string, error) {
	fullargs := append([]string{"exec", pod.Name, "-n", pod.Namespace, "--", cmd}, args...)
	out, err := exec.Command("kubectl", fullargs...).CombinedOutput()
	return string(out), err
}
