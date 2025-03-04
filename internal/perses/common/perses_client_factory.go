package common

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/perses/perses/pkg/model/api/v1/secret"

	v1 "github.com/perses/perses/pkg/client/api/v1"
	clientConfig "github.com/perses/perses/pkg/client/config"
	"github.com/perses/perses/pkg/model/api/v1/common"

	persesv1alpha1 "github.com/perses/perses-operator/api/v1alpha1"
)

type PersesClientFactory interface {
	CreateClient(perses persesv1alpha1.Perses) (v1.ClientInterface, error)
}

type PersesClientFactoryWithConfig struct{}

func NewWithConfig() PersesClientFactory {
	return &PersesClientFactoryWithConfig{}
}

func (f *PersesClientFactoryWithConfig) CreateClient(perses persesv1alpha1.Perses) (v1.ClientInterface, error) {
	var urlStr string

	var httpProtocol = "http"
	if isTLSEnabled(&perses) {
		httpProtocol = "https"
	}

	serverURLFlag := flag.Lookup(PersesServerURLFlag)
	if serverURLFlag != nil && serverURLFlag.Value.String() != "" {
		urlStr = serverURLFlag.Value.String()
	} else {
		urlStr = fmt.Sprintf("%s://%s.%s.svc.cluster.local:%d", httpProtocol, perses.Name, perses.Namespace, perses.Spec.ContainerPort)
	}
	parsedURL, err := common.ParseURL(urlStr)
	if err != nil {
		return nil, err
	}

	config := clientConfig.RestConfigClient{
		URL: parsedURL,
	}

	if isTLSEnabled(&perses) {
		tlsConfig := &secret.TLSConfig{
			InsecureSkipVerify: perses.Spec.Client.TLS.InsecureSkipVerify,
			CAFile:             filepath.Join(caMountPath, perses.Spec.Client.TLS.CaCert.CertFile),
		}

		if hasTLSConfiguration(&perses) {
			tlsConfig.CertFile = filepath.Join(tlsCertMountPath, perses.Spec.Client.TLS.UserCert.CertFile)
			tlsConfig.KeyFile = filepath.Join(tlsCertMountPath, perses.Spec.Client.TLS.UserCert.CertKeyFile)
		}

		config.TLSConfig = tlsConfig
	}

	restClient, err := clientConfig.NewRESTClient(config)
	if err != nil {
		return nil, err
	}

	persesClient := v1.NewWithClient(restClient)

	return persesClient, nil
}

type PersesClientFactoryWithClient struct {
	client v1.ClientInterface
}

func NewWithClient(client v1.ClientInterface) PersesClientFactory {
	return &PersesClientFactoryWithClient{client: client}
}

func (f *PersesClientFactoryWithClient) CreateClient(config persesv1alpha1.Perses) (v1.ClientInterface, error) {
	return f.client, nil
}
