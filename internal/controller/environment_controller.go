/*
Copyright 2024.

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

package controller

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	platformv1 "opencloud.io/environment/api/v1"
	v1 "opencloud.io/environment/api/v1"
)

// EnvironmentReconciler reconciles a Environment object
type EnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger logr.Logger
}

// +kubebuilder:rbac:groups=platform.opencloud.io,resources=environments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=platform.opencloud.io,resources=environments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=platform.opencloud.io,resources=environments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Environment object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.2/pkg/reconcile
func (e *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := e.Logger
	log.WithValues("environment_name ", req.Name, "environment_namespace", req.Namespace).V(1).Info("Reconcile started")
	//fetch environment object
	var environment v1.Environment

	if err := e.Get(ctx, req.NamespacedName, &environment); err != nil {
		if apierrors.IsNotFound(err) {
			// The resource no longer exists, so it must have been deleted
			log.Info("Environment resource not found. It must have been deleted.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch environment object")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// make a deep copy object for environment
	envCopy := environment.DeepCopy()
	// Check if the object has deletionTimeStamp.
	if environment.ObjectMeta.DeletionTimestamp.IsZero() {
		// If the obejct does not have finalizers and if it is not being deleted - then add finalizers.
		if !controllerutil.ContainsFinalizer(envCopy, "service.compute") {
			if controllerutil.AddFinalizer(envCopy, "service.compute") {
				log.Info("finalizer service.compute has been added to the environment - ", "environment", envCopy.Name)
				if err := e.Update(ctx, envCopy); err != nil {
					return ctrl.Result{}, err // Do not add the object back in the queue
				}
			} else {
				log.Info("could not add finalizer service.compute to the environment - ", envCopy.Name)
			}
		}
		// provision compute resources
		if err := e.provisionCompute(ctx, &environment); err != nil {
			log.Error(err, "failed to provision compute resources")
			return ctrl.Result{}, err
		}

		// computeReady := checkIFComputeReady()

		currentStatus := environment.Status.DeepCopy()
		newCondition := metav1.Condition{
			Type:               "ComputeReady",
			Status:             metav1.ConditionTrue,
			Reason:             "ResourcesProvisioned",
			Message:            "Compute resources are ready",
			LastTransitionTime: metav1.Now(),
		}

		// Check if the condition already exists and update it
		existingCondition := findCondition(currentStatus.Conditions, newCondition.Type)
		if existingCondition != nil {
			if existingCondition.Status != newCondition.Status || existingCondition.Reason != newCondition.Reason || existingCondition.Message != newCondition.Message {
				existingCondition.Status = newCondition.Status
				existingCondition.Reason = newCondition.Reason
				existingCondition.Message = newCondition.Message
				existingCondition.LastTransitionTime = metav1.Now()
			}
		} else {
			currentStatus.Conditions = append(currentStatus.Conditions, newCondition)
		}

		err := e.UpdateStatus(ctx, envCopy, currentStatus)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if controllerutil.ContainsFinalizer(&environment, "service.compute") {
			log.Info("deprovisioning compute service before removing finalizer")
			err := e.deProvisionCompute(ctx, &environment)
			if err != nil {
				// TODO - Handle error in deprovisioning
			}
			controllerutil.RemoveFinalizer(&environment, "service.compute")
			if err := e.Update(ctx, &environment); err != nil {
				return ctrl.Result{}, err // Do not add the object back in the queue
			}
		}
	}

	return ctrl.Result{RequeueAfter: 2 * time.Minute}, nil
}

// TODO - Fix Logger - No Debug level as of now. Create a central logger and pass it everywhere.

// SetupWithManager sets up the controller with the Manager.
func (e *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1.Environment{}).
		Complete(e)
}

// provisionCompute provisions the compute resources (e.g., EC2 instances)
func (e *EnvironmentReconciler) provisionCompute(ctx context.Context, environment *v1.Environment) error {
	/*
		sess := session.Must(session.NewSession())
		ec2Svc := ec2.New(sess)

		for _, server := range environment.Spec.ComputeService.Servers {
			// Create EC2 instance
			input := &ec2.RunInstancesInput{
				ImageId:      aws.String("ami-0c55b159cbfafe1f0"), // Update with your AMI ID
				InstanceType: aws.String("t2.micro"),              // Update with your instance type
				MinCount:     aws.Int64(1),
				MaxCount:     aws.Int64(1),
			}
			_, err := ec2Svc.RunInstances(input)
			if err != nil {
				return err
			}
		}
	*/
	return nil
}

// provisionCompute provisions the compute resources (e.g., EC2 instances)
func (e *EnvironmentReconciler) deProvisionCompute(ctx context.Context, environment *v1.Environment) error {
	/*
		sess := session.Must(session.NewSession())
		ec2Svc := ec2.New(sess)

		for _, server := range environment.Spec.ComputeService.Servers {
			// Create EC2 instance
			input := &ec2.RunInstancesInput{
				ImageId:      aws.String("ami-0c55b159cbfafe1f0"), // Update with your AMI ID
				InstanceType: aws.String("t2.micro"),              // Update with your instance type
				MinCount:     aws.Int64(1),
				MaxCount:     aws.Int64(1),
			}
			_, err := ec2Svc.RunInstances(input)
			if err != nil {
				return err
			}
		}
	*/
	return nil
}

func checkIFComputeReady() bool {
	return true
}

func (e *EnvironmentReconciler) UpdateStatus(ctx context.Context, env *platformv1.Environment, desiredStatus *platformv1.EnvironmentStatus) error {
	if !reflect.DeepEqual(env.Status, desiredStatus) {
		env.Status = *desiredStatus
		if err := e.Status().Update(ctx, env); err != nil {
			e.Logger.Error(err, "Failed to update Environment status")
			return err
		}
	}

	return nil
}

// Helper function to find a condition by type
func findCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
