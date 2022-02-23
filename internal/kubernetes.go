package internal

import (
	"path"

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

func KubernetesListNodes(ctx *Context) (*v1.NodeList, error) {
	clientset, _, err := kubernetesInit(ctx)
	if err != nil {
		return nil, err
	}
	nodes, err := clientset.CoreV1().Nodes().List(*ctx.Ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func KubernetesCreateFromManifest(ctx *Context, namespace string, manifest string) error {
	clientset, config, err := kubernetesInit(ctx)
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
	err = kubernetesCreateObject(clientset, *config, namespace, obj)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	return nil
}

func kubernetesInit(ctx *Context) (*kubernetes.Clientset, *rest.Config, error) {
	kubeconfigFile := path.Join(ctx.Dir, "kubeconfig")
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

func kubernetesCreateObject(kubeClientset kubernetes.Interface, restConfig rest.Config, namespace string, obj runtime.Object) error {
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
