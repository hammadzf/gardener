// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package apiserver

import (
	"context"

	"google.golang.org/protobuf/types/known/durationpb"
	istioapinetworkingv1beta1 "istio.io/api/networking/v1beta1"
	istionetworkingv1beta1 "istio.io/client-go/pkg/apis/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kubeapiserverconstants "github.com/gardener/gardener/pkg/component/kubernetes/apiserver/constants"
	"github.com/gardener/gardener/pkg/controllerutils"
	kubernetesutils "github.com/gardener/gardener/pkg/utils/kubernetes"
)

func (k *kubeAPIServer) emptyDestinationRule() *istionetworkingv1beta1.DestinationRule {
	return &istionetworkingv1beta1.DestinationRule{ObjectMeta: metav1.ObjectMeta{Name: kubeapiserverconstants.ServiceName(k.values.NamePrefix) + "-mesh", Namespace: k.namespace}}
}

func (k *kubeAPIServer) emptyVirtualService() *istionetworkingv1beta1.VirtualService {
	return &istionetworkingv1beta1.VirtualService{ObjectMeta: metav1.ObjectMeta{Name: kubeapiserverconstants.ServiceName(k.values.NamePrefix) + "-mesh", Namespace: k.namespace}}
}

func (k *kubeAPIServer) destinationRule() *istionetworkingv1beta1.DestinationRule {
	dr := k.emptyDestinationRule()
	serviceName := kubeapiserverconstants.ServiceName(k.values.NamePrefix)
	dr.Labels = getLabels()
	dr.Spec = istioapinetworkingv1beta1.DestinationRule{
		ExportTo: []string{"."},
		Host:     kubernetesutils.FQDNForService(serviceName, k.namespace),
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
	}
	return dr
}

func (k *kubeAPIServer) virtualService() *istionetworkingv1beta1.VirtualService {
	vs := k.emptyVirtualService()
	serviceName := kubeapiserverconstants.ServiceName(k.values.NamePrefix)
	host := kubernetesutils.FQDNForService(serviceName, k.namespace)

	vs.Labels = getLabels()
	vs.Spec = istioapinetworkingv1beta1.VirtualService{
		ExportTo: []string{"."},
		Hosts: []string{
			host,
			serviceName,
		},
		Tls: []*istioapinetworkingv1beta1.TLSRoute{{
			Match: []*istioapinetworkingv1beta1.TLSMatchAttributes{{
				Port: kubeapiserverconstants.Port,
				SniHosts: []string{
					serviceName,
					host,
				},
			}},
			Route: []*istioapinetworkingv1beta1.RouteDestination{{
				Destination: &istioapinetworkingv1beta1.Destination{
					Host: host,
					Port: &istioapinetworkingv1beta1.PortSelector{
						Number: kubeapiserverconstants.Port,
					},
				},
			}},
		}},
	}
	return vs
}

func (k *kubeAPIServer) reconcileDestinationRule(ctx context.Context, destinationRule *istionetworkingv1beta1.DestinationRule) error {
	desired := k.destinationRule()
	_, err := controllerutils.GetAndCreateOrMergePatch(ctx, k.client.Client(), destinationRule, func() error {
		destinationRule.Labels = desired.Labels
		destinationRule.Spec = desired.Spec
		return nil
	})
	return err
}

func (k *kubeAPIServer) reconcileVirtualService(ctx context.Context, virtualService *istionetworkingv1beta1.VirtualService) error {
	desired := k.virtualService()
	_, err := controllerutils.GetAndCreateOrMergePatch(ctx, k.client.Client(), virtualService, func() error {
		virtualService.Labels = desired.Labels
		virtualService.Spec = desired.Spec
		return nil
	})
	return err
}
