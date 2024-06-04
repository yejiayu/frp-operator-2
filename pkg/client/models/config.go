package models

import (
	"context"

	corev1 "k8s.io/api/core/v1"

	frpv1alpha1 "github.com/zufardhiyaulhaq/frp-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServerAuthenticationType int64

const DEFAULT_ADMIN_ADDRESS = "0.0.0.0"
const DEFAULT_ADMIN_PORT = 7400
const DEFAULT_ADMIN_USERNAME = "frpc-user"
const DEFAULT_ADMIN_PASSWORD = "frpc-password"

const (
	Token ServerAuthenticationType = iota
)

type Config struct {
	ServerAddress        string
	ServerPort           int
	ServerAuthentication ServerAuthentication
	AdminAddress         string
	AdminPort            int
	AdminUsername        string
	AdminPassword        string

	Upstreams []Upstream
}

type ServerAuthentication struct {
	Type  ServerAuthenticationType
	Token string
}

type UpstreamType int64

const (
	TCP UpstreamType = iota
	UDP UpstreamType = iota
)

type Upstream struct {
	Name string
	Type UpstreamType
	TCP  Upstream_TCP
	UDP  Upstream_UDP
}

type Upstream_TCP struct {
	Host          string
	Port          int
	ServerPort    int
	ProxyProtocol *string
	HealthCheck   *Upstream_TCP_HealthCheck
}

type Upstream_TCP_HealthCheck struct {
	TimeoutSeconds  int
	MaxFailed       int
	IntervalSeconds int
}

type Upstream_UDP struct {
	Host       string
	Port       int
	ServerPort int
}

func NewConfig(k8sclient client.Client, clientObject *frpv1alpha1.Client, upstreamObjects []frpv1alpha1.Upstream) (Config, error) {
	config := Config{
		ServerAddress: clientObject.Spec.Server.Host,
			ServerPort:    clientObject.Spec.Server.Port,
			AdminAddress:  DEFAULT_ADMIN_ADDRESS,
			AdminPort:     DEFAULT_ADMIN_PORT,
			AdminUsername: DEFAULT_ADMIN_USERNAME,
			AdminPassword: DEFAULT_ADMIN_PASSWORD,
	}

	if clientObject.Spec.Server.Authentication.Token != nil {
		config.ServerAuthentication.Type = 1

		secret := &corev1.Secret{}
		err := k8sclient.Get(context.TODO(), types.NamespacedName{Name: clientObject.Spec.Server.Authentication.Token.Secret.Name, Namespace: clientObject.Namespace}, secret)
		if err != nil && errors.IsNotFound(err) {
			return config, err
		} else if err != nil {
			return config, err
		}

		tokenByte, ok := secret.Data[clientObject.Spec.Server.Authentication.Token.Secret.Key]
		if !ok {
			return config, err
		}

		config.ServerAuthentication.Token = string(tokenByte)
	}

	upstreams := []Upstream{}
	for _, upstreamObject := range upstreamObjects {
		upstream := Upstream{
			Name: upstreamObject.Name,
		}

		if upstreamObject.Spec.TCP == nil && upstreamObject.Spec.UDP == nil {
			return config, errors.NewBadRequest("TCP or UDP upstream is required")
		}

		if upstreamObject.Spec.TCP != nil && upstreamObject.Spec.UDP != nil {
			return config, errors.NewBadRequest("Multiple protocol on the same Upstream object")
		}

		if upstreamObject.Spec.TCP != nil {
			upstream.Type = 1
			upstream.TCP.Host = upstreamObject.Spec.TCP.Host
			upstream.TCP.Port = upstreamObject.Spec.TCP.Port
			upstream.TCP.ServerPort = upstreamObject.Spec.TCP.Server.Port

			if upstreamObject.Spec.TCP.ProxyProtocol != nil {
				upstream.TCP.ProxyProtocol = upstreamObject.Spec.TCP.ProxyProtocol
			}

			if upstreamObject.Spec.TCP.HealthCheck != nil {
				upstream.TCP.HealthCheck.IntervalSeconds = upstreamObject.Spec.TCP.HealthCheck.IntervalSeconds
				upstream.TCP.HealthCheck.MaxFailed = upstreamObject.Spec.TCP.HealthCheck.MaxFailed
				upstream.TCP.HealthCheck.TimeoutSeconds = upstreamObject.Spec.TCP.HealthCheck.TimeoutSeconds
			}
		}

		if upstreamObject.Spec.UDP != nil {
			upstream.Type = 2
			upstream.UDP.Host = upstreamObject.Spec.UDP.Host
			upstream.UDP.Port = upstreamObject.Spec.UDP.Port
			upstream.UDP.ServerPort = upstreamObject.Spec.UDP.Server.Port
		}

		upstreams = append(upstreams, upstream)
	}
	config.Upstreams = upstreams

	return config, nil
}
