package web_ext

import (
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/revel/revel"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/urlutil"
	sso "github.com/three-plus-three/sso/client"
	"github.com/three-plus-three/sso/client/revel_sso"
)

func initSSO(env *environment.Environment) revel_sso.CheckFunc {
	ssoURL := env.GetMasterConfig().URLFor(env.DaemonUrlPath, "sessions")
	if mode := env.Config.StringWithDefault("users.login_url", ""); mode == "sso" {
		ssoURL = env.GetMasterConfig().URLFor(env.DaemonUrlPath, "sso")
	}
	ssoClient, err := sso.NewClient(ssoURL)
	if err != nil {
		log.Println(err)
		os.Exit(-1)
		return nil
	}

	return revel_sso.SSO(ssoClient, 30*time.Minute, func(u *url.URL, ctrl *revel.Controller) url.URL {
		copyURL := *u
		copyURL.Scheme = ""
		copyURL.Host = ""
		if !strings.HasPrefix(copyURL.Path, revel.AppRoot) {
			copyURL.Path = urlutil.Join(revel.AppRoot, copyURL.Path)
		}
		return copyURL
	})
}
