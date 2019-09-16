/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ssetcdcluster

import (
	"context"
	"reflect"

	infrav1 "github.com/jkevlin/etcd-operator-statefulset/pkg/apis/infra/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new SSEtcdCluster Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSSEtcdCluster{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ssetcdcluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to SSEtcdCluster
	err = c.Watch(&source.Kind{Type: &infrav1.SSEtcdCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a StatefulSet created by SSEtcdCluster - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &infrav1.SSEtcdCluster{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileSSEtcdCluster{}

// ReconcileSSEtcdCluster reconciles a SSEtcdCluster object
type ReconcileSSEtcdCluster struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a SSEtcdCluster object and makes changes based on the state read
// and what is in the SSEtcdCluster.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a StatefulSet as an example
// Automatically generate RBAC rules to allow the Controller to read and write StatefulSets
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infra.github.ibm.com,resources=ssetcdclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.github.ibm.com,resources=ssetcdclusters/status,verbs=get;update;patch
func (r *ReconcileSSEtcdCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the SSEtcdCluster instance
	instance := &infrav1.SSEtcdCluster{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// TODO(user): Change this to be the object type created by your controller
	// Define the desired StatefulSet object
	replicas := int32(3)

	deploy := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"statefulset": instance.Name, "app": instance.Name},
			},
			Replicas:    &replicas,
			ServiceName: instance.Name,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name: instance.Name,
					Labels: map[string]string{
						"statefulset": instance.Name,
						"app":         instance.Name,
					}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            instance.Name,
							Image:           "quay.io/coreos/etcd:latest",
							ImagePullPolicy: "IfNotPresent",
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name:          "peer",
									ContainerPort: 2380,
								},
								corev1.ContainerPort{
									Name:          "client",
									ContainerPort: 2379,
								},
							},
							Env: []corev1.EnvVar{
								corev1.EnvVar{
									Name:  "INITIAL_CLUSTER_SIZE",
									Value: "3",
								},
								corev1.EnvVar{
									Name:  "SET_NAME",
									Value: instance.Name,
								},
								corev1.EnvVar{
									Name:  "SET_DOMAIN",
									Value: instance.Name,
								},
							},
							Command: []string{
								"/bin/sh",
								"-ec",
								`
								echo "HOSTNAME=${HOSTNAME}"
								echo "SET_NAME=${SET_NAME}"
								echo "SET_DOMAIN=${SET_DOMAIN}"
								PEERS="${SET_NAME}-0=http://${SET_NAME}-0.${SET_DOMAIN}:2380,${SET_NAME}-1=http://${SET_NAME}-1.${SET_DOMAIN}:2380,${SET_NAME}-2=http://${SET_NAME}-2.${SET_DOMAIN}:2380"
								exec etcd --name ${HOSTNAME} \
								  --listen-peer-urls http://0.0.0.0:2380 \
								  --listen-client-urls http://0.0.0.0:2379 \
								  --advertise-client-urls http://${HOSTNAME}:2379 \
								  --initial-advertise-peer-urls http://${HOSTNAME}:2380 \
								  --initial-cluster-token ${SET_NAME}-cluster-1 \
								  --initial-cluster ${PEERS} \
								  --initial-cluster-state new \
								  --data-dir /var/run/etcd/default.${SET_DOMAIN}
								`,
							},
						},
					},
				},
			},
		},
	}
	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the StatefulSet already exists
	found := &appsv1.StatefulSet{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating StatefulSet", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Create(context.TODO(), deploy)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		log.Info("Updating StatefulSet", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}
