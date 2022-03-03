package actions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"main/pkg/models"

	"github.com/operator-framework/audit/pkg"
	log "github.com/sirupsen/logrus"
)

type Manifest struct {
	Config string
	Layers []string
}

func DownloadImage(image string, containerEngine string) error {
	log.Infof("Downloading image %s to audit...", image)
	cmd := exec.Command(containerEngine, "pull", image)
	_, err := pkg.RunCommand(cmd)
	// if found an error try again
	// Sometimes it faces issues to download the image
	if err != nil {
		log.Warnf("error %s faced to downlad the image. Let's try more one time.", err)
		cmd := exec.Command(containerEngine, "pull", image)
		_, err = pkg.RunCommand(cmd)
	}
	return err
}

func CreateBundleDir(auditBundle *models.AuditBundle) string {
	currentPath, err := os.Getwd()
	if err != nil {
		log.Info(err)
		os.Exit(1)
	}

	dir := fmt.Sprintf("%s/tmp/%s", currentPath, auditBundle.OperatorBundleName)
	cmd := exec.Command("mkdir", dir)
	_, err = pkg.RunCommand(cmd)
	if err != nil {
		auditBundle.Errors = append(auditBundle.Errors,
			fmt.Errorf("unable to create the dir for the bundle: %s", err).Error())
	}
	return dir
}

func ExtractBundleFromImage(auditBundle *models.AuditBundle, bundleDir string, containerEngine string) {
	// imageName := strings.Split(auditBundle.OperatorBundleImagePath, "@")[0]
	imageName := auditBundle.OperatorBundleImagePath
	tarPath := fmt.Sprintf("%s/%s.tar", bundleDir, auditBundle.OperatorBundleName)
	cmd := exec.Command(containerEngine, "save", imageName, "-o", tarPath)
	_, err := pkg.RunCommand(cmd)
	if err != nil {
		log.Errorf("unable to save the bundle image : %s", err)
		auditBundle.Errors = append(auditBundle.Errors,
			fmt.Errorf("unable to save the bundle image : %s", err).Error())
	}

	cmd = exec.Command("tar", "-xvf", tarPath, "-C", bundleDir)
	_, err = pkg.RunCommand(cmd)
	if err != nil {
		log.Errorf("unable to untar the bundle image: %s", err)
		auditBundle.Errors = append(auditBundle.Errors,
			fmt.Errorf("unable to untar the bundle image : %s", err).Error())
	}

	cmd = exec.Command("mkdir", filepath.Join(bundleDir, "bundle"))
	_, err = pkg.RunCommand(cmd)
	if err != nil {
		log.Errorf("error to create the bundle bundleDir: %s", err)
		auditBundle.Errors = append(auditBundle.Errors,
			fmt.Errorf("error to create the bundle bundleDir : %s", err).Error())
	}

	bundleConfigFilePath := filepath.Join(bundleDir, "manifest.json")
	existingFile, err := ioutil.ReadFile(bundleConfigFilePath)
	if err == nil {
		var bundleLayerConfig []Manifest
		if err := json.Unmarshal(existingFile, &bundleLayerConfig); err != nil {
			log.Errorf("unable to Unmarshal manifest.json: %s", err)
			auditBundle.Errors = append(auditBundle.Errors,
				fmt.Errorf("unable to Unmarshal manifest.json: %s", err).Error())
		}
		if bundleLayerConfig == nil {
			log.Errorf("error to untar layers")
			auditBundle.Errors = append(auditBundle.Errors,
				fmt.Errorf("error to untar layers: %s", err).Error())
		}

		for _, layer := range bundleLayerConfig[0].Layers {
			cmd = exec.Command("tar", "-xvf", filepath.Join(bundleDir, layer), "-C", filepath.Join(bundleDir, "bundle"))
			_, err = pkg.RunCommand(cmd)
			if err != nil {
				log.Errorf("unable to untar layer : %s", err)
				auditBundle.Errors = append(auditBundle.Errors,
					fmt.Errorf("error to untar layers : %s", err).Error())
			}
		}
	} else {
		// If the docker manifest was not found then check if has just one layer
		cmd = exec.Command("tar", "-xvf", fmt.Sprintf("%s/layer.tar", bundleDir), "-C", filepath.Join(bundleDir, "bundle"))
		_, err = pkg.RunCommand(cmd)
		if err != nil {
			log.Errorf("unable to untar layer : %s", err)
			auditBundle.Errors = append(auditBundle.Errors,
				fmt.Errorf("unable to untar layer: %s", err).Error())
		}
	}

	// Remove files in the image to allow load the bundle
	cmd = exec.Command("rm", "-rf", fmt.Sprintf("%s/bundle/manifests/.wh..wh..opq", bundleDir))
	_, _ = pkg.RunCommand(cmd)

	cmd = exec.Command("rm", "-rf", fmt.Sprintf("%s/bundle/metadata/.wh..wh..opq", bundleDir))
	_, _ = pkg.RunCommand(cmd)

	cmd = exec.Command("rm", "-rf", fmt.Sprintf("%s/bundle/root/", bundleDir))
	_, _ = pkg.RunCommand(cmd)

	cmd = exec.Command("rm", "-rf", fmt.Sprintf("%s/bundle/manifests/.DS_Store", bundleDir))
	_, _ = pkg.RunCommand(cmd)
}

func CleanupBundleDir(auditBundle *models.AuditBundle, dir string) {
	cmd := exec.Command("rm", "-rf", dir)
	_, _ = pkg.RunCommand(cmd)
}
