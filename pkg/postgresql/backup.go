package postgresql

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/martient/golang-utils/utils"
)

func RunBackup(database PostgresqlRequirements) error {
	if (database == PostgresqlRequirements{}) {
		return fmt.Errorf("database can't be empty")
	}

	var builderPATH strings.Builder
	var builderPARAMS strings.Builder

	// if database.Password != "" {
	// 	builderPATH.WriteString("PGPASSWORD=")
	// 	builderPATH.WriteString(database.Password)
	// 	builderPATH.WriteString(" ")
	// }
	builderPATH.WriteString("pg_dump")

	if database.Hostname != "" {
		builderPARAMS.WriteString("-h ")
		builderPARAMS.WriteString(database.Hostname)
		builderPARAMS.WriteString(" ")
	}
	if database.Port != "" {
		builderPARAMS.WriteString("-p ")
		builderPARAMS.WriteString(database.Port)
		builderPARAMS.WriteString(" ")
	}
	if database.User != "" {
		builderPARAMS.WriteString("-U ")
		builderPARAMS.WriteString(database.User)
		builderPARAMS.WriteString(" ")
	}
	if database.Name != "" {
		builderPARAMS.WriteString(database.Name)
	}

	_, err := exec.LookPath("pg_dump")
	if err != nil {
		utils.LogError("pg_dump must be installed, please install it", "POSTGRESQL", err)
		return err
	}

	utils.LogInfo("CLI cmd: %s %s", "POSTGRESQL", builderPATH.String(), builderPARAMS.String())
	output, err := exec.Command(builderPATH.String(), builderPARAMS.String()).Output()
	if err != nil {
		fmt.Println(err)
		fmt.Println(output)
		utils.LogError("Something went wrong during the execution", "POSTGRESQL", err)
		return err
	}
	fmt.Println(output)
	return nil
}
