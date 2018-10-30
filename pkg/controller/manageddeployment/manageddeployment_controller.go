/*
Copyright 2018 Steve Sloka.

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

package manageddeployment

import (
	"context"
	"fmt"

	managedv1beta1 "github.com/stevesloka/deployment-manager/pkg/apis/managed/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ManagedDeployment Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this managed.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileManagedDeployment{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("manageddeployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to ManagedDeployment
	err = c.Watch(&source.Kind{Type: &managedv1beta1.ManagedDeployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by ManagedDeployment - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &managedv1beta1.ManagedDeployment{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileManagedDeployment{}

// ReconcileManagedDeployment reconciles a ManagedDeployment object
type ReconcileManagedDeployment struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a ManagedDeployment object and makes changes based on the state read
// and what is in the ManagedDeployment.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=managed.stevesloka.com,resources=manageddeployments,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileManagedDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	// Fetch the ManagedDeployment instance
	instance := &managedv1beta1.ManagedDeployment{}
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

	fmt.Println("----------- ManagedDeploymentCRD Changed! ", instance.GetName())

	// // TODO(user): Change this to be the object type created by your controller
	// // Define the desired Deployment object
	// deploy := &appsv1.Deployment{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      instance.Name + "-deployment",
	// 		Namespace: instance.Namespace,
	// 	},
	// 	Spec: appsv1.DeploymentSpec{
	// 		Selector: &metav1.LabelSelector{
	// 			MatchLabels: map[string]string{"deployment": instance.Name + "-deployment"},
	// 		},
	// 		Template: corev1.PodTemplateSpec{
	// 			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"deployment": instance.Name + "-deployment"}},
	// 			Spec: corev1.PodSpec{
	// 				Containers: []corev1.Container{
	// 					{
	// 						Name:  "nginx",
	// 						Image: "nginx",
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	// if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // TODO(user): Change this for the object type created by your controller
	// // Check if the Deployment already exists
	// found := &appsv1.Deployment{}
	// err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	log.Printf("Creating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
	// 	err = r.Create(context.TODO(), deploy)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // TODO(user): Change this for the object type created by your controller
	// // Update the found object and write the result back if there are any changes
	// if !reflect.DeepEqual(deploy.Spec, found.Spec) {
	// 	found.Spec = deploy.Spec
	// 	log.Printf("Updating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
	// 	err = r.Update(context.TODO(), found)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}
	// }
	return reconcile.Result{}, nil
}
