package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	labelPatchLabelsPod = `[{"op":"add","path":"/metadata/labels", "value":{"team":"ops"}}]`

	mutatedAnnotation = `[{"op":"add","path":"/metadata/annotations", "value":{"mutated":"true"}}]`

	config = Config{
		CertFile: "./ssl-certs/cert.pem",
		KeyFile:  "./ssl-certs/key.pem",
	}

	scheme = runtime.NewScheme()

	codecs = serializer.NewCodecFactory(scheme)
)

func addToScheme(scheme *runtime.Scheme) {
	corev1.AddToScheme(scheme)
	admissionregistrationv1beta1.AddToScheme(scheme)
}

type Config struct {
	CertFile string
	KeyFile  string
}

func toAdmissionResponse(err error) *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

// Only provision pods and deployments which have allowed labels
func admitPodsDeployments(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	logrus.Info("admission review for resources")
	var (
		resourceLabel map[string]string
		resourceType  string
	)

	req := ar.Request

	switch req.Kind.Kind {
	case "Deployment":
		var deployment appsv1.Deployment
		if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceType = "Deployment"
		resourceLabel = deployment.Spec.Template.ObjectMeta.Labels

	case "Pod":
		var pod corev1.Pod
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			logrus.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceType = "Pod"
		resourceLabel = pod.Labels
	}

	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	var msg string
	if v, ok := resourceLabel["team"]; ok {
		if v != "ops" {
			reviewResponse.Allowed = false
			msg = fmt.Sprintf(msg+"The Pod template in the %s contains an unwanted label", resourceType)
		}
	}
	if !reviewResponse.Allowed {
		reviewResponse.Result = &metav1.Status{
			Message: strings.TrimSpace(msg),
		}
		return &reviewResponse
	}
	return mutatePodsDeployments(ar)
}

// Mutate pods and/or deployments
func mutatePodsDeployments(ar v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	logrus.Info("adding label to a Resource")

	req := ar.Request

	var pod corev1.Pod
	resourceAnnotation := pod.ObjectMeta.Annotations
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		logrus.Errorf("Could not unmarshal raw object: %v", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	if val, ok := resourceAnnotation["mutated"]; ok {
		logrus.Info("Annotation exists")
		if val == "true" {
			logrus.Info("already mutated")
			return &reviewResponse
		}
	}

	if req.Kind.Kind == "Pod" {
		logrus.Info("patching a pod")
		reviewResponse.Patch = []byte(labelPatchLabelsPod)
	}

	reviewResponse.Patch = []byte(mutatedAnnotation)

	pt := v1beta1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	logrus.Printf("added patch %v", string(reviewResponse.Patch))
	return &reviewResponse
}

func serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		logrus.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse

	ar := v1beta1.AdmissionReview{}

	deserializer := codecs.UniversalDeserializer()
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		logrus.Error(err)
		admissionResponse = toAdmissionResponse(err)
	} else {
		admissionResponse = admitPodsDeployments(ar)
	}
	returnedAdmissionReview := v1beta1.AdmissionReview{}

	if admissionResponse != nil {
		returnedAdmissionReview.Response = admissionResponse
		returnedAdmissionReview.Response.UID = ar.Request.UID
	}

	responseInBytes, err := json.Marshal(returnedAdmissionReview)
	logrus.Info(string(responseInBytes))

	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Info("Writing response")
	if _, err := w.Write(responseInBytes); err != nil {
		logrus.Error(err)
	}
}

func init() {

	addToScheme(scheme)
}

func main() {
	sCert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		logrus.Fatal(err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}

	http.HandleFunc("/mutate", serve)
	server := &http.Server{
		Addr:      ":443",
		TLSConfig: tlsConfig,
	}

	logrus.Info("Starting server...")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		logrus.Errorf("Unable to start server, %s", err)
	}
}
