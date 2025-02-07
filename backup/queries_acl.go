package backup

/*
 * This file contains structs and functions related to executing specific
 * queries to gather metadata for the objects handled in predata_shared.go.
 */

import (
	"database/sql"
	"fmt"
	"sort"

	"github.com/greenplum-db/gp-common-go-libs/dbconn"
	"github.com/greenplum-db/gp-common-go-libs/gplog"
	"github.com/greenplum-db/gpbackup/utils"
)

/*
 * Structs and functions relating to generic metadata handling.
 */

type MetadataQueryParams struct {
	NameField    string
	SchemaField  string
	OidField     string
	ACLField     string
	OwnerField   string
	OidTable     string
	CommentTable string
	CatalogTable string
	Shared       bool
}

var (
	TYPE_AGGREGATE          MetadataQueryParams
	TYPE_CAST               MetadataQueryParams
	TYPE_COLLATION          MetadataQueryParams
	TYPE_CONSTRAINT         MetadataQueryParams
	TYPE_CONVERSION         MetadataQueryParams
	TYPE_DATABASE           MetadataQueryParams
	TYPE_EVENTTRIGGER       MetadataQueryParams
	TYPE_EXTENSION          MetadataQueryParams
	TYPE_FOREIGNDATAWRAPPER MetadataQueryParams
	TYPE_FOREIGNSERVER      MetadataQueryParams
	TYPE_FUNCTION           MetadataQueryParams
	TYPE_INDEX              MetadataQueryParams
	TYPE_PROCLANGUAGE       MetadataQueryParams
	TYPE_OPERATOR           MetadataQueryParams
	TYPE_OPERATORCLASS      MetadataQueryParams
	TYPE_OPERATORFAMILY     MetadataQueryParams
	TYPE_PROTOCOL           MetadataQueryParams
	TYPE_RELATION           MetadataQueryParams
	TYPE_RESOURCEGROUP      MetadataQueryParams
	TYPE_RESOURCEQUEUE      MetadataQueryParams
	TYPE_ROLE               MetadataQueryParams
	TYPE_RULE               MetadataQueryParams
	TYPE_SCHEMA             MetadataQueryParams
	TYPE_TABLESPACE         MetadataQueryParams
	TYPE_TSCONFIGURATION    MetadataQueryParams
	TYPE_TSDICTIONARY       MetadataQueryParams
	TYPE_TSPARSER           MetadataQueryParams
	TYPE_TSTEMPLATE         MetadataQueryParams
	TYPE_TRIGGER            MetadataQueryParams
	TYPE_TYPE               MetadataQueryParams
)

func InitializeMetadataParams(connectionPool *dbconn.DBConn) {
	TYPE_AGGREGATE = MetadataQueryParams{NameField: "proname", SchemaField: "pronamespace", ACLField: "proacl", OwnerField: "proowner", CatalogTable: "pg_proc"}
	TYPE_CAST = MetadataQueryParams{NameField: "typname", OidField: "oid", OidTable: "pg_type", CatalogTable: "pg_cast"}
	TYPE_COLLATION = MetadataQueryParams{NameField: "collname", OidField: "oid", SchemaField: "collnamespace", OwnerField: "collowner", CatalogTable: "pg_collation"}
	TYPE_CONSTRAINT = MetadataQueryParams{NameField: "conname", SchemaField: "connamespace", OidField: "oid", CatalogTable: "pg_constraint"}
	TYPE_CONVERSION = MetadataQueryParams{NameField: "conname", OidField: "oid", SchemaField: "connamespace", OwnerField: "conowner", CatalogTable: "pg_conversion"}
	TYPE_DATABASE = MetadataQueryParams{NameField: "datname", ACLField: "datacl", OwnerField: "datdba", CatalogTable: "pg_database", Shared: true}
	TYPE_EVENTTRIGGER = MetadataQueryParams{NameField: "evtname", OidField: "oid", OwnerField: "evtowner", CatalogTable: "pg_event_trigger"}
	TYPE_EXTENSION = MetadataQueryParams{NameField: "extname", OidField: "oid", CatalogTable: "pg_extension"}
	TYPE_FOREIGNDATAWRAPPER = MetadataQueryParams{NameField: "fdwname", ACLField: "fdwacl", OwnerField: "fdwowner", CatalogTable: "pg_foreign_data_wrapper"}
	TYPE_FOREIGNSERVER = MetadataQueryParams{NameField: "srvname", ACLField: "srvacl", OwnerField: "srvowner", CatalogTable: "pg_foreign_server"}
	TYPE_FUNCTION = TYPE_AGGREGATE // Aggregates are functions. So the metadata call to get them are the same.
	TYPE_INDEX = MetadataQueryParams{NameField: "relname", OidField: "indexrelid", OidTable: "pg_class", CommentTable: "pg_class", CatalogTable: "pg_index"}
	TYPE_PROCLANGUAGE = MetadataQueryParams{NameField: "lanname", ACLField: "lanacl", CatalogTable: "pg_language"}
	if connectionPool.Version.Before("5") {
		TYPE_PROCLANGUAGE.OwnerField = "10" // In GPDB 4.3, there is no lanowner field in pg_language, but languages have an implicit owner
	} else {
		TYPE_PROCLANGUAGE.OwnerField = "lanowner"
	}
	TYPE_OPERATOR = MetadataQueryParams{NameField: "oprname", SchemaField: "oprnamespace", OidField: "oid", OwnerField: "oprowner", CatalogTable: "pg_operator"}
	TYPE_OPERATORCLASS = MetadataQueryParams{NameField: "opcname", SchemaField: "opcnamespace", OidField: "oid", OwnerField: "opcowner", CatalogTable: "pg_opclass"}
	TYPE_OPERATORFAMILY = MetadataQueryParams{NameField: "opfname", SchemaField: "opfnamespace", OidField: "oid", OwnerField: "opfowner", CatalogTable: "pg_opfamily"}
	TYPE_PROTOCOL = MetadataQueryParams{NameField: "ptcname", ACLField: "ptcacl", OwnerField: "ptcowner", CatalogTable: "pg_extprotocol"}
	TYPE_RELATION = MetadataQueryParams{NameField: "relname", SchemaField: "relnamespace", ACLField: "relacl", OwnerField: "relowner", CatalogTable: "pg_class"}
	TYPE_RESOURCEGROUP = MetadataQueryParams{NameField: "rsgname", OidField: "oid", CatalogTable: "pg_resgroup", Shared: true}
	TYPE_RESOURCEQUEUE = MetadataQueryParams{NameField: "rsqname", OidField: "oid", CatalogTable: "pg_resqueue", Shared: true}
	TYPE_ROLE = MetadataQueryParams{NameField: "rolname", OidField: "oid", CatalogTable: "pg_authid", Shared: true}
	TYPE_RULE = MetadataQueryParams{NameField: "rulename", OidField: "oid", CatalogTable: "pg_rewrite"}
	TYPE_SCHEMA = MetadataQueryParams{NameField: "nspname", ACLField: "nspacl", OwnerField: "nspowner", CatalogTable: "pg_namespace"}
	TYPE_TABLESPACE = MetadataQueryParams{NameField: "spcname", ACLField: "spcacl", OwnerField: "spcowner", CatalogTable: "pg_tablespace", Shared: true}
	TYPE_TSCONFIGURATION = MetadataQueryParams{NameField: "cfgname", OidField: "oid", SchemaField: "cfgnamespace", OwnerField: "cfgowner", CatalogTable: "pg_ts_config"}
	TYPE_TSDICTIONARY = MetadataQueryParams{NameField: "dictname", OidField: "oid", SchemaField: "dictnamespace", OwnerField: "dictowner", CatalogTable: "pg_ts_dict"}
	TYPE_TSPARSER = MetadataQueryParams{NameField: "prsname", OidField: "oid", SchemaField: "prsnamespace", CatalogTable: "pg_ts_parser"}
	TYPE_TSTEMPLATE = MetadataQueryParams{NameField: "tmplname", OidField: "oid", SchemaField: "tmplnamespace", CatalogTable: "pg_ts_template"}
	TYPE_TRIGGER = MetadataQueryParams{NameField: "tgname", OidField: "oid", CatalogTable: "pg_trigger"}
	TYPE_TYPE = MetadataQueryParams{NameField: "typname", SchemaField: "typnamespace", OwnerField: "typowner", CatalogTable: "pg_type"}
	if connectionPool.Version.AtLeast("6") {
		TYPE_TYPE.ACLField = "typacl"
	}
}

type MetadataQueryStruct struct {
	UniqueID
	Privileges            sql.NullString
	Kind                  string
	Owner                 string
	Comment               string
	SecurityLabel         string
	SecurityLabelProvider string
}

func GetMetadataForObjectType(connectionPool *dbconn.DBConn, params MetadataQueryParams) MetadataMap {
	gplog.Verbose("Getting metadata for object types")
	aclStr := "''"
	kindStr := "''"
	schemaStr := ""
	ownerStr := "''"

	if params.ACLField != "" {
		aclStr = fmt.Sprintf(`CASE
		WHEN %[1]s IS NULL OR array_upper(%[1]s, 1) = 0 THEN %[1]s[0]
		ELSE unnest(%[1]s)
	END`, params.ACLField)
		kindStr = fmt.Sprintf(`CASE
		WHEN %[1]s IS NULL THEN 'Default'
		WHEN array_upper(%[1]s, 1) = 0 THEN 'Empty'
		ELSE ''
	END`, params.ACLField)
	}

	if params.SchemaField != "" {
		schemaStr = fmt.Sprintf(`JOIN pg_namespace n ON o.%s = n.oid
WHERE %s`, params.SchemaField, SchemaFilterClause("n"))
	}
	descFunc := "pg_description"
	subidStr := " AND d.objsubid = 0"
	if params.Shared {
		descFunc = "pg_shdescription"
		subidStr = ""
	}
	if params.OwnerField != "" {
		ownerStr = fmt.Sprintf("quote_ident(pg_get_userbyid(%s))", params.OwnerField)
	}

	secCols := ""
	secStr := ""
	if connectionPool.Version.AtLeast("6") {
		secCols = "coalesce(sec.label,'') AS securitylabel, coalesce(sec.provider, '') AS securitylabelprovider,"
		secTable := "pg_seclabel"
		secSubidStr := " AND sec.objsubid = 0"
		if params.Shared {
			secTable = "pg_shseclabel"
			secSubidStr = ""
		}
		secStr = fmt.Sprintf("LEFT JOIN %s sec ON (sec.objoid = o.oid AND sec.classoid = '%s'::regclass%s)", secTable, params.CatalogTable, secSubidStr)
	}

	query := fmt.Sprintf(`
SELECT
	'%s'::regclass::oid AS classid,
	o.oid,
	%s AS privileges,
	%s AS kind,
	%s AS owner,
	%s
	coalesce(description,'') AS comment
FROM %s o LEFT JOIN %s d ON (d.objoid = o.oid AND d.classoid = '%s'::regclass%s)
%s
%s
AND o.oid NOT IN (SELECT objid FROM pg_depend WHERE deptype='e')
ORDER BY o.oid;
`, params.CatalogTable, aclStr, kindStr, ownerStr, secCols, params.CatalogTable, descFunc, params.CatalogTable, subidStr, secStr, schemaStr)

	results := make([]MetadataQueryStruct, 0)
	err := connectionPool.Select(&results, query)
	gplog.FatalOnError(err)

	return ConstructMetadataMap(results)
}

func sortACLs(privileges []ACL) []ACL {
	sort.Slice(privileges, func(i, j int) bool {
		return privileges[i].Grantee < privileges[j].Grantee
	})
	return privileges
}

func GetCommentsForObjectType(connectionPool *dbconn.DBConn, params MetadataQueryParams) MetadataMap {
	schemaStr := ""
	if params.SchemaField != "" {
		schemaStr = fmt.Sprintf(` JOIN pg_namespace n ON o.%s = n.oid
	 WHERE %s`, params.SchemaField, SchemaFilterClause("n"))
	}
	descTable := "pg_description"
	subidStr := " AND d.objsubid = 0"
	if params.Shared {
		descTable = "pg_shdescription"
		subidStr = ""
	}
	commentTable := params.CatalogTable
	if params.CommentTable != "" {
		commentTable = params.CommentTable
	}
	query := fmt.Sprintf(`
SELECT
	'%s'::regclass::oid AS classid,
	o.%s AS oid,
	coalesce(description,'') AS comment
	FROM %s o JOIN %s d ON (d.objoid = %s AND d.classoid = '%s'::regclass%s)%s;
`, params.CatalogTable, params.OidField, params.CatalogTable, descTable, params.OidField, commentTable, subidStr, schemaStr)

	results := make([]struct {
		UniqueID
		Comment string
	}, 0)
	err := connectionPool.Select(&results, query)
	gplog.FatalOnError(err)

	metadataMap := make(MetadataMap)
	if len(results) > 0 {
		for _, result := range results {
			metadataMap[result.UniqueID] = ObjectMetadata{[]ACL{}, "", result.Comment, "", ""}
		}
	}
	return metadataMap
}

type DefaultPrivilegesQueryStruct struct {
	Oid        uint32
	Owner      string
	Schema     string
	Privileges sql.NullString
	Kind       string
	ObjectType string
}

type DefaultPrivileges struct {
	Owner      string
	Schema     string
	Privileges []ACL
	ObjectType string
}

func (dp DefaultPrivileges) GetMetadataEntry() (string, utils.MetadataEntry) {
	return "postdata",
		utils.MetadataEntry{
			Schema:          dp.Schema,
			Name:            "",
			ObjectType:      "DEFAULT PRIVILEGES",
			ReferenceObject: "",
			StartByte:       0,
			EndByte:         0,
		}
}

func GetDefaultPrivileges(connectionPool *dbconn.DBConn) []DefaultPrivileges {
	query := `SELECT
	a.oid,
	quote_ident(r.rolname) AS owner,
	coalesce(quote_ident(n.nspname),'') AS schema,
	CASE
		WHEN a.defaclacl IS NULL OR array_upper(a.defaclacl, 1) = 0 THEN a.defaclacl[0]
		ELSE unnest(a.defaclacl)
	END AS privileges,
	CASE
		WHEN a.defaclacl IS NULL THEN 'Default'
		WHEN array_upper(a.defaclacl, 1) = 0 THEN 'Empty'
		ELSE ''
	END AS kind,
	a.defaclobjtype AS objecttype
FROM pg_default_acl a
JOIN pg_roles r ON r.oid = a.defaclrole
LEFT JOIN pg_namespace n ON n.oid = a.defaclnamespace
ORDER BY n.nspname, a.defaclobjtype, r.rolname`
	results := make([]DefaultPrivilegesQueryStruct, 0)
	err := connectionPool.Select(&results, query)
	gplog.FatalOnError(err)

	return ConstructDefaultPrivileges(results)
}

func GetQuotedRoleNames(connectionPool *dbconn.DBConn) map[string]string {
	quotedRoleNames := make(map[string]string)

	quotedRoleNamesQuery := `
SELECT rolname AS rolename, quote_ident(rolname) AS quotedrolename FROM pg_authid`
	results := make([]struct {
		RoleName       string
		QuotedRoleName string
	}, 0)
	err := connectionPool.Select(&results, quotedRoleNamesQuery)
	gplog.FatalOnError(err)
	for _, result := range results {
		quotedRoleNames[result.RoleName] = result.QuotedRoleName
	}

	return quotedRoleNames
}
