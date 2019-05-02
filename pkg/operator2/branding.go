package operator2

import (
	"fmt"

	"gopkg.in/yaml.v2"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"

	configv1 "github.com/openshift/api/config/v1"
	osinv1 "github.com/openshift/api/osin/v1"
)

func (c *authOperator) handleBrandingTemplates(oauthConfig *configv1.OAuth, syncData configSyncData) (*osinv1.OAuthTemplates, error) {
	var templates *osinv1.OAuthTemplates

	brand, err := c.getConsoleBranding()
	if err != nil {
		return nil, err
	}

	switch brand {
	case "ocp", "dedicated", "online", "azure":
		klog.Infof("console configured for platform %s - using ocp branding", brand)
		templates = &osinv1.OAuthTemplates{
			Login:             ocpBrandingLoginPath,
			ProviderSelection: ocpBrandingProviderPath,
			Error:             ocpBrandingErrorKey,
		}
	case "okd":
		klog.Infof("console configured with okd branding")
		templates = &osinv1.OAuthTemplates{}
	}

	if templates == nil && DEFAULT_BRAND == "ocp" {
		// Build-time ocp selection
		klog.Infof("using build-time ocp branding")
		templates = &osinv1.OAuthTemplates{
			Login:             ocpBrandingLoginPath,
			ProviderSelection: ocpBrandingProviderPath,
			Error:             ocpBrandingErrorKey,
		}
	}

	emptyTemplates := configv1.OAuthTemplates{}
	// User-configured overrides everything else, individually.
	if configTemplates := oauthConfig.Spec.Templates; configTemplates != emptyTemplates {

		if templates == nil {
			templates = &osinv1.OAuthTemplates{}
		}
		if len(configTemplates.Login.Name) > 0 {
			klog.Infof("override login templates")
			templates.Login = syncData.addTemplateSecret(configTemplates.Login, loginField, configv1.LoginTemplateKey)
		}
		if len(configTemplates.ProviderSelection.Name) > 0 {
			klog.Infof("override provider templates")
			templates.ProviderSelection = syncData.addTemplateSecret(configTemplates.ProviderSelection, providerSelectionField, configv1.ProviderSelectionTemplateKey)
		}
		if len(configTemplates.Error.Name) > 0 {
			klog.Infof("override error templates")
			templates.Error = syncData.addTemplateSecret(configTemplates.Error, errorField, configv1.ErrorsTemplateKey)
		}
	}

	return templates, nil
}

func (c *authOperator) handleOCPBrandingSecret() (*corev1.Secret, error) {
	return c.secrets.Secrets(targetNamespace).Get(ocpBrandingSecretName, metav1.GetOptions{})
}

func (c *authOperator) getConsoleBranding() (string, error) {
	cm, err := c.configMaps.ConfigMaps("openshift-console").Get("console-config", metav1.GetOptions{})
	if errors.IsNotFound(err) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("error getting console-config: %v", err)
	}

	config := ConsoleConfig{}
	err = yaml.Unmarshal([]byte(cm.Data["console-config.yaml"]), &config)
	if err != nil {
		return "", fmt.Errorf("error parsing console-config: %v", err)
	}

	return config.Branding, nil
}
