package clients

import (
	"fmt"
	"path"
	"time"

	"github.com/airfocusio/hcloud-talos/internal/cluster"
	"github.com/airfocusio/hcloud-talos/internal/utils"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

var metadataAccessor = apimeta.NewAccessor()

func KubernetesWaitNodeRegistered(cl *cluster.Cluster, name string) error {
	clientset, _, err := KubernetesInit(cl)
	if err != nil {
		return err
	}
	return utils.RetrySlow(cl.Logger, func() error {
		_, err := clientset.CoreV1().Nodes().Get(*cl.Ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func KubernetesWaitPodRunning(cl *cluster.Cluster, namespace string, name string) error {
	clientset, _, err := KubernetesInit(cl)
	if err != nil {
		return err
	}
	return utils.RetrySlow(cl.Logger, func() error {
		pod, err := clientset.CoreV1().Pods(namespace).Get(*cl.Ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if pod.Status.Phase != v1.PodRunning {
			return fmt.Errorf("pod not yet running")
		}
		return nil
	})
}

func KubernetesCreateFromManifest(cl *cluster.Cluster, manifest string) error {
	clientset, config, err := KubernetesInit(cl)
	if err != nil {
		return err
	}

	sch := runtime.NewScheme()
	err = scheme.AddToScheme(sch)
	if err != nil {
		return err
	}
	decode := serializer.NewCodecFactory(sch).UniversalDeserializer().Decode
	obj, _, err := decode([]byte(manifest), nil, nil)
	if err != nil {
		return err
	}
	err = KubernetesCreateObject(clientset, *config, obj)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func KubernetesDeleteNode(cl *cluster.Cluster, name string) error {
	clientset, _, err := KubernetesInit(cl)
	if err != nil {
		return err
	}

	gracePeriodSeconds := int64(time.Minute.Seconds())
	err = clientset.CoreV1().Nodes().Delete(*cl.Ctx, name, metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	})
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	return nil
}

func KubernetesInit(cl *cluster.Cluster) (*kubernetes.Clientset, *rest.Config, error) {
	kubeconfigFile := path.Join(cl.Dir, "kubeconfig")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile)
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return clientset, config, nil
}

func KubernetesCreateObject(kubeClientset kubernetes.Interface, restConfig rest.Config, obj runtime.Object) error {
	// Create a REST mapper that tracks information about the available resources in the cluster.
	groupResources, err := restmapper.GetAPIGroupResources(kubeClientset.Discovery())
	if err != nil {
		return err
	}
	rm := restmapper.NewDiscoveryRESTMapper(groupResources)

	// Get some metadata needed to make the REST request.
	gvk := obj.GetObjectKind().GroupVersionKind()
	gk := schema.GroupKind{Group: gvk.Group, Kind: gvk.Kind}
	mapping, err := rm.RESTMapping(gk, gvk.Version)
	if err != nil {
		return err
	}

	_, err = apimeta.NewAccessor().Name(obj)
	if err != nil {
		return err
	}

	// Create a client specifically for creating the object.
	restClient, err := kubernetesNewRestClient(restConfig, mapping.GroupVersionKind.GroupVersion())
	if err != nil {
		return err
	}

	// Detect namespace
	namespace, err := metadataAccessor.Namespace(obj)
	if err != nil {
		return err
	}

	// Use the REST helper to create the object in the "default" namespace.
	restHelper := resource.NewHelper(restClient, mapping)
	_, err = restHelper.Create(namespace, false, obj)
	return err
}

func kubernetesNewRestClient(restConfig rest.Config, gv schema.GroupVersion) (rest.Interface, error) {
	restConfig.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	restConfig.GroupVersion = &gv
	if len(gv.Group) == 0 {
		restConfig.APIPath = "/api"
	} else {
		restConfig.APIPath = "/apis"
	}

	return rest.RESTClientFor(&restConfig)
}
