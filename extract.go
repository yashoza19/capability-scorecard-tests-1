package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"main/pkg/actions"
	bundle "main/pkg/bundle"
	"main/pkg/models"
	"os"
	"path/filepath"
	"strings"

	apimanifests "github.com/operator-framework/api/pkg/manifests"
	"github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/audit/pkg"
	log "github.com/sirupsen/logrus"
)

const DefaultContainerTool = Docker
const Docker = "docker"
const Podman = "podman"

func GetDataFromIndexDB(report bundle.Data) (bundle.Data, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", "./output/index.db")
	if err != nil {
		return report, fmt.Errorf("unable to connect in to the database : %s", err)
	}

	sql, err := report.BuildBundlesQuery()
	if err != nil {
		return report, err
	}

	row, err := db.Query(sql)
	if err != nil {
		return report, fmt.Errorf("unable to query the index db : %s", err)
	}

	defer row.Close()
	for row.Next() {
		var bundleName string
		var csv *string
		var bundlePath string
		var csvStruct *v1alpha1.ClusterServiceVersion

		err = row.Scan(&bundleName, &bundlePath)
		if err != nil {
			log.Errorf("unable to scan data from index %s\n", err.Error())
		}
		log.Infof("Generating data from the bundle (%s)", bundleName)
		auditBundle := models.NewAuditBundle(bundleName, bundlePath)

		// the csv is pruned from the database to save space.
		// See that is store only what is needed to populate the package manifest on cluster, all the extra
		// manifests are pruned to save storage space
		if csv != nil {
			err = json.Unmarshal([]byte(*csv), &csvStruct)
			if err == nil {
				auditBundle.CSVFromIndexDB = csvStruct
			} else {
				auditBundle.Errors = append(auditBundle.Errors,
					fmt.Errorf("unable to parse the csv from the index.db: %s", err).Error())
			}
		}

		auditBundle = GetDataFromBundleImage(auditBundle, flags.ContainerEngine)

		sqlString := fmt.Sprintf("SELECT c.channel_name, c.package_name FROM channel_entry c "+
			"where c.operatorbundle_name = '%s'", auditBundle.OperatorBundleName)
		row, err := db.Query(sqlString)
		if err != nil {
			return report, fmt.Errorf("unable to query channel entry in the index db : %s", err)
		}

		defer row.Close()
		var channelName string
		var packageName string
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&channelName, &packageName)
			auditBundle.Channels = append(auditBundle.Channels, channelName)
			auditBundle.PackageName = packageName
		}

		if len(strings.TrimSpace(auditBundle.PackageName)) == 0 && auditBundle.Bundle != nil {
			auditBundle.PackageName = auditBundle.Bundle.Package
		}

		sqlString = fmt.Sprintf("SELECT default_channel FROM package WHERE name = '%s'", auditBundle.PackageName)
		row, err = db.Query(sqlString)
		if err != nil {
			return report, fmt.Errorf("unable to query default channel entry in the index db : %s", err)
		}

		defer row.Close()
		var defaultChannelName string
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&defaultChannelName)
			auditBundle.DefaultChannel = defaultChannelName
		}

		sqlString = fmt.Sprintf("SELECT type, value FROM properties WHERE operatorbundle_name = '%s'",
			auditBundle.OperatorBundleName)
		row, err = db.Query(sqlString)
		if err != nil {
			return report, fmt.Errorf("unable to query properties entry in the index db : %s", err)
		}

		defer row.Close()
		var properType string
		var properValue string
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&properType, &properValue)
			auditBundle.PropertiesDB = append(auditBundle.PropertiesDB,
				pkg.PropertiesAnnotation{Type: properType, Value: properValue})
		}

		sqlString = fmt.Sprintf("select count(*) from channel where head_operatorbundle_name = '%s'",
			auditBundle.OperatorBundleName)
		row, err = db.Query(sqlString)
		if err != nil {
			return report, fmt.Errorf("unable to query properties entry in the index db : %s", err)
		}

		defer row.Close()
		var found int
		for row.Next() { // Iterate and fetch the records from result cursor
			_ = row.Scan(&found)
			auditBundle.IsHeadOfChannel = found > 0
		}

		report.AuditBundle = append(report.AuditBundle, *auditBundle)
	}

	return report, nil
}

func GetDataFromBundleImage(auditBundle *models.AuditBundle,
	containerEngine string) *models.AuditBundle {

	if len(auditBundle.OperatorBundleImagePath) < 1 {
		auditBundle.Errors = append(auditBundle.Errors,
			errors.New("not found bundle path stored in the index.db").Error())
		return auditBundle
	}

	err := actions.DownloadImage(auditBundle.OperatorBundleImagePath, containerEngine)
	if err != nil {
		auditBundle.Errors = append(auditBundle.Errors,
			fmt.Errorf("unable to download container image (%s): %s", auditBundle.OperatorBundleImagePath, err).Error())
		return auditBundle
	}

	bundleDir := actions.CreateBundleDir(auditBundle)
	actions.ExtractBundleFromImage(auditBundle, bundleDir, containerEngine)

	inspectManifest, err := pkg.RunDockerInspect(auditBundle.OperatorBundleImagePath, containerEngine)
	if err != nil {
		auditBundle.Errors = append(auditBundle.Errors, err.Error())
	} else {
		// Gathering data by inspecting the operator bundle image
		auditBundle.BundleImageLabels = inspectManifest.DockerConfig.Labels

	}

	// Read the bundle
	auditBundle.Bundle, err = apimanifests.GetBundleFromDir(filepath.Join(bundleDir, "bundle"))
	if err != nil {
		auditBundle.Errors = append(auditBundle.Errors, fmt.Errorf("unable to get the bundle: %s", err).Error())
		return auditBundle
	}

	return auditBundle
}

func GetContainerToolFromEnvVar() string {
	if value, ok := os.LookupEnv("CONTAINER_ENGINE"); ok {
		return value
	}
	return DefaultContainerTool
}
