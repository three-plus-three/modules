package web_ext

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/revel/revel"
	mylog "github.com/runner-mei/log"
	_ "github.com/three-plus-three/modules/bind"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/menus"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/types"
	"github.com/three-plus-three/modules/urlutil"
	"github.com/three-plus-three/sessions"
	sso "github.com/three-plus-three/sso/client"
)

var lifecycleData *Lifecycle

func init() {
	revel.TimeZone = time.Local
}

func Init(env *environment.Environment, serviceID environment.ENV_PROXY_TYPE, projectTitle string,
	cb func(*Lifecycle) error, createMenuList func(*Lifecycle) ([]toolbox.Menu, error)) {

	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.HTTPMethodOverride,
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		SessionFilter,                 // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		GlobalVariablesFilter,         // set global variables
		revel.ActionInvoker,           // Invoke the action.
	}

	revel.OnAppStart(func() {
		if env == nil {
			initEnv, err := environment.NewEnvironment(environment.Options{CurrentApplication: serviceID})
			if nil != err {
				log.Println(err)
				os.Exit(-1)
				return
			}
			env = initEnv
		}

		if revel.RunMode == "test" {
			env.Db.Models.Schema = env.Db.Models.Schema + "_models_test"
			env.Db.Data.Schema = env.Db.Data.Schema + "_test"
		}

		lifecycle, err := NewLifecycle(env, serviceID)
		if nil != err {
			log.Println(err)
			os.Exit(-1)
			return
		}

		appSo := env.GetServiceConfig(serviceID)
		//wserviceObject := env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID)
		if !revel.DevMode {
			if fp := flag.Lookup("port"); nil != fp {
				if fp.Value.String() == fp.DefValue {
					revel.HTTPPort, _ = strconv.Atoi(appSo.Port)
					revel.ServerEngineInit.Port = revel.HTTPPort
					revel.ServerEngineInit.Network, revel.ServerEngineInit.Address = appSo.ListenAddr("", "")
				}
			} else {
				revel.HTTPPort, _ = strconv.Atoi(appSo.Port)
				revel.ServerEngineInit.Port = revel.HTTPPort
				revel.ServerEngineInit.Network, revel.ServerEngineInit.Address = appSo.ListenAddr("", "")
			}
		} else {
			appSo.Port = "9000"
		}

		projectTitle := revel.Config.StringDefault("app.simpletitle", projectTitle)
		if s := env.Config.StringWithDefault(appSo.Name+".app.simpletitle", ""); s != "" {
			projectTitle = s
		}

		projectContext := appSo.Name
		lifecycle.URLPrefix = env.DaemonUrlPath
		lifecycle.URLRoot = env.DaemonUrlPath
		lifecycle.ApplicationContext = urlutil.Join(env.DaemonUrlPath, projectContext)
		lifecycle.ApplicationRoot = urlutil.Join(env.DaemonUrlPath, projectContext)

		lifecycle.Variables = ReadVariables(env, projectTitle)

		lifecycle.Variables["urlPrefix"] = lifecycle.URLPrefix
		lifecycle.Variables["url_prefix"] = lifecycle.URLPrefix
		lifecycle.Variables["urlRoot"] = lifecycle.URLRoot
		lifecycle.Variables["url_root"] = lifecycle.URLRoot

		lifecycle.Variables["application_context"] = lifecycle.ApplicationContext
		lifecycle.Variables["application_root"] = lifecycle.ApplicationRoot

		lifecycle.Variables["user_logout_url"] = urlutil.Join(env.DaemonUrlPath, "sso/logout")
		lifecycle.Variables["managed_objects_url"] = urlutil.Join(env.DaemonUrlPath, "web/layouts/stat")
		lifecycle.Variables["backgroud_tasks_url"] = urlutil.Join(env.DaemonUrlPath, "mc")
		//lifecycle.Variables["alert_stat_new_url"] = urlutil.Join(env.DaemonUrlPath, "web/notifications")
		lifecycle.Variables["alert_stat_new_url"] = urlutil.Join(env.DaemonUrlPath, "web/alert_events/stat_new")

		var constants map[string]interface{}
		if o := lifecycle.Variables["constants"]; o == nil {
			constants = map[string]interface{}{}
			lifecycle.Variables["constants"] = constants
		} else if constants, _ = o.(map[string]interface{}); constants == nil {
			log.Fatalln(fmt.Errorf("lifecycle.Variables[constants] isnot map[string]interface{}, got %T", o))
			os.Exit(-1)
			return
		}
		constants["user_admin"] = toolbox.UserAdmin
		constants["user_guest"] = toolbox.UserGuest
		constants["user_tpt_nm"] = toolbox.UserTPTNetwork

		if revel.DevMode {
			lifecycle.ModelEngine.ShowSQL()
			lifecycle.DataEngine.ShowSQL()
		}

		lifecycleData = lifecycle

		revel.AppRoot = lifecycle.ApplicationRoot

		//revel.Config.SetOption("app.secret", Env.Config.StringWithDefault("app.secret", ""))
		revel.Config.SetOption("cookie.prefix", "PLAY")
		revel.Config.SetOption("cookie.path", env.RawDaemonUrlPath)
		revel.CookiePrefix = "PLAY"

		var secretKey []byte
		if secretStr := env.Config.StringWithDefault("app.secret", ""); secretStr != "" {
			secretKey = []byte(secretStr)
		}
		cookiesPath := env.RawDaemonUrlPath
		if !strings.HasPrefix(cookiesPath, "/") {
			cookiesPath = "/" + cookiesPath
		}
		// Play 中我们有删除后面的 '/', 所以我们这里也删除一下
		if strings.HasSuffix(cookiesPath, "/") {
			cookiesPath = strings.TrimSuffix(cookiesPath, "/")
		}

		GlobalSessionFilter = sessions.SessionFilter(sso.DefaultSessionKey,
			cookiesPath, sha1.New, secretKey)

		lifecycle.UserManager = InitUser(lifecycle)
		lifecycle.GetUser = func(username string, opts ...toolbox.UserOption) toolbox.User {
			u, err := lifecycle.UserManager.ByName(username, opts...)
			if err != nil {
				if errors.IsNotFound(err) {
					return nil
				}
				panic(err)
			}
			return u
		}
		lifecycle.CurrentUser = func(c *revel.Controller) toolbox.User {
			username := c.Session[sso.SESSION_USER_KEY]
			if username == "" {
				return nil
			}
			return lifecycle.GetUser(username)
		}
		lifecycle.CheckUser = initSSO(env)

		initTemplateFuncs(lifecycle)

		applicationEnabled := revel.Config.StringDefault("hengwei.menu.products", "enabled")
		if mode := env.Config.StringWithDefault(appSo.Name+".menu.products", ""); mode != "" {
			applicationEnabled = mode
		}
		if applicationEnabled == "enabled" {
			version := revel.Config.StringDefault("version", "1.0")
			icon := revel.Config.StringDefault("app.icon", "")
			if s := env.Config.StringWithDefault(appSo.Name+".app.icon", ""); s != "" {
				icon = s
			}
			classes := revel.Config.StringDefault("hengwei.menu.classes", "")
			if s := env.Config.StringWithDefault(appSo.Name+".menu.classes", ""); s != "" {
				classes = s
			}

			err = menus.UpdateProduct(lifecycle.Env,
				lifecycle.ApplicationID, version, projectTitle, icon, classes,
				lifecycleData.ModelEngine.DB().DB)
			if err != nil {
				log.Println("UpdataProduct", err)
				os.Exit(-1)
				return
			}
		}

		types.RegisterEnumerationProvider("sql", &types.DbProvider{DB: lifecycle.ModelEngine.DB().DB})
		types.RegisterEnumerationProvider("usernames", &toolbox.UserProvider{UM: lifecycle.UserManager})
		types.RegisterEnumerationProvider("usergroups", &toolbox.UsergroupProvider{UM: lifecycle.UserManager})

		if err := cb(lifecycle); err != nil {
			log.Println(err)
			os.Exit(-1)
			return
		}
	}, 0)

	revel.OnAppStart(func() {
		if env == nil {
			initEnv, err := environment.NewEnvironment(environment.Options{CurrentApplication: serviceID})
			if nil != err {
				log.Println(err)
				os.Exit(-1)
				return
			}
			env = initEnv
		}
		appSo := env.GetServiceConfig(serviceID)
		menuMode := revel.Config.StringDefault("hengwei.menu.mode", "")
		if mode := env.Config.StringWithDefault(appSo.Name+".menu.mode", ""); mode != "" {
			menuMode = mode
		}

		menuClient := menus.Connect(lifecycleData.Env,
			serviceID,
			menus.Callback(func() ([]toolbox.Menu, error) {
				return createMenuList(lifecycleData)
			}),
			menuMode,
			"menus.changed",
			urlutil.Join(lifecycleData.Env.DaemonUrlPath, "/menu/"),
			mylog.New(os.Stderr).Named("menu"))

		lifecycleData.OnClosing(menuClient)
		lifecycleData.menuClient = menuClient

		applicationEnabled := revel.Config.StringDefault("hengwei.menu.products", "enabled")
		if mode := env.Config.StringWithDefault(appSo.Name+".menu.products", ""); mode != "" {
			applicationEnabled = mode
		}
		if applicationEnabled == "enabled" {
			lifecycleData.menuHook = menus.ProductsWrap(lifecycleData.Env,
				lifecycleData.ApplicationID,
				lifecycleData.ModelEngine.DB().DB,
				menus.Callback(func() ([]toolbox.Menu, error) {
					return lifecycleData.menuClient.Read()
				}))
		}
	}, 2)
}

// TODO turn this into revel.HeaderFilter
// should probably also have a filter for CSRF
// not sure if it can go in the same filter or not
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	// Add some common security headers
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

var GlobalSessionFilter revel.Filter

func SessionFilter(c *revel.Controller, filterChain []revel.Filter) {
	//if GlobalSessionFilter != nil {
	GlobalSessionFilter(c, filterChain)
	//}
}

// GlobalVariablesFilter will set global variables
func GlobalVariablesFilter(c *revel.Controller, fc []revel.Filter) {
	// Make global vars available in templates as {{.global.xyz}}
	c.ViewArgs["global"] = lifecycleData.Variables

	fc[0](c, fc[1:]) // Execute the next filter stage.
}
