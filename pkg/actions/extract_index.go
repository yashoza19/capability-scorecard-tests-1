package actions

import (
	"fmt"
	"os/exec"

	"main"

	"github.com/operator-framework/audit/pkg"
	log "github.com/sirupsen/logrus"
)

const catalogIndex = "audit-catalog-index"

func ExtractIndexDB(image string, containerEngine string) error {
	log.Info("Extracting database...")
	// Remove image if exists already
	command := exec.Command(containerEngine, "rm", catalogIndex)
	_, _ = main.RunCommand(command)

	// Download the image
	command = exec.Command(containerEngine, "create", "--name", catalogIndex, image, "\"yes\"")
	_, err := pkg.RunCommand(command)
	if err != nil {
		return fmt.Errorf("unable to create container image %s : %s", image, err)
	}

	// Extract
	command = exec.Command(containerEngine, "cp", fmt.Sprintf("%s:/database/index.db", catalogIndex), "./output/")
	_, err = pkg.RunCommand(command)
	if err != nil {
		return fmt.Errorf("unable to extract the image for index.db %s : %s", image, err)
	}
	return nil
}
