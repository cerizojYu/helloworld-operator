/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	demov1alpha1 "github.com/cerizoj/helloword/api/v1alpha1"
)

// HelloworldReconciler reconciles a Helloworld object
type HelloworldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var finalizer = "helloworlds.demo.github.com/finalizer"

//+kubebuilder:rbac:groups=demo.github.com,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=demo.github.com,resources=helloworlds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=demo.github.com,resources=helloworlds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Helloworld object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *HelloworldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	l.Info("receive cr change", "namespacedName", req.NamespacedName)
	cr := &demov1alpha1.Helloworld{}
	// fetch cr by namespacedName
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		l.Info("can't fetch cr, maybe deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if cr.ObjectMeta.DeletionTimestamp.IsZero() {
		// add finalizer if not exist
		if !controllerutil.ContainsFinalizer(cr, finalizer) {
			l.Info("add finalizer", "finalizer", finalizer)
			controllerutil.AddFinalizer(cr, finalizer)
			if err := r.Update(ctx, cr); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// cr delete logic
		err := r.deleteComponents(ctx, cr)
		if err != nil {
			l.Error(err, "delete components failed")
			return ctrl.Result{}, err
		}
		l.Info("delete components successfully")
		controllerutil.RemoveFinalizer(cr, finalizer)
		if err := r.Update(ctx, cr); err != nil {
			return ctrl.Result{}, err
		}
	}
	// exit loop with failed status
	if cr.Status.DeployStatus == demov1alpha1.FailedStatus || cr.Status.SerivceStatus == demov1alpha1.FailedStatus || cr.Status.IngressStatus == demov1alpha1.FailedStatus {
		return reconcile.Result{}, fmt.Errorf("exit loop with deploy_status: %s; serivce_status: %s; ingress_status: %s", cr.Status.DeployStatus, cr.Status.SerivceStatus, cr.Status.IngressStatus)
	}
	// apply deployment
	// TODO: wait pods running
	if cr.Status.DeployStatus != demov1alpha1.ActiveStatus {
		deploy, err := r.constructDeployment(ctx, cr)
		if err != nil {
			l.Error(err, "construct deployment failed")
			cr.Status.DeployStatus = demov1alpha1.FailedStatus
			return ctrl.Result{}, r.Status().Update(ctx, cr)
		}
		if err := r.Create(ctx, deploy); err != nil {
			l.Error(err, "create deployment failed")
			cr.Status.DeployStatus = demov1alpha1.FailedStatus
		} else {
			l.Info("create deployment successfully")
			cr.Status.DeployStatus = demov1alpha1.ActiveStatus

		}
		return ctrl.Result{}, r.Status().Update(ctx, cr)
	}
	// apply service
	if cr.Status.SerivceStatus != demov1alpha1.ActiveStatus {
		service, err := r.constructService(ctx, cr)
		if err != nil {
			l.Error(err, "construct deployment failed")
			cr.Status.SerivceStatus = demov1alpha1.FailedStatus
			return ctrl.Result{}, r.Status().Update(ctx, cr)
		}
		if err := r.Create(ctx, service); err != nil {
			l.Error(err, "create service failed")
			cr.Status.SerivceStatus = demov1alpha1.FailedStatus
		} else {
			l.Info("create service successfully")
			cr.Status.SerivceStatus = demov1alpha1.ActiveStatus

		}
		return ctrl.Result{}, r.Status().Update(ctx, cr)
	}
	// apply ingress rule
	if cr.Status.IngressStatus != demov1alpha1.ActiveStatus {
		service, err := r.constructIngress(ctx, cr)
		if err != nil {
			l.Error(err, "construct ingress failed")
			cr.Status.IngressStatus = demov1alpha1.FailedStatus
			return ctrl.Result{}, r.Status().Update(ctx, cr)
		}
		if err := r.Create(ctx, service); err != nil {
			l.Error(err, "create ingress failed")
			cr.Status.IngressStatus = demov1alpha1.FailedStatus
		} else {
			l.Info("create ingress successfully")
			cr.Status.IngressStatus = demov1alpha1.ActiveStatus

		}
		return ctrl.Result{}, r.Status().Update(ctx, cr)
	}
	return ctrl.Result{}, nil
}

func generateDeploymentName(cr *demov1alpha1.Helloworld) string {
	return fmt.Sprintf("deployment-%s", cr.Name)
}

func generateServiceName(cr *demov1alpha1.Helloworld) string {
	return fmt.Sprintf("service-%s", cr.Name)
}

func generateIngressName(cr *demov1alpha1.Helloworld) string {
	return fmt.Sprintf("ingress-%s", cr.Name)
}

func (r *HelloworldReconciler) constructDeployment(ctx context.Context, cr *demov1alpha1.Helloworld) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      generateDeploymentName(cr),
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "helloworld",
				},
			},
			Replicas: &cr.Spec.Replicas,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "helloworld",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "helloworld",
							Image: "gcr.io/google-samples/hello-app:1.0",
						},
					},
				},
			},
		},
	}
	err := controllerutil.SetControllerReference(cr, deploy, r.Scheme)
	return deploy, err
}

func (r *HelloworldReconciler) constructService(ctx context.Context, cr *demov1alpha1.Helloworld) (*v1.Service, error) {
	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      generateServiceName(cr),
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{
				"app": "helloworld",
			},
			Ports: []v1.ServicePort{
				{
					Protocol:   v1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(8080),
				},
			},
			Type: v1.ServiceTypeClusterIP,
		},
	}
	err := controllerutil.SetControllerReference(cr, svc, r.Scheme)
	return svc, err
}

func (r *HelloworldReconciler) constructIngress(ctx context.Context, cr *demov1alpha1.Helloworld) (*networkingv1.Ingress, error) {
	ingressRule := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      generateIngressName(cr),
		},
		Spec: *cr.Spec.Ingress.DeepCopy(),
	}
	err := controllerutil.SetControllerReference(cr, ingressRule, r.Scheme)
	return ingressRule, err
}

func (r *HelloworldReconciler) deleteComponents(ctx context.Context, cr *demov1alpha1.Helloworld) error {
	// delete ingress rule
	ingressRule := &networkingv1.Ingress{}
	ingressRuleKey := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      generateIngressName(cr),
	}
	if err := r.Get(ctx, ingressRuleKey, ingressRule); err == nil {
		return r.Delete(ctx, ingressRule)
	}
	// delete service
	service := &v1.Service{}
	serviceKey := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      generateServiceName(cr),
	}
	if err := r.Get(ctx, serviceKey, service); err == nil {
		return r.Delete(ctx, service)
	}
	// delete deployment
	deploy := &appsv1.Deployment{}
	deployKey := types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      generateDeploymentName(cr),
	}
	if err := r.Get(ctx, deployKey, deploy); err == nil {
		return r.Delete(ctx, deploy)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloworldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&demov1alpha1.Helloworld{}).
		Complete(r)
}
