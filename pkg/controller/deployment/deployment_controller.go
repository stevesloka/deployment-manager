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

package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	semver "github.com/coreos/go-semver/semver"
	managedv1beta1 "github.com/stevesloka/deployment-manager/pkg/apis/managed/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const managedLabelKey = "managed.stevesloka.com"
const managedLabelVersion = "0.0.1"

// Add creates a new Deployment Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this apps.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDeployment{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("deployment-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDeployment{}

// ReconcileDeployment reconciles a Deployment object
type ReconcileDeployment struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Deployment object and makes changes based on the state read
// and what is in the Deployment.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileDeployment) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the ManagedDeployment instance
	instance := &appsv1.Deployment{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("---------------- Deployment not found")

			instanceManaged := &managedv1beta1.ManagedDeployment{}
			err = r.Get(context.TODO(), request.NamespacedName, instanceManaged)
			if err == nil {
				err = r.Delete(context.TODO(), instanceManaged)
				if err != nil {
					fmt.Println("---------- ERROR deleting managed resource: ", err)
				}
			}
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Serialize the deployment to json
	deploymentSerialized, err := json.Marshal(instance)
	if err != nil {
		fmt.Println("----------- ERROR serialzing struct to json", err)
		return reconcile.Result{}, err
	}

	// Only process Deployments if they are labeled as managed
	if mapContains(instance.Labels) {
		// Lookup manageddeployed
		instanceManaged := &managedv1beta1.ManagedDeployment{}
		err = r.Get(context.TODO(), request.NamespacedName, instanceManaged)
		if err != nil {
			if errors.IsNotFound(err) {
				version := semver.New(managedLabelVersion)
				version.BumpPatch()

				instanceManaged = &managedv1beta1.ManagedDeployment{
					ObjectMeta: metav1.ObjectMeta{
						Name:      instance.Name,
						Namespace: instance.Namespace,
						Labels: map[string]string{
							"managedVersion": version.String(),
						},
					},
					Spec: managedv1beta1.ManagedDeploymentSpec{
						PreviousConfiguration: string(deploymentSerialized),
					},
				}

				err = r.Create(context.TODO(), instanceManaged)
				if err != nil {
					fmt.Println("------- ERROR creating managed crd")
					return reconcile.Result{}, err
				}
			}
		} else {
			// Get the previous version
			version := semver.New(instanceManaged.Labels["managedVersion"])
			version.BumpPatch()

			// Update the CRD
			instanceManaged.Labels = map[string]string{
				"managedVersion": version.String(),
			}

			instanceManaged.Spec.PreviousConfiguration = string(deploymentSerialized)
			err = r.Update(context.TODO(), instanceManaged)
			if err != nil {
				fmt.Println("------- ERROR updating managed crd")
				return reconcile.Result{}, err
			}
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileDeployment) deleteExternalDependency(instance *appsv1.Deployment) error {
	log.Printf("deleting the external dependencies")
	//
	// delete the external dependency here
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple types for same object.
	return nil
}

//
// Helper functions to check and remove string from a slice of strings.
//
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func mapContains(values map[string]string) bool {
	if val, ok := values[managedLabelKey]; ok {
		if val == "true" {
			return true
		}
	}
	return false
}
