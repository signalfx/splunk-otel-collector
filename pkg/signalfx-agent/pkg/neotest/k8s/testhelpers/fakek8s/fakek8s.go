package fakek8s

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimejson "k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/runtime/serializer/streaming"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	restwatch "k8s.io/client-go/rest/watch"
)

type resourceKind string
type resourceName string

// FakeK8s is a mock K8s API server.  It can serve both list and watch
// requests.
type FakeK8s struct {
	sync.RWMutex
	server *httptest.Server
	router http.Handler
	// Resources that have been inserted on the ResourceInput channel
	resources  map[resourceKind]map[string]map[resourceName]runtime.Object
	eventInput chan watch.Event
	// Channels to send new resources to watchers (we only support one watcher
	// per resource)
	subs          map[resourceKind]chan watch.Event
	pendingEvents map[resourceKind][]watch.Event
	subsMutex     sync.Mutex
	// Stops the resource accepter goroutine
	eventStopper chan struct{}
	// Stops all of the watchers
	stoppers map[resourceKind]chan struct{}
}

// NewFakeK8s makes a new FakeK8s
func NewFakeK8s() *FakeK8s {
	f := &FakeK8s{
		resources:     make(map[resourceKind]map[string]map[resourceName]runtime.Object),
		eventInput:    make(chan watch.Event),
		subs:          make(map[resourceKind]chan watch.Event),
		pendingEvents: make(map[resourceKind][]watch.Event),
		stoppers:      make(map[resourceKind]chan struct{}),
	}

	r := mux.NewRouter()

	r.HandleFunc("/api/v1/{resource}", f.handleListResource).Methods("GET")
	r.HandleFunc(`/api/v1/namespaces`, f.handleCreateOrReplaceResource).Methods("POST")
	r.HandleFunc(`/api/v1/namespaces`, f.handleCreateOrReplaceResource).Methods("PUT")
	r.HandleFunc(`/api/v1/namespaces`, f.handleCreateOrReplaceResource).Methods("DELETE")
	r.HandleFunc("/api/v1/namespaces/{namespace}/{resource}", f.handleListResource).Methods("GET")
	r.HandleFunc(`/api/v1/namespaces/{namespace}/{resource}`, f.handleCreateOrReplaceResource).Methods("POST")
	r.HandleFunc(`/api/v1/namespaces/{namespace}/{resource}/{name}`, f.handleCreateOrReplaceResource).Methods("PUT")
	r.HandleFunc(`/api/v1/namespaces/{namespace}/{resource}/{name}`, f.handleDeleteResource).Methods("DELETE")
	r.HandleFunc(`/api/v1/namespaces/{namespace}/{resource}/{name}`, f.handleGetResourceByName).Methods("GET")

	r.HandleFunc("/apis/apps/v1/{resource}", f.handleListResource).Methods("GET")
	r.HandleFunc(`/apis/apps/v1/namespaces/{namespace}/{resource}/{name}`, f.handleGetResourceByName).Methods("GET")
	r.HandleFunc("/apis/apps/v1/namespaces/{namespace}/{resource}", f.handleListResource).Methods("GET")
	r.HandleFunc("/apis/apps/v1/namespaces/{namespace}/{resource}", f.handleCreateOrReplaceResource).Methods("POST")
	r.HandleFunc("/apis/apps/v1/namespaces/{namespace}/{resource}/{name}", f.handleDeleteResource).Methods("DELETE")

	r.HandleFunc("/apis/batch/v1/{resource}", f.handleListResource).Methods("GET")
	r.HandleFunc("/apis/batch/v1/namespaces/{namespace}/{resource}", f.handleListResource).Methods("GET")
	r.HandleFunc("/apis/batch/v1/namespaces/{namespace}/{resource}", f.handleCreateOrReplaceResource).Methods("POST")
	r.HandleFunc(`/apis/batch/v1/namespaces/{namespace}/{resource}/{name}`, f.handleGetResourceByName).Methods("GET")
	r.HandleFunc("/apis/batch/v1/namespaces/{namespace}/{resource}/{name}", f.handleDeleteResource).Methods("DELETE")

	r.HandleFunc("/apis/batch/v1beta1/{resource}", f.handleListResource).Methods("GET")
	r.HandleFunc("/apis/batch/v1beta1/namespaces/{namespace}/{resource}", f.handleListResource).Methods("GET")
	r.HandleFunc("/apis/batch/v1beta1/namespaces/{namespace}/{resource}", f.handleCreateOrReplaceResource).Methods("POST")
	r.HandleFunc(`/apis/batch/v1beta1/namespaces/{namespace}/{resource}/{name}`, f.handleGetResourceByName).Methods("GET")
	r.HandleFunc("/apis/batch/v1beta1/namespaces/{namespace}/{resource}/{name}", f.handleDeleteResource).Methods("DELETE")

	r.Use(loggingMiddleware)

	f.router = r
	return f
}

// Start creates the server and starts it
func (f *FakeK8s) Start() {
	f.server = httptest.NewUnstartedServer(f.router)
	f.server.StartTLS()

	f.eventStopper = make(chan struct{})
	go f.acceptEvents(f.eventStopper)
}

// Close stops the server and all watchers
func (f *FakeK8s) Close() {
	f.subsMutex.Lock()
	defer f.subsMutex.Unlock()

	close(f.eventStopper)
	for _, ch := range f.stoppers {
		close(ch)
	}

	f.server.Listener.Close()
	//f.server.Close()

}

// URL is the of the mock server to point your objects under test to
func (f *FakeK8s) URL() string {
	return f.server.URL
}

// SetInitialList adds resources to the server state that are served when doing
// list requests.  l can be a list of any of the supported resource types.
func (f *FakeK8s) SetInitialList(l []runtime.Object) {
	for _, r := range l {
		resKind := resourceKind(r.GetObjectKind().GroupVersionKind().Kind)
		objMeta := r.(metav1.ObjectMetaAccessor).GetObjectMeta()
		f.addToResources(resKind, objMeta.GetNamespace(), objMeta.GetName(), r)
	}
}

func (f *FakeK8s) acceptEvents(stopper <-chan struct{}) {
	log.Debugf("Generating Fake k8s events")
	for {
		select {
		case <-stopper:
			return
		case e := <-f.eventInput:
			resKind := resourceKind(e.Object.GetObjectKind().GroupVersionKind().Kind)

			f.subsMutex.Lock()
			// Send it out to any watchers
			// TODO: This only supports watchers across all namespaces
			if f.subs[resKind] != nil {
				log.Infof("Watch event sent to subscription: %s", spew.Sdump(e))
				f.subs[resKind] <- e
			} else {
				f.pendingEvents[resKind] = append(f.pendingEvents[resKind], e)
				log.Infof("Watch event ignored because nothing was watching for %s", resKind)
			}
			f.subsMutex.Unlock()
		}
	}
}

// Returns whether the object was created (true) or simply replaced (false)
func (f *FakeK8s) addToResources(resKind resourceKind, namespace string, name string, resource runtime.Object) bool {
	f.Lock()
	defer f.Unlock()

	if f.resources[resKind] == nil {
		f.resources[resKind] = make(map[string]map[resourceName]runtime.Object)
	}
	if f.resources[resKind][namespace] == nil {
		f.resources[resKind][namespace] = make(map[resourceName]runtime.Object)
	}

	_, exists := f.resources[resKind][namespace][resourceName(name)]
	f.resources[resKind][namespace][resourceName(name)] = resource
	return !exists
}

func (f *FakeK8s) handleCreateOrReplaceResource(rw http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		panic("Got bad request")
	}
	obj, _, _ := scheme.Codecs.UniversalDeserializer().Decode(content, nil, nil)
	ns := mux.Vars(r)["namespace"]
	if ns != "" {
		obj.(metav1.ObjectMetaAccessor).GetObjectMeta().SetNamespace(ns)
	}
	f.CreateOrReplaceResource(obj)
}

// CreateOrReplaceResource can be used by unit tests.  It will generate a watch
// event.
func (f *FakeK8s) CreateOrReplaceResource(obj runtime.Object) {
	resKind := resourceKind(obj.GetObjectKind().GroupVersionKind().Kind)
	objMeta := obj.(metav1.ObjectMetaAccessor).GetObjectMeta()
	created := f.addToResources(resKind, objMeta.GetNamespace(), objMeta.GetName(), obj)

	var eType watch.EventType
	if created {
		eType = watch.Added
	} else {
		eType = watch.Modified
	}

	f.eventInput <- watch.Event{Type: eType, Object: obj}
}

func (f *FakeK8s) handleDeleteResource(rw http.ResponseWriter, r *http.Request) {
	f.DeleteResourceByName(string(pluralNameToKind(mux.Vars(r)["resource"])), mux.Vars(r)["namespace"], mux.Vars(r)["name"])
}

// DeleteResource deletes a resource and sends a watch event if actually
// deleted.
func (f *FakeK8s) DeleteResource(obj runtime.Object) bool {
	objMeta := obj.(metav1.ObjectMetaAccessor).GetObjectMeta()
	return f.DeleteResourceByName(obj.GetObjectKind().GroupVersionKind().Kind, objMeta.GetNamespace(), objMeta.GetName())
}

// DeleteResourceByName removes a resource from the fake api server.  It will
// generate a watch event for the deletion if the resource existed.
func (f *FakeK8s) DeleteResourceByName(resKind string, namespace string, name string) bool {
	if namespaces := f.resources[resourceKind(resKind)]; namespaces != nil {
		if names := namespaces[namespace]; names != nil {
			name := resourceName(name)
			if obj, ok := names[name]; ok {
				log.Infof("Deleting %s %s/%s", resKind, namespace, name)
				delete(names, name)
				f.eventInput <- watch.Event{Type: watch.Deleted, Object: obj}
				return true
			}
		}
	}
	return false
}

func (f *FakeK8s) handleGetResourceByName(rw http.ResponseWriter, r *http.Request) {
	namespaces := f.resources[pluralNameToKind(mux.Vars(r)["resource"])]
	if namespaces == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	names := namespaces[mux.Vars(r)["namespace"]]
	if names == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	obj := names[resourceName(mux.Vars(r)["name"])]
	if obj == nil {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	s, _ := json.Marshal(obj)
	rw.Header().Add("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(s)
}

func (f *FakeK8s) handleListResource(rw http.ResponseWriter, r *http.Request) {

	rw.Header().Add("Content-Type", "application/json")

	isWatch := strings.Contains(r.URL.RawQuery, "watch=true")
	namespace := mux.Vars(r)["namespace"]
	resKind := pluralNameToKind(mux.Vars(r)["resource"])

	if isWatch {
		rw.Header().Add("Transfer-Encoding", "chunked")
		// This must block in order to continue to be able to write to the
		// ResponseWriter
		if namespace != "" {
			panic("Watches within a single namespace aren't supported")
		}
		f.startWatcher(resKind, rw)
	} else {
		f.sendList(resKind, namespace, rw)
	}
}

// Start a long running routine that will send everything received on the
// `EventInput` channel as JSON back to the client
func (f *FakeK8s) startWatcher(resKind resourceKind, rw http.ResponseWriter) {
	f.subsMutex.Lock()

	if f.subs[resKind] != nil {
		panic("We don't support more than one watcher at a time!")
	}

	log.Infof("Adding watcher for %s", resKind)
	eventCh := make(chan watch.Event)
	f.stoppers[resKind] = make(chan struct{})
	// Alias so we only access map inside lock
	stopper := f.stoppers[resKind]
	f.subs[resKind] = eventCh
	if len(f.pendingEvents[resKind]) > 0 {
		go func() {
			for _, e := range f.pendingEvents[resKind] {
				f.subs[resKind] <- e
			}
		}()
	}

	f.subsMutex.Unlock()
	rw.WriteHeader(200)

	for {
		select {
		case r := <-eventCh:
			buf := &bytes.Buffer{}
			jsonSerializer := runtimejson.NewSerializer(runtimejson.DefaultMetaFactory, scheme.Scheme, scheme.Scheme, false)
			innerEncoder := scheme.Codecs.WithoutConversion().EncoderForVersion(jsonSerializer, v1.SchemeGroupVersion)

			encoder := restwatch.NewEncoder(streaming.NewEncoder(buf, innerEncoder), innerEncoder)
			if err := encoder.Encode(&r); err != nil {
				panic("could not encode watch event")
			}
			_, _ = rw.Write(buf.Bytes())
			_, _ = rw.Write([]byte("\n"))
			rw.(http.Flusher).Flush()
		case <-stopper:
			f.subsMutex.Lock()
			delete(f.subs, resKind)
			f.subsMutex.Unlock()
			return
		}
	}
}

func (f *FakeK8s) sendList(resKind resourceKind, namespace string, rw http.ResponseWriter) {
	items := make([]runtime.RawExtension, 0)

	addFromNamespace := func(ns string) {
		for _, i := range f.resources[resKind][ns] {
			items = append(items, runtime.RawExtension{
				Object: i,
			})
		}
	}

	if namespace == "" {
		for ns := range f.resources[resKind] {
			addFromNamespace(ns)
		}
	} else {
		addFromNamespace(namespace)
	}

	l := v1.List{
		TypeMeta: typeMeta(resKind),
		ListMeta: metav1.ListMeta{},
		Items:    items,
	}

	d, _ := json.Marshal(l)
	log.Debugf("list: %s", string(d))

	rw.WriteHeader(200)
	_, _ = rw.Write(d)
}

func pluralNameToKind(name string) resourceKind {
	switch name {
	case "pods":
		return "Pod"
	case "replicationcontrollers":
		return "ReplicationController"
	case "deployments":
		return "Deployment"
	case "statefulsets":
		return "StatefulSet"
	case "daemonsets":
		return "DaemonSet"
	case "replicasets":
		return "ReplicaSet"
	case "namespaces":
		return "Namespace"
	case "resourcequotas":
		return "ResourceQuota"
	case "nodes":
		return "Node"
	case "secrets":
		return "Secret"
	case "services":
		return "Service"
	case "jobs":
		return "Job"
	case "cronjobs":
		return "CronJob"
	default:
		panic("Unknown resource type: " + name)
	}
}
func typeMeta(rt resourceKind) metav1.TypeMeta {
	switch string(rt) {
	case "Pod":
		return metav1.TypeMeta{Kind: "PodList", APIVersion: "v1"}
	case "ReplicationController":
		return metav1.TypeMeta{Kind: "ReplicationControllerList", APIVersion: "v1"}
	case "Deployment":
		return metav1.TypeMeta{Kind: "DeploymentList", APIVersion: "apps/v1"}
	case "StatefulSet":
		return metav1.TypeMeta{Kind: "StatefulSetList", APIVersion: "apps/v1"}
	case "DaemonSet":
		return metav1.TypeMeta{Kind: "DaemonSetList", APIVersion: "apps/v1"}
	case "ReplicaSet":
		return metav1.TypeMeta{Kind: "ReplicaSetList", APIVersion: "apps/v1"}
	case "Namespace":
		return metav1.TypeMeta{Kind: "NamespaceList", APIVersion: "v1"}
	case "ResourceQuota":
		return metav1.TypeMeta{Kind: "ResourceQuotaList", APIVersion: "v1"}
	case "Node":
		return metav1.TypeMeta{Kind: "NodeList", APIVersion: "v1"}
	case "Secret":
		return metav1.TypeMeta{Kind: "SecretList", APIVersion: "v1"}
	case "Service":
		return metav1.TypeMeta{Kind: "ServiceList", APIVersion: "v1"}
	case "Job":
		return metav1.TypeMeta{Kind: "JobList", APIVersion: "batch/v1"}
	case "CronJob":
		return metav1.TypeMeta{Kind: "CronJobList", APIVersion: "batch/v1beta1"}
	default:
		panic("Unknown resource type: " + string(rt))
	}
}
