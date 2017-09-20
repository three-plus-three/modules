package web_ext

import (
	"crypto/sha1"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/revel/revel"
	_ "github.com/three-plus-three/modules/bind"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/menus"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/urlutil"
	"github.com/three-plus-three/sessions"
	sso "github.com/three-plus-three/sso/client"
)

var lifecycleData *Lifecycle
var menuClient menus.Client

func Init(serviceID environment.ENV_PROXY_TYPE, projectTitle string,
	cb func(*Lifecycle) error,
	mode string, createMenuList func(*Lifecycle) ([]toolbox.Menu, error)) {

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
		projectName := ""
		for _, so := range environment.ServiceOptions {
			if so.Id == serviceID {
				projectName = so.Name
			}
		}

		env, err := environment.NewEnvironment(environment.Options{Name: projectName,
			ConfDir: filepath.Join(os.Getenv("hw_root_dir"), "conf")})
		if nil != err {
			log.Println(err)
			os.Exit(-1)
			return
		}

		if revel.RunMode == "test" {
			env.Db.Models.Schema = env.Db.Models.Schema + "_models_test"
			env.Db.Data.Schema = env.Db.Data.Schema + "_test"
		}

		lifecycle, err := NewLifecycle(env)
		if nil != err {
			log.Println(err)
			os.Exit(-1)
			return
		}

		serviceObject := env.GetServiceConfig(serviceID)
		projectContext := serviceObject.Name
		lifecycle.URLPrefix = env.DaemonUrlPath
		lifecycle.URLRoot = env.DaemonUrlPath
		lifecycle.ApplicationContext = env.DaemonUrlPath + projectContext
		lifecycle.ApplicationRoot = env.DaemonUrlPath + projectContext

		lifecycle.Variables = ReadVariables(env, projectTitle)

		lifecycle.Variables["urlPrefix"] = lifecycle.URLPrefix
		lifecycle.Variables["url_prefix"] = lifecycle.URLPrefix
		lifecycle.Variables["url_root"] = lifecycle.URLRoot

		lifecycle.Variables["application_context"] = lifecycle.ApplicationContext
		lifecycle.Variables["application_root"] = lifecycle.ApplicationRoot

		wserviceObject := env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID)
		lifecycle.Variables["user_logout_url"] = wserviceObject.UrlFor(env.DaemonUrlPath, "/sso/logout")

		lifecycle.GetUser = InitUser(lifecycle)

		lifecycle.CurrentUser = func(c *revel.Controller) User {
			username := c.Session[sso.SESSION_USER_KEY]
			if username == "" {
				return nil
			}
			return lifecycle.GetUser(username)
		}
		lifecycle.CheckUser = initSSO(env)

		if revel.DevMode {
			lifecycle.ModelEngine.ShowSQL()
			lifecycle.DataEngine.ShowSQL()
		}

		if err := cb(lifecycle); err != nil {
			log.Println(err)
			os.Exit(-1)
			return
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
		GlobalSessionFilter = sessions.SessionFilter(sso.DefaultSessionKey,
			env.RawDaemonUrlPath, sha1.New, secretKey)

		initTemplateFuncs(lifecycle)

		if !revel.DevMode {
			if fp := flag.Lookup("port"); nil != fp && fp.Value.String() == fp.DefValue {
				revel.Server.Addr = serviceObject.ListenAddr("")
			}
		}

		menuClient = menus.Connect(lifecycleData.Env,
			serviceID,
			menus.Callback(func() ([]toolbox.Menu, error) {
				return createMenuList(lifecycleData)
			}),
			mode,
			"menus.changed",
			urlutil.Join(lifecycleData.Env.DaemonUrlPath, "/menu/"),
			log.New(os.Stderr, "[menus]", log.LstdFlags))

		lifecycleData.OnClosing(menuClient)
	}, 0)

	revel.OnAppStart(func() {
		menuList, err := menuClient.Read()
		if err != nil {
			log.Println(err)
			os.Exit(-1)
			return
		}

		lifecycleData.MenuList = menuList
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
