// package postgresql

// import (
// 	"fmt"
// 	"os/exec"
// 	"strings"

// 	"github.com/martient/golang-utils/utils"
// )

// func RunBackup(database PostgresqlRequirements) error {
// 	if (database == PostgresqlRequirements{}) {
// 		return fmt.Errorf("database can't be empty")
// 	}

// 	var builderPASSWORD strings.Builder
// 	var builderPATH strings.Builder
// 	var builderHOSTNAME strings.Builder
// 	var builderPORT strings.Builder
// 	var builderUSER strings.Builder
// 	var builderNAME strings.Builder
// 	if database.Password != "" {
// 		builderPASSWORD.WriteString("PGPASSWORD=")
// 		builderPASSWORD.WriteString(database.Password)
// 		builderPASSWORD.WriteString(" ")
// 	}
// 	builderPATH.WriteString("pg_dump")

// 	if database.Hostname != "" {
// 		builderHOSTNAME.WriteString("-h ")
// 		builderHOSTNAME.WriteString(database.Hostname)
// 		builderHOSTNAME.WriteString(" ")
// 	}
// 	if database.Port != "" {
// 		builderPORT.WriteString("-p ")
// 		builderPORT.WriteString(database.Port)
// 		builderPORT.WriteString(" ")
// 	}
// 	if database.User != "" {
// 		builderUSER.WriteString("-U ")
// 		builderUSER.WriteString(database.User)
// 		builderUSER.WriteString(" ")
// 	}
// 	if database.Name != "" {
// 		builderNAME.WriteString(database.Name)
// 	}

// 	_, err := exec.LookPath("pg_dump")
// 	if err != nil {
// 		utils.LogError("pg_dump must be installed, please install it", "POSTGRESQL", err)
// 		return err
// 	}

// utils.LogInfo("CLI cmd: %s %s %s %s %s %s", "POSTGRESQL", builderPASSWORD.String(), builderPATH.String(), builderHOSTNAME.String(), builderPORT.String(), builderUSER.String(), builderNAME.String())
// output, err := exec.Command(builderPASSWORD.String(), builderPATH.String(), builderHOSTNAME.String(), builderPORT.String(), builderUSER.String(), builderNAME.String()).CombinedOutput()
// if err != nil {
package postgresql

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/martient/golang-utils/utils"
)

const pgDumpCommand = "pg_dump"

func RunBackup(database PostgresqlRequirements) (*bytes.Buffer, error) {
	if err := validateRequirements(database); err != nil {
		return nil, err
	}

	pgDumpPath, err := exec.LookPath(pgDumpCommand)
	if err != nil {
		return nil, fmt.Errorf("pg_dump command not found: %w", err)
	}

	args := buildCommandArgs(database)
	cmd := exec.Command(pgDumpPath, args...)
	utils.LogInfo(cmd.String(), "POSTGRESQL")

	var buffer bytes.Buffer
	cmd.Stdout = &buffer

	err = cmd.Run()
	if err != nil {
		utils.LogError("Failed to backup database '%s'", "POSTGRESQL", err)
		return nil, fmt.Errorf("backup failed: %w", err)
	}
	return &buffer, nil
}

func validateRequirements(database PostgresqlRequirements) error {
	if database == (PostgresqlRequirements{}) {
		return fmt.Errorf("database requirements cannot be empty")
	}

	// Add additional validation checks for port, hostname, etc.

	return nil
}

func buildCommandArgs(database PostgresqlRequirements) []string {
	var args []string

	// if database.Password != "" {
	// 	args = append(args, "PGPASSWORD="+database.Password)
	// }

	if database.Hostname != "" {
		args = append(args, "-h", database.Hostname)
	}

	if database.Port != "" {
		args = append(args, "-p", database.Port)
	}

	if database.User != "" {
		args = append(args, "-U", database.User)
	}

	args = append(args, database.Name)

	return args
}

// 		fmt.Println(err)
// 		fmt.Println(string(output))
// 		utils.LogError("Something went wrong during the execution", "POSTGRESQL", err)
// 		return err
// 	}
// 	fmt.Println(output)
// 	return nil
// }
