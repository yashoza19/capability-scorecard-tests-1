package bundle

import (
	"fmt"
	"main/pkg/models"

	sq "github.com/Masterminds/squirrel"
)

type Data struct {
	AuditBundle       []models.AuditBundle
	Flags             BindFlags
	IndexImageInspect DockerInspect
}

type DockerConfig struct {
	Labels map[string]string `json:"Labels"`
}

type DockerInspect struct {
	ID           string       `json:"ID"`
	RepoDigests  []string     `json:"RepoDigests"`
	Created      string       `json:"Created"`
	DockerConfig DockerConfig `json:"Config"`
}

func (d *Data) BuildBundlesQuery() (string, error) {
	query := sq.Select("o.name, o.csv, o.bundlepath").From(
		"operatorbundle o")

	query.OrderBy("o.name")

	sql, _, err := query.ToSql()
	if err != nil {
		return "", fmt.Errorf("unable to create sql : %s", err)
	}
	return sql, nil
}
