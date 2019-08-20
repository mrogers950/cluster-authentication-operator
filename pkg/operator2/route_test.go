package operator2

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/fake"
	testing2 "k8s.io/client-go/testing"

	configv1 "github.com/openshift/api/config/v1"
	v1 "github.com/openshift/api/route/v1"
	routefake "github.com/openshift/client-go/route/clientset/versioned/fake"
)

func Test_authOperator_handleRoute(t *testing.T) {
	var tests = map[string]struct {
		ingress             *configv1.Ingress
		expectedRoute       *v1.Route
		routeStatusOnCreate *v1.RouteStatus
		routeStatusOnUpdate *v1.RouteStatus
		expectedSecret      *corev1.Secret
		objects             []runtime.Object
		routeObjects        []runtime.Object
		expectedErr         string
		expectRouteUpdate   bool
		expectRouteCreate   bool
	}{
		"create-route": {
			ingress: &configv1.Ingress{
				Spec: configv1.IngressSpec{
					Domain: "apps.example.com",
				},
			},
			expectedRoute: &v1.Route{
				ObjectMeta: defaultMeta(),
				Spec: v1.RouteSpec{
					Host: "oauth-openshift.apps.example.com",
					To: v1.RouteTargetReference{
						Kind: "Service",
						Name: targetName,
					},
					Port: &v1.RoutePort{
						TargetPort: intstr.FromInt(containerPort),
					},
					TLS: &v1.TLSConfig{
						Termination:                   v1.TLSTerminationPassthrough,
						InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
					},
				},
				Status: v1.RouteStatus{
					Ingress: []v1.RouteIngress{
						{
							Host: "oauth-openshift.apps.example.com",
							Conditions: []v1.RouteIngressCondition{
								{
									Type:   v1.RouteAdmitted,
									Status: corev1.ConditionTrue,
								},
							},
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerCertsLocalName,
					Namespace: targetNamespace,
				},
				Data: map[string][]byte{
					"f": []byte("a"), // contents dont matter, just that data is present
				},
			},
			objects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routerCertsLocalName,
						Namespace: targetNamespace,
					},
					Data: map[string][]byte{
						"f": []byte("a"), // contents dont matter, just that data is present
					},
				},
			},
			expectRouteCreate: true,
			routeStatusOnCreate: &v1.RouteStatus{
				Ingress: []v1.RouteIngress{
					{
						Host: "oauth-openshift.apps.example.com",
						Conditions: []v1.RouteIngressCondition{
							{
								Type:   v1.RouteAdmitted,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
		},
		"route-exists": {
			ingress: &configv1.Ingress{
				Spec: configv1.IngressSpec{
					Domain: "apps.example.com",
				},
			},
			expectedRoute: &v1.Route{
				ObjectMeta: defaultMeta(),
				Spec: v1.RouteSpec{
					Host: "oauth-openshift.apps.example.com",
					To: v1.RouteTargetReference{
						Kind: "Service",
						Name: targetName,
					},
					Port: &v1.RoutePort{
						TargetPort: intstr.FromInt(containerPort),
					},
					TLS: &v1.TLSConfig{
						Termination:                   v1.TLSTerminationPassthrough,
						InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
					},
				},
				Status: v1.RouteStatus{
					Ingress: []v1.RouteIngress{
						{
							Host: "oauth-openshift.apps.example.com",
							Conditions: []v1.RouteIngressCondition{
								{
									Type:   v1.RouteAdmitted,
									Status: corev1.ConditionTrue,
								},
							},
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerCertsLocalName,
					Namespace: targetNamespace,
				},
				Data: map[string][]byte{
					"f": []byte("a"), // contents dont matter, just that data is present
				},
			},
			objects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routerCertsLocalName,
						Namespace: targetNamespace,
					},
					Data: map[string][]byte{
						"f": []byte("a"), // contents dont matter, just that data is present
					},
				},
			},
			routeObjects: []runtime.Object{
				&v1.Route{
					ObjectMeta: defaultMeta(),
					Spec: v1.RouteSpec{
						Host: "oauth-openshift.apps.example.com", // mimic the behavior of subdomain
						To: v1.RouteTargetReference{
							Kind: "Service",
							Name: targetName,
						},
						Port: &v1.RoutePort{
							TargetPort: intstr.FromInt(containerPort),
						},
						TLS: &v1.TLSConfig{
							Termination:                   v1.TLSTerminationPassthrough,
							InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
						},
					},
					Status: v1.RouteStatus{
						Ingress: []v1.RouteIngress{
							{
								Host: "oauth-openshift.apps.example.com",
								Conditions: []v1.RouteIngressCondition{
									{
										Type:   v1.RouteAdmitted,
										Status: corev1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			},
		},
		"route-update": {
			ingress: &configv1.Ingress{
				Spec: configv1.IngressSpec{
					Domain: "bar.example.com",
				},
			},
			expectedRoute: &v1.Route{
				ObjectMeta: defaultMeta(),
				Spec: v1.RouteSpec{
					Host: "oauth-openshift.bar.example.com",
					To: v1.RouteTargetReference{
						Kind: "Service",
						Name: targetName,
					},
					Port: &v1.RoutePort{
						TargetPort: intstr.FromInt(containerPort),
					},
					TLS: &v1.TLSConfig{
						Termination:                   v1.TLSTerminationPassthrough,
						InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
					},
				},
				Status: v1.RouteStatus{
					Ingress: []v1.RouteIngress{
						{
							Host: "oauth-openshift.bar.example.com",
							Conditions: []v1.RouteIngressCondition{
								{
									Type:   v1.RouteAdmitted,
									Status: corev1.ConditionTrue,
								},
							},
						},
					},
				},
			},
			expectRouteUpdate: true,
			routeStatusOnUpdate: &v1.RouteStatus{
				Ingress: []v1.RouteIngress{
					{
						Host: "oauth-openshift.bar.example.com",
						Conditions: []v1.RouteIngressCondition{
							{
								Type:   v1.RouteAdmitted,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerCertsLocalName,
					Namespace: targetNamespace,
				},
				Data: map[string][]byte{
					"f": []byte("a"), // contents dont matter, just that data is present
				},
			},
			objects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routerCertsLocalName,
						Namespace: targetNamespace,
					},
					Data: map[string][]byte{
						"f": []byte("a"), // contents dont matter, just that data is present
					},
				},
			},
			routeObjects: []runtime.Object{
				&v1.Route{
					ObjectMeta: defaultMeta(),
					Spec: v1.RouteSpec{
						Host: "oauth-openshift.apps.example.com", // mimic the behavior of subdomain
						To: v1.RouteTargetReference{
							Kind: "Service",
							Name: targetName,
						},
						Port: &v1.RoutePort{
							TargetPort: intstr.FromInt(containerPort),
						},
						TLS: &v1.TLSConfig{
							Termination:                   v1.TLSTerminationPassthrough,
							InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
						},
					},
					Status: v1.RouteStatus{
						Ingress: []v1.RouteIngress{
							{
								Host: "oauth-openshift.apps.example.com",
								Conditions: []v1.RouteIngressCondition{
									{
										Type:   v1.RouteAdmitted,
										Status: corev1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			},
		},
		"route-update-invalid-route": {
			ingress: &configv1.Ingress{
				Spec: configv1.IngressSpec{
					Domain: "apps.example.com",
				},
			},
			expectedRoute: &v1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      targetName,
					Namespace: targetNamespace,
					Labels: map[string]string{
						"app": targetName,
					},
					Annotations: map[string]string{
						"annotationToPreserve": "foo",
					},
				},
				Spec: v1.RouteSpec{
					Host: "oauth-openshift.apps.example.com",
					To: v1.RouteTargetReference{
						Kind: "Service",
						Name: targetName,
					},
					Port: &v1.RoutePort{
						TargetPort: intstr.FromInt(containerPort),
					},
					TLS: &v1.TLSConfig{
						Termination:                   v1.TLSTerminationPassthrough,
						InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
					},
				},
				Status: v1.RouteStatus{
					Ingress: []v1.RouteIngress{
						{
							Host: "oauth-openshift.apps.example.com",
							Conditions: []v1.RouteIngressCondition{
								{
									Type:   v1.RouteAdmitted,
									Status: corev1.ConditionTrue,
								},
							},
						},
					},
				},
			},
			expectRouteUpdate: true,
			routeStatusOnUpdate: &v1.RouteStatus{
				Ingress: []v1.RouteIngress{
					{
						Host: "oauth-openshift.apps.example.com",
						Conditions: []v1.RouteIngressCondition{
							{
								Type:   v1.RouteAdmitted,
								Status: corev1.ConditionTrue,
							},
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerCertsLocalName,
					Namespace: targetNamespace,
				},
				Data: map[string][]byte{
					"f": []byte("a"), // contents dont matter, just that data is present
				},
			},
			objects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routerCertsLocalName,
						Namespace: targetNamespace,
					},
					Data: map[string][]byte{
						"f": []byte("a"), // contents dont matter, just that data is present
					},
				},
			},
			routeObjects: []runtime.Object{
				&v1.Route{
					ObjectMeta: metav1.ObjectMeta{
						Name:      targetName,
						Namespace: targetNamespace,
						Labels: map[string]string{
							"app": targetName,
						},
						Annotations: map[string]string{
							"annotationToPreserve": "foo",
						},
					},
					Spec: v1.RouteSpec{
						Host: "oauth-openshift.apps.example.com", // mimic the behavior of subdomain
						To: v1.RouteTargetReference{
							Kind: "Service",
							Name: targetName,
						},
						Port: &v1.RoutePort{
							TargetPort: intstr.FromInt(containerPort),
						},
						TLS: nil, // This invalidates the route
					},
					Status: v1.RouteStatus{
						Ingress: []v1.RouteIngress{
							{
								Host: "oauth-openshift.apps.example.com",
								Conditions: []v1.RouteIngressCondition{
									{
										Type:   v1.RouteAdmitted,
										Status: corev1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			},
		},
		"route-secret-empty": {
			ingress: &configv1.Ingress{
				Spec: configv1.IngressSpec{
					Domain: "apps.example.com",
				},
			},
			expectedRoute: &v1.Route{
				ObjectMeta: defaultMeta(),
				Spec: v1.RouteSpec{
					Host: "oauth-openshift.apps.example.com",
					To: v1.RouteTargetReference{
						Kind: "Service",
						Name: targetName,
					},
					Port: &v1.RoutePort{
						TargetPort: intstr.FromInt(containerPort),
					},
					TLS: &v1.TLSConfig{
						Termination:                   v1.TLSTerminationPassthrough,
						InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
					},
				},
				Status: v1.RouteStatus{
					Ingress: []v1.RouteIngress{
						{
							Host: "oauth-openshift.apps.example.com",
							Conditions: []v1.RouteIngressCondition{
								{
									Type:   v1.RouteAdmitted,
									Status: corev1.ConditionTrue,
								},
							},
						},
					},
				},
			},
			expectedSecret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routerCertsLocalName,
					Namespace: targetNamespace,
				},
			},
			objects: []runtime.Object{
				&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      routerCertsLocalName,
						Namespace: targetNamespace,
					},
				},
			},
			routeObjects: []runtime.Object{
				&v1.Route{
					ObjectMeta: defaultMeta(),
					Spec: v1.RouteSpec{
						Host: "oauth-openshift.apps.example.com", // mimic the behavior of subdomain
						To: v1.RouteTargetReference{
							Kind: "Service",
							Name: targetName,
						},
						Port: &v1.RoutePort{
							TargetPort: intstr.FromInt(containerPort),
						},
						TLS: &v1.TLSConfig{
							Termination:                   v1.TLSTerminationPassthrough,
							InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
						},
					},
					Status: v1.RouteStatus{
						Ingress: []v1.RouteIngress{
							{
								Host: "oauth-openshift.apps.example.com",
								Conditions: []v1.RouteIngressCondition{
									{
										Type:   v1.RouteAdmitted,
										Status: corev1.ConditionTrue,
									},
								},
							},
						},
					},
				},
			},
			expectedErr: "router secret openshift-authentication/v4-0-config-system-router-certs is empty",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := fake.NewSimpleClientset(tt.objects...)
			routeClient := routefake.NewSimpleClientset(tt.routeObjects...)
			routeClient.PrependReactor("create", "routes", func(action testing2.Action) (bool, runtime.Object, error) {
				t.Logf("create route")
				create := action.(testing2.CreateAction)
				rt := create.GetObject().(*v1.Route)
				rt.Status = *tt.routeStatusOnCreate
				return true, rt, nil
			})
			routeClient.PrependReactor("update", "routes", func(action testing2.Action) (bool, runtime.Object, error) {
				t.Logf("update route")
				update := action.(testing2.UpdateAction)
				rt := update.GetObject().(*v1.Route)
				rt.Status = *tt.routeStatusOnUpdate
				return true, rt, nil
			})

			c := &authOperator{
				secrets:    client.CoreV1(),
				configMaps: client.CoreV1(),
				route:      routeClient.RouteV1().Routes(targetNamespace),
			}

			route, secret, _, err := c.handleRoute(tt.ingress)
			if err != nil {
				if len(tt.expectedErr) == 0 {
					t.Errorf("unexpected error %s", err)
				} else if tt.expectedErr != err.Error() {
					t.Errorf("expected error %s, got %s", tt.expectedErr, err)
				}
			} else {
				if len(tt.expectedErr) != 0 {
					t.Errorf("expected error %s, got no error", tt.expectedErr)
				}

				routeActions := routeClient.Actions()
				if tt.expectRouteCreate {
					var created bool
					for _, act := range routeActions {
						if act.GetVerb() == "create" {
							created = true
						}
					}
					if !created {
						t.Errorf("expected route creation")
					}
				}
				if tt.expectRouteUpdate {
					var updated bool
					for _, act := range routeActions {
						if act.GetVerb() == "update" {
							updated = true
						}
					}
					if !updated {
						t.Errorf("expected route creation")
					}
				}

				if !reflect.DeepEqual(tt.expectedRoute, route) {
					t.Errorf("expected route %#v, got %#v", tt.expectedRoute, route)
				}
				if !reflect.DeepEqual(tt.expectedSecret, secret) {
					t.Errorf("handleConfigSync() secrets got = %v, want %v", secret, tt.expectedSecret)
				}
			}
		})
	}
}
