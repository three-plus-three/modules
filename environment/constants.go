package environment

type ENV_PROXY_TYPE int

// 服务的常量
const (
	ENV_REDIS_PROXY_ID ENV_PROXY_TYPE = iota
	ENV_SAMPLING_PROXY_ID
	ENV_POLL_PROXY_ID
	ENV_SCHD_PROXY_ID
	ENV_LCN_PROXY_ID
	ENV_IP_MGR_PROXY_ID
	ENV_DELAYED_JOB_PROXY_ID
	ENV_TERMINAL_PROXY_ID
	ENV_JBRIDGE_PROXY_ID
	ENV_JBRIDGE15_PROXY_ID
	ENV_REST_PROXY_ID
	ENV_WSERVER_PROXY_ID
	ENV_LUA_BRIDGE_PROXY_ID
	ENV_WEB_PROXY_ID
	ENV_LOGGING_PROXY_ID
	ENV_NFLOW1_PROXY_ID
	ENV_NFLOW2_PROXY_ID
	ENV_MC_PROXY_ID
	ENV_MC_DEV_PROXY_ID
	ENV_INFLUXDB_PROXY_ID
	ENV_INFLUXDB_ADM_PROXY_ID
	ENV_FORK_PROXY_ID
	ENV_IMS_PROXY_ID
	ENV_WSERVER_SSL_PROXY_ID
	ENV_CMDB_PROXY_ID
	ENV_ASSET_MANAGE_PROXY_ID
	ENV_NSM_PROXY_ID
	ENV_MINIO_PROXY_ID
	ENV_USER_MANAGE_PROXY_ID
	ENV_ITSM_PROXY_ID
	ENV_LOGANALYZER_PROXY_ID
	ENV_AM_RECORD_MGR_ID
	ENV_VPN_ID
	ENV_MODELS_PROXY_ID        // Deprecated
	ENV_SAMPLING_STUB_PROXY_ID // Deprecated
	ENV_TSDB_PROXY_ID          // Deprecated
	ENV_MAX_PROXY_ID

	ENV_MIN_PROXY_ID = ENV_REDIS_PROXY_ID
	ENV_ES_PROXY_ID  = ENV_LOGANALYZER_PROXY_ID // Deprecated
	ENV_DS_PROXY_ID  = ENV_MODELS_PROXY_ID      // Deprecated
	ENV_AM_PROXY_ID  = ENV_ASSET_MANAGE_PROXY_ID
	ENV_UM_PROXY_ID  = ENV_USER_MANAGE_PROXY_ID
)

func IsValidProxyID(id ENV_PROXY_TYPE) bool {
	return id >= ENV_MIN_PROXY_ID && id < ENV_MAX_PROXY_ID
}

// 服务的缺省配置
var (
	ServiceOptions = []ServiceOption{
		{ID: ENV_REDIS_PROXY_ID, Name: "redis", Host: "127.0.0.1", Port: "36379"},

		/////////// Deprecated
		{ID: ENV_MODELS_PROXY_ID, Name: "ds", Host: "127.0.0.1", Port: "37071"},
		{ID: ENV_TSDB_PROXY_ID, Name: "tsdb", Host: "127.0.0.1", Port: "37074"},
		{ID: ENV_SAMPLING_STUB_PROXY_ID, Name: "sampling_stub", Host: "127.0.0.1", Port: "37081"},
		///////////

		{ID: ENV_SAMPLING_PROXY_ID, Name: "sampling", Host: "127.0.0.1", Port: "37072"},
		{ID: ENV_POLL_PROXY_ID, Name: "poll", Host: "127.0.0.1", Port: "37073"},
		{ID: ENV_SCHD_PROXY_ID, Name: "schd", Host: "127.0.0.1", Port: "37075"},
		{ID: ENV_LCN_PROXY_ID, Name: "lcn", Host: "127.0.0.1", Port: "37076"},
		{ID: ENV_IP_MGR_PROXY_ID, Name: "ip_mgr", Host: "127.0.0.1", Port: "37077"},
		{ID: ENV_DELAYED_JOB_PROXY_ID, Name: "delayed_jobs", Host: "127.0.0.1", Port: "37078"},
		{ID: ENV_TERMINAL_PROXY_ID, Name: "terminal", Host: "127.0.0.1", Port: "37079"},
		{ID: ENV_JBRIDGE_PROXY_ID, Name: "jbridge", Host: "127.0.0.1", Port: "37080"},
		{ID: ENV_REST_PROXY_ID, Name: "rest", Host: "127.0.0.1", Port: "39301"},
		{ID: ENV_WSERVER_PROXY_ID, Name: "wserver", Host: "127.0.0.1", Port: "37070"},
		{ID: ENV_WSERVER_SSL_PROXY_ID, Name: "daemon_ssl", Host: "127.0.0.1", Port: "37090"},
		{ID: ENV_LUA_BRIDGE_PROXY_ID, Name: "lua_bridge", Host: "127.0.0.1", Port: "37082"},
		{ID: ENV_WEB_PROXY_ID, Name: "web", Host: "127.0.0.1", Port: "39000"},
		{ID: ENV_LOGGING_PROXY_ID, Name: "es", Host: "127.0.0.1", Port: "37083"},
		{ID: ENV_NFLOW1_PROXY_ID, Name: "nflow", Host: "127.0.0.1", Port: "37084"},
		{ID: ENV_MC_PROXY_ID, Name: "mc", Host: "127.0.0.1", Port: "37085"},
		{ID: ENV_MC_DEV_PROXY_ID, Name: "mc_dev", Host: "127.0.0.1", Port: "9000"},
		{ID: ENV_INFLUXDB_PROXY_ID, Name: "influxdb", Host: "127.0.0.1", Port: "37086"},
		{ID: ENV_INFLUXDB_ADM_PROXY_ID, Name: "influxdb_adm", Host: "127.0.0.1", Port: "39183"},
		{ID: ENV_FORK_PROXY_ID, Name: "fork", Host: "127.0.0.1", Port: "37087"},
		{ID: ENV_JBRIDGE15_PROXY_ID, Name: "jbridge15", Host: "127.0.0.1", Port: "37088"},
		{ID: ENV_IMS_PROXY_ID, Name: "ims", Host: "127.0.0.1", Port: "37089"},
		{ID: ENV_CMDB_PROXY_ID, Name: "cmdb", Host: "127.0.0.1", Port: "37091"},
		{ID: ENV_ASSET_MANAGE_PROXY_ID, Name: "am", Host: "127.0.0.1", Port: "37092"},
		{ID: ENV_NSM_PROXY_ID, Name: "nsm", Host: "127.0.0.1", Port: "37093"},
		{ID: ENV_MINIO_PROXY_ID, Name: "minio", Host: "127.0.0.1", Port: "37094"},
		{ID: ENV_UM_PROXY_ID, Name: "um", Host: "127.0.0.1", Port: "37095"},
		{ID: ENV_ITSM_PROXY_ID, Name: "itsm", Host: "127.0.0.1", Port: "37096"},
		{ID: ENV_VPN_ID, Name: "vpn_management", Host: "127.0.0.1", Port: "39001"},

		// {ID: ENV_ES_PROXY_ID, Name: "es_old", Host: "127.0.0.1", Port: "39300"},
		{ID: ENV_LOGANALYZER_PROXY_ID, Name: "loganalyzer", Host: "127.0.0.1", Port: "37097"},
		{ID: ENV_AM_RECORD_MGR_ID, Name: "record_mgr", Host: "127.0.0.1", Port: "37098"},
		{ID: ENV_NFLOW2_PROXY_ID, Name: "nflow2", Host: "127.0.0.1", Port: "37099"},
	}
)

type ServiceOption struct {
	ID   ENV_PROXY_TYPE
	Name string
	Host string
	Port string
	Path string
}
