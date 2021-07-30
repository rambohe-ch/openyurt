package util

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	insecureListenAddr = "1.1.1.1:10264"
	secureListenAddr   = "1.1.1.1:10263"
)

func TestResolveProxyPortsAndMappings(t *testing.T) {
	testcases := map[string]struct {
		configMap    *corev1.ConfigMap
		expectResult struct {
			ports        []string
			portMappings map[string]string
			err          error
		}
	}{
		"setting ports on dnat-ports-pair": {
			configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "yurt-tunnel-server-cfg",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"dnat-ports-pair": "9100=10264",
				},
			},
			expectResult: struct {
				ports        []string
				portMappings map[string]string
				err          error
			}{
				ports: []string{"9100"},
				portMappings: map[string]string{
					"9100": insecureListenAddr,
				},
				err: nil,
			},
		},
		"setting ports on http-proxy-ports": {
			configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "yurt-tunnel-server-cfg",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"http-proxy-ports": "9100,9200",
				},
			},
			expectResult: struct {
				ports        []string
				portMappings map[string]string
				err          error
			}{
				ports: []string{"9100", "9200"},
				portMappings: map[string]string{
					"9100": insecureListenAddr,
					"9200": insecureListenAddr,
				},
				err: nil,
			},
		},
		"setting ports on https-proxy-ports": {
			configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "yurt-tunnel-server-cfg",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"https-proxy-ports": "9100,9200",
				},
			},
			expectResult: struct {
				ports        []string
				portMappings map[string]string
				err          error
			}{
				ports: []string{"9100", "9200"},
				portMappings: map[string]string{
					"9100": secureListenAddr,
					"9200": secureListenAddr,
				},
				err: nil,
			},
		},
		"setting ports on http-proxy-ports and https-proxy-ports": {
			configMap: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ConfigMap",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "yurt-tunnel-server-cfg",
					Namespace: "kube-system",
				},
				Data: map[string]string{
					"http-proxy-ports":  "9100,9200",
					"https-proxy-ports": "9300,9400",
				},
			},
			expectResult: struct {
				ports        []string
				portMappings map[string]string
				err          error
			}{
				ports: []string{"9100", "9200", "9300", "9400"},
				portMappings: map[string]string{
					"9100": insecureListenAddr,
					"9200": insecureListenAddr,
					"9300": secureListenAddr,
					"9400": secureListenAddr,
				},
				err: nil,
			},
		},
	}

	for k, tt := range testcases {
		t.Run(k, func(t *testing.T) {
			ports, portMappings, err := resolveProxyPortsAndMappings(tt.configMap, insecureListenAddr, secureListenAddr)
			if tt.expectResult.err != err {
				t.Errorf("expect error: %v, but got error: %v", tt.expectResult.err, err)
			}

			// check the ports
			if len(tt.expectResult.ports) != len(ports) {
				t.Errorf("expect %d ports, but got %d ports", len(tt.expectResult.ports), len(ports))
			}

			foundPort := false
			for i := range tt.expectResult.ports {
				foundPort = false
				for j := range ports {
					if tt.expectResult.ports[i] == ports[j] {
						foundPort = true
						break
					}
				}

				if !foundPort {
					t.Errorf("expect %v ports, but got ports %v", tt.expectResult.ports, ports)
					break
				}
			}

			for i := range ports {
				foundPort = false
				for j := range tt.expectResult.ports {
					if tt.expectResult.ports[j] == ports[i] {
						foundPort = true
						break
					}
				}

				if !foundPort {
					t.Errorf("expect %v ports, but got ports %v", tt.expectResult.ports, ports)
					break
				}
			}

			// check the portMappings
			if len(tt.expectResult.portMappings) != len(portMappings) {
				t.Errorf("expect port mappings %v, but got port mappings %v", tt.expectResult.portMappings, portMappings)
			}

			for port, v := range tt.expectResult.portMappings {
				if gotV, ok := portMappings[port]; !ok {
					t.Errorf("expect port %s, but not got port", k)
				} else if v != gotV {
					t.Errorf("port(%s): expect dst value %s, but got dst value %s", k, v, gotV)
				}
				delete(portMappings, port)
			}
		})
	}
}
