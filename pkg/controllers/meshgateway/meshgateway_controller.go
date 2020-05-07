package meshgateway

import (
	"context"
	"reflect"
	"time"

	"github.com/go-logr/logr"
	"github.com/gofrs/uuid"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	devopsv1beta1 "github.com/symcn/mid-operator/pkg/apis/devops/v1beta1"
	"github.com/symcn/mid-operator/pkg/controllers/resources/gateways"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

func GetWatchPredicateForMeshGateway() predicate.Funcs {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldObj := e.ObjectOld.(*devopsv1beta1.MeshGateway)
			newObj := e.ObjectNew.(*devopsv1beta1.MeshGateway)
			if !reflect.DeepEqual(oldObj.Spec, newObj.Spec) ||
				oldObj.GetDeletionTimestamp() != newObj.GetDeletionTimestamp() ||
				oldObj.GetGeneration() != newObj.GetGeneration() {
				return true
			}
			return false
		},
	}
}

// Add creates a new MeshGateway Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	dy, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return emperror.Wrap(err, "failed to create dynamic client")
	}

	return add(mgr, newReconciler(mgr, dy))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, d dynamic.Interface) reconcile.Reconciler {
	return &ReconcileMeshGateway{Client: mgr.GetClient(), dynamic: d, scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("meshgateway-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to MeshGateway
	err = c.Watch(&source.Kind{Type: &devopsv1beta1.MeshGateway{}}, &handler.EnqueueRequestForObject{}, GetWatchPredicateForMeshGateway())
	if err != nil {
		return err
	}

	// Watch for changes to resources created by the controller
	for _, t := range []runtime.Object{
		&corev1.ServiceAccount{TypeMeta: metav1.TypeMeta{Kind: "ServiceAccount", APIVersion: "v1"}},
		&rbacv1.Role{TypeMeta: metav1.TypeMeta{Kind: "Role", APIVersion: "v1"}},
		&rbacv1.RoleBinding{TypeMeta: metav1.TypeMeta{Kind: "RoleBinding", APIVersion: "v1"}},
		&rbacv1.ClusterRole{TypeMeta: metav1.TypeMeta{Kind: "ClusterRole", APIVersion: "v1"}},
		&rbacv1.ClusterRoleBinding{TypeMeta: metav1.TypeMeta{Kind: "ClusterRoleBinding", APIVersion: "v1"}},
		&corev1.Service{TypeMeta: metav1.TypeMeta{Kind: "Service", APIVersion: "v1"}},
		&appsv1.Deployment{TypeMeta: metav1.TypeMeta{Kind: "Deployment", APIVersion: "v1"}},
		&autoscalingv2beta1.HorizontalPodAutoscaler{TypeMeta: metav1.TypeMeta{Kind: "HorizontalPodAutoscaler", APIVersion: "v2beta1"}},
	} {
		err = c.Watch(&source.Kind{Type: t}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &devopsv1beta1.MeshGateway{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileMeshGateway{}

// ReconcileMeshGateway reconciles a MeshGateway object
type ReconcileMeshGateway struct {
	client.Client
	dynamic dynamic.Interface
	scheme  *runtime.Scheme
}

// Reconcile reads that state of the cluster for a MeshGateway object and makes changes based on the state read
// and what is in the MeshGateway.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=meshgateways;meshgateways/finalizers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=istio.banzaicloud.io,resources=meshgateways/status,verbs=get;update;patch
func (r *ReconcileMeshGateway) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger := log.WithValues("trigger", request.Namespace+"/"+request.Name, "correlationID", uuid.Must(uuid.NewV4()).String())

	// Fetch the MeshGateway instance
	instance := &devopsv1beta1.MeshGateway{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	instance.SetDefaults()

	err = updateStatus(r.Client, instance, devopsv1beta1.Reconciling, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	istio, err := r.getIstioForMeshGateway()
	if err != nil {
		log.Error(err, "failed to get istio")
		return reconcile.Result{}, err
	}

	reconciler := gateways.New(r.Client, r.dynamic, istio, instance)
	err = reconciler.Reconcile(log)
	if err == nil {
		instance.Status.GatewayAddress, err = reconciler.GetGatewayAddress()
		if err != nil {
			log.Error(err, "gateway address pending")
			return reconcile.Result{
				RequeueAfter: time.Second * 30,
			}, nil
		}
	} else {
		updateErr := updateStatus(r.Client, instance, devopsv1beta1.ReconcileFailed, err.Error(), logger)
		if updateErr != nil {
			logger.Error(updateErr, "failed to update state")
			return reconcile.Result{}, errors.WithStack(err)
		}
		return reconcile.Result{}, emperror.Wrap(err, "could not reconcile mesh gateway")
	}

	err = updateStatus(r.Client, instance, devopsv1beta1.Available, "", logger)
	if err != nil {
		return reconcile.Result{}, errors.WithStack(err)
	}

	return reconcile.Result{}, nil
}

func updateStatus(c client.Client, instance *devopsv1beta1.MeshGateway, status devopsv1beta1.ConfigState, errorMessage string, logger logr.Logger) error {
	typeMeta := instance.TypeMeta
	instance.Status.Status = status
	instance.Status.ErrorMessage = errorMessage
	err := c.Status().Update(context.Background(), instance)
	if k8serrors.IsNotFound(err) {
		err = c.Update(context.Background(), instance)
	}
	if err != nil {
		if !k8serrors.IsConflict(err) {
			return emperror.Wrapf(err, "could not update mesh gateway state to '%s'", status)
		}
		var actualInstance devopsv1beta1.MeshGateway
		err := c.Get(context.TODO(), types.NamespacedName{
			Namespace: instance.Namespace,
			Name:      instance.Name,
		}, &actualInstance)
		if err != nil {
			return emperror.Wrap(err, "could not get resource for updating status")
		}
		actualInstance.Status.Status = status
		actualInstance.Status.ErrorMessage = errorMessage
		err = c.Status().Update(context.Background(), &actualInstance)
		if k8serrors.IsNotFound(err) {
			err = c.Update(context.Background(), &actualInstance)
		}
		if err != nil {
			return emperror.Wrapf(err, "could not update mesh gateway state to '%s'", status)
		}
	}

	// update loses the typeMeta of the instace that's used later when setting ownerrefs
	instance.TypeMeta = typeMeta
	logger.Info("mesh gateway state updated", "status", status)
	return nil
}

func (r *ReconcileMeshGateway) getIstioForMeshGateway() (*devopsv1beta1.Istio, error) {
	var configs devopsv1beta1.IstioList
	err := r.Client.List(context.TODO(), &configs, &client.ListOptions{})
	if err != nil {
		return nil, emperror.Wrap(err, "could not list istio resources")
	}

	if len(configs.Items) != 1 {
		return nil, emperror.Wrap(err, "could not found istio resource")
	}

	istio := &configs.Items[0]
	return istio, nil
}
