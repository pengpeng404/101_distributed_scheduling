package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	admission "k8s.io/api/admission/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	deployResource = metav1.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
)

func affinityFunc(req *admission.AdmissionRequest) ([]patchOperation, error) {
	// only process deployment resources; other objects are passed through
	if req.Resource != deployResource {
		log.Printf("The resource is %s, expect resource to be %s", req.Resource, deployResource)
		return nil, nil
	}

	// parse the deployment object
	raw := req.Object.Raw
	deploy := appsv1.Deployment{}
	if _, _, err := universalDeserializer.Decode(raw, nil, &deploy); err != nil {
		return nil, fmt.Errorf("could not deserialize deploy object: %v", err)
	}

	var patches []patchOperation
	if deploy.Spec.Template.Spec.Affinity == nil {
		deploy.Spec.Template.Spec.Affinity = &corev1.Affinity{}
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/spec/template/spec/affinity",
			Value: map[string]interface{}{},
		})
	}

	if deploy.Spec.Template.Spec.Affinity.NodeAffinity == nil {
		deploy.Spec.Template.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/spec/template/spec/affinity/nodeAffinity",
			Value: map[string]interface{}{},
		})
	}

	if deploy.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
		deploy.Spec.Template.Spec.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.PreferredSchedulingTerm{}
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/spec/template/spec/affinity/nodeAffinity/preferredDuringSchedulingIgnoredDuringExecution",
			Value: []interface{}{},
		})
	}

	// single instance deployment might be a pseudo-requirement
	if *deploy.Spec.Replicas == 1 {
		// single instance should be on on-demand
		patches = append(patches, patchOperation{
			Op:   "add",
			Path: "/spec/template/spec/affinity/nodeAffinity/preferredDuringSchedulingIgnoredDuringExecution/-",
			Value: corev1.PreferredSchedulingTerm{
				Weight: 10,
				Preference: corev1.NodeSelectorTerm{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "node.kubernetes.io/capacity",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"on-demand"},
						},
					},
				},
			},
		})
		return patches, nil
	}

	patches = append(patches, patchOperation{
		Op:   "add",
		Path: "/spec/template/spec/affinity/nodeAffinity/preferredDuringSchedulingIgnoredDuringExecution/-",
		Value: corev1.PreferredSchedulingTerm{
			Weight: 20,
			Preference: corev1.NodeSelectorTerm{
				MatchExpressions: []corev1.NodeSelectorRequirement{
					{
						Key:      "node.kubernetes.io/capacity",
						Operator: corev1.NodeSelectorOpIn,
						Values:   []string{"spot"},
					},
				},
			},
		},
	})

	if deploy.Spec.Template.Spec.Affinity.PodAntiAffinity == nil {
		deploy.Spec.Template.Spec.Affinity.PodAntiAffinity = &corev1.PodAntiAffinity{}
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/spec/template/spec/affinity/podAntiAffinity",
			Value: map[string]interface{}{},
		})
	}

	if deploy.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution == nil {
		deploy.Spec.Template.Spec.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = []corev1.WeightedPodAffinityTerm{}
		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  "/spec/template/spec/affinity/podAntiAffinity/preferredDuringSchedulingIgnoredDuringExecution",
			Value: []interface{}{},
		})
	}

	for key, value := range deploy.Spec.Template.Labels {
		patches = append(patches, patchOperation{
			Op:   "add",
			Path: "/spec/template/spec/affinity/podAntiAffinity/preferredDuringSchedulingIgnoredDuringExecution/-",
			Value: corev1.WeightedPodAffinityTerm{
				Weight: 30,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      key,
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{value},
							},
						},
					},
					TopologyKey: "topology.kubernetes.io/zone",
				},
			},
		})

		patches = append(patches, patchOperation{
			Op:   "add",
			Path: "/spec/template/spec/affinity/podAntiAffinity/preferredDuringSchedulingIgnoredDuringExecution/-",
			Value: corev1.WeightedPodAffinityTerm{
				Weight: 20,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      key,
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{value},
							},
						},
					},
					TopologyKey: "beta.kubernetes.io/instance-type",
				},
			},
		})

		patches = append(patches, patchOperation{
			Op:   "add",
			Path: "/spec/template/spec/affinity/podAntiAffinity/preferredDuringSchedulingIgnoredDuringExecution/-",
			Value: corev1.WeightedPodAffinityTerm{
				Weight: 10,
				PodAffinityTerm: corev1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							{
								Key:      key,
								Operator: metav1.LabelSelectorOpIn,
								Values:   []string{value},
							},
						},
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		})
	}

	return patches, nil
}

func main() {
	certPath := filepath.Join(tlsDir, tlsCertFile)
	keyPath := filepath.Join(tlsDir, tlsKeyFile)

	mux := http.NewServeMux()
	mux.Handle("/mutate", admitFuncHandler(affinityFunc))
	server := &http.Server{
		Addr:    ":8443",
		Handler: mux,
	}
	log.Fatal(server.ListenAndServeTLS(certPath, keyPath))
}
