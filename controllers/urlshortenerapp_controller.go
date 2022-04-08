/*
Copyright 2022.

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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"reflect"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	batchv1 "tutorial.kubebuilder.io/project/api/v1"
)

// UrlShortenerAppReconciler reconciles a UrlShortenerApp object
type UrlShortenerAppReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=urlshortenerapps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=urlshortenerapps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.tutorial.kubebuilder.io,resources=urlshortenerapps/finalizers,verbs=update
func (r *UrlShortenerAppReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("urlshortener", req.NamespacedName)

	urlshortener := &batchv1.UrlShortenerApp{}
	err := r.Get(ctx, req.NamespacedName, urlshortener)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("UrlShortener resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get UrlShortener")
		return ctrl.Result{}, err
	}

	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: urlshortener.Name, Namespace: urlshortener.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		dep := r.deploymentForUrlShortener(urlshortener)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}
	size := urlshortener.Spec.Size
	if *found.Spec.Replicas != size {
		found.Spec.Replicas = &size
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(urlshortener.Namespace),
		client.MatchingLabels(labelsForUrlShortener(urlshortener.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "UrlShortenerApp.Namespace", urlshortener.Namespace, "UrlShortenerApp.Name", urlshortener.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	if !reflect.DeepEqual(podNames, urlshortener.Status.Nodes) {
		urlshortener.Status.Nodes = podNames
		err := r.Status().Update(ctx, urlshortener)
		if err != nil {
			log.Error(err, "Failed to update UrlShortener status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *UrlShortenerAppReconciler) deploymentForUrlShortener(app *batchv1.UrlShortenerApp) *appsv1.Deployment {
	ls := labelsForUrlShortener(app.Name)
	replicas := app.Spec.Size

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           "ashwin901/url-shortener",
						Name:            "urlshortener",
						ImagePullPolicy: corev1.PullAlways,
						Ports: []corev1.ContainerPort{{
							ContainerPort: app.Spec.ContainerPort,
							Name:          "urlshortener",
						}},
					}},
				},
			},
		},
	}
	ctrl.SetControllerReference(app, dep, r.Scheme)
	return dep
}

func labelsForUrlShortener(name string) map[string]string {
	return map[string]string{"app": "urlshortener", "urlshortener_cr": name}
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *UrlShortenerAppReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.UrlShortenerApp{}).
		Complete(r)
}
