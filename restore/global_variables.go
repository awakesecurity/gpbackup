package restore

import (
	"sync"

	"github.com/greenplum-db/gp-common-go-libs/cluster"
	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gpbackup/backup_filepath"
	"github.com/greenplum-db/gpbackup/backup_history"
	"github.com/greenplum-db/gpbackup/utils"
	"github.com/spf13/pflag"
)

/*
 * This file contains global variables and setter functions for those variables
 * used in testing.
 */

/*
 * Non-flag variables
 */

var (
	backupConfig        *backup_history.BackupConfig
	connectionPool      *dbconn.DBConn
	globalCluster       *cluster.Cluster
	globalFPInfo        backup_filepath.FilePathInfo
	globalTOC           *utils.TOC
	pluginConfig        *utils.PluginConfig
	restoreStartTime    string
	version             string
	wasTerminated       bool
	errorTablesData     []string
	errorTablesMetadata []string

	/*
	 * Used for synchronizing DoCleanup.  In DoInit() we increment the group
	 * and then wait for at least one DoCleanup to finish, either in DoTeardown
	 * or the signal handler.
	 */
	CleanupGroup *sync.WaitGroup
)

/*
 * Command-line flags
 */
var cmdFlags *pflag.FlagSet

/*
 * Setter functions
 */

func SetCmdFlags(flagSet *pflag.FlagSet) {
	cmdFlags = flagSet
}

func SetBackupConfig(config *backup_history.BackupConfig) {
	backupConfig = config
}

func SetConnection(conn *dbconn.DBConn) {
	connectionPool = conn
}

func SetCluster(cluster *cluster.Cluster) {
	globalCluster = cluster
}

func SetFPInfo(fpInfo backup_filepath.FilePathInfo) {
	globalFPInfo = fpInfo
}

func SetPluginConfig(config *utils.PluginConfig) {
	pluginConfig = config
}

func SetTOC(toc *utils.TOC) {
	globalTOC = toc
}

// Util functions to enable ease of access to global flag values

func MustGetFlagString(flagName string) string {
	return utils.MustGetFlagString(cmdFlags, flagName)
}

func MustGetFlagInt(flagName string) int {
	return utils.MustGetFlagInt(cmdFlags, flagName)
}

func MustGetFlagBool(flagName string) bool {
	return utils.MustGetFlagBool(cmdFlags, flagName)
}

func MustGetFlagStringSlice(flagName string) []string {
	return utils.MustGetFlagStringSlice(cmdFlags, flagName)
}

func GetVersion() string {
	return version
}

func SetVersion(v string) {
	version = v
}
