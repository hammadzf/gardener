// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package apiserver

import (
	"google.golang.org/protobuf/types/known/durationpb"
	istioapinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
)

func (g *gardenerAPIServer) destinationRule() *istionetworkingv1beta1.DestinationRule {
	return &istionetworkingv1beta1.DestinationRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName,
			Namespace: g.namespace,
			Labels:    GetLabels(),
		},
		Spec: istioapinetworkingv1beta1.DestinationRule{
			Host: kubernetesutils.FQDNForService(serviceName, g.namespace),
			TrafficPolicy: &istioapinetworkingv1beta1.TrafficPolicy{
				LoadBalancer: &istioapinetworkingv1beta1.LoadBalancerSettings{
					LbPolicy: &istioapinetworkingv1beta1.LoadBalancerSettings_Simple{
						Simple: istioapinetworkingv1beta1.LoadBalancerSettings_LEAST_REQUEST,
					},
				},
				Tls: &istioapinetworkingv1beta1.ClientTLSSettings{
					Mode: istioapinetworkingv1beta1.ClientTLSSettings_DISABLE,
				},
				ConnectionPool: &istioapinetworkingv1beta1.ConnectionPoolSettings{
					Tcp: &istioapinetworkingv1beta1.ConnectionPoolSettings_TCPSettings{
						MaxConnections: 1024,
						TcpKeepalive: &istioapinetworkingv1beta1.ConnectionPoolSettings_TCPSettings_TcpKeepalive{
							Time:     &durationpb.Duration{Seconds: 30},
							Interval: &durationpb.Duration{Seconds: 10},
							Probes:   3,
						},
					},
				},
			},
		},
	}
}

func (g *gardenerAPIServer) virtualService() *istionetworkingv1beta1.VirtualService {
	host := kubernetesutils.FQDNForService(serviceName, g.namespace)

	return &istionetworkingv1beta1.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DeploymentName,
			Namespace: g.namespace,
			Labels:    GetLabels(),
		},
		Spec: istioapinetworkingv1beta1.VirtualService{
			Hosts: []string{
				host,
				serviceName,
			},
			Tls: []*istioapinetworkingv1beta1.TLSRoute{{
				Match: []*istioapinetworkingv1beta1.TLSMatchAttributes{{
					Port: servicePort,
					SniHosts: []string{
						serviceName,
						host,
					},
				}},
				Route: []*istioapinetworkingv1beta1.RouteDestination{{
					Destination: &istioapinetworkingv1beta1.Destination{
						Host: host,
						Port: &istioapinetworkingv1beta1.PortSelector{
							Number: servicePort,
						},
					},
				}},
			}},
		},
	}
}
