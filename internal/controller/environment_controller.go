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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	platformv1 "opencloud.io/environment/api/v1"
	v1 "opencloud.io/environment/api/v1"
)

// EnvironmentReconciler reconciles a Environment object
type EnvironmentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
func (r *EnvironmentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	//fetch environment object
	var environment v1.Environment

	if err := r.Get(ctx, req.NamespacedName, &environment); err != nil {
		logger.Error(err, "unable to fetch environment")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	// provision compute resources
	if err := r.provisionCompute(ctx, &environment); err != nil {
		logger.Error(err, "failed to provision compute resources")
		return ctrl.Result{}, err
	}

	// Update the status
	environment.Status.LastUpdateTime = metav1.Now() // metav1.Now() is a function from the k8s.io/apimachinery/pkg/apis/meta/v1 package, part of the Kubernetes API machinery. This function returns the current time in a format that is suitable for Kubernetes API objects, specifically as a metav1.Time
	if err := r.Status().Update(ctx, &environment); err != nil {
		logger.Error(err, "failed to update Environment status")
		return ctrl.Result{}, err
	}

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *EnvironmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1.Environment{}).
		Complete(r)
}

// provisionCompute provisions the compute resources (e.g., EC2 instances)
func (r *EnvironmentReconciler) provisionCompute(ctx context.Context, environment *v1.Environment) error {
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

	return nil
}
