package models

import (
	"github.com/operator-framework/api/pkg/apis/scorecard/v1alpha3"
	apimanifests "github.com/operator-framework/api/pkg/manifests"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/api/pkg/validation/errors"
	"github.com/operator-framework/audit/pkg"
)

type AuditBundle struct {
	Bundle                  *apimanifests.Bundle
	FoundLabel              bool
	OperatorBundleName      string
	OperatorBundleImagePath string
	PackageName             string
	DefaultChannel          string
	ScorecardResults        v1alpha3.TestList
	ValidatorsResults       []errors.ManifestResult
	CSVFromIndexDB          *v1alpha1.ClusterServiceVersion
	PropertiesDB            []pkg.PropertiesAnnotation
	Channels                []string
	HasCustomScorecardTests bool
	IsHeadOfChannel         bool
	BundleImageLabels       map[string]string `json:"bundleImageLabels,omitempty"`
	BundleAnnotations       map[string]string `json:"bundleAnnotations,omitempty"`
	Errors                  []string
}

func NewAuditBundle(operatorBundleName, operatorBundleImagePath string) *AuditBundle {
	auditBundle := AuditBundle{}
	auditBundle.OperatorBundleName = operatorBundleName
	auditBundle.OperatorBundleImagePath = operatorBundleImagePath
	return &auditBundle
}
