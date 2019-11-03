module github.com/three-plus-three/modules

go 1.13

require (
	github.com/AreaHQ/go-fixtures v0.0.0-20160603063519-9384cea840d9
	github.com/BurntSushi/toml v0.3.1
	github.com/GeertJohan/go.rice v1.0.0 // indirect
	github.com/Microsoft/go-winio v0.4.14
	github.com/agtorre/gocolorize v1.0.0 // indirect
	github.com/aryann/difflib v0.0.0-20170710044230-e206f873d14a
	github.com/cheekybits/genny v1.0.0
	github.com/google/gops v0.3.6
	github.com/grsmv/inflect v0.0.0-20140723132642-a28d3de3b3ad
	github.com/hjson/hjson-go v3.0.0+incompatible
	github.com/kardianos/osext v0.0.0-20190222173326-2bc1f35cddc0
	github.com/labstack/echo v3.3.10+incompatible
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mitchellh/go-ps v0.0.0-20190716172923-621e5597135b
	github.com/mitchellh/mapstructure v1.1.2
	github.com/opentracing/opentracing-go v1.1.0
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/revel/revel v0.21.0
	github.com/robfig/config v0.0.0-20141207224736-0f78529c8c7e // indirect
	github.com/robfig/pathtree v0.0.0-20140121041023-41257a1839e9 // indirect
	github.com/runner-mei/GoBatis v1.1.1
	github.com/runner-mei/command v0.0.0-20180116024942-77677390382b
	github.com/runner-mei/errors v0.0.0-20191030090348-38af1672ff66
	github.com/runner-mei/log v1.0.0
	github.com/runner-mei/loong v1.0.2
	github.com/runner-mei/orm v1.1.0
	github.com/runner-mei/resty v0.0.0-20191102140647-fa73802f0b7f
	github.com/runner-mei/snmpclient2 v1.1.0
	github.com/three-plus-three/forms v0.0.0-20191018015256-f7b18b98b8e1
	github.com/three-plus-three/sessions v0.0.0-20190127084926-d1bedf90d9a2
	github.com/three-plus-three/sso v0.0.0-20191027121323-ef45d081defa
	github.com/valyala/fasttemplate v1.1.0 // indirect
	github.com/yeka/zip v0.0.0-20180914125537-d046722c6feb
	go.uber.org/zap v1.12.0
	golang.org/x/net v0.0.0-20191101175033-0deb6923b6d9
	golang.org/x/tools v0.0.0-20191101200257-8dbcdeb83d3f
	gopkg.in/ldap.v3 v3.1.0
	xorm.io/xorm v0.8.0
)

replace (
	github.com/AreaHQ/go-fixtures => github.com/runner-mei/go-fixtures v0.0.0-20161015121039-0c8d65b1a339
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.2.0
	github.com/ekanite/ekanite => github.com/runner-mei/ekanite v1.2.0
	github.com/revel/cmd => github.com/runner-mei/revel_cmd v0.12.0
	github.com/revel/revel => github.com/runner-mei/revel v0.12.0
	github.com/superchalupa/go-redfish => github.com/runner-mei/go-redfish v0.0.0-20180620095514-1107c422344a
	github.com/uber/jaeger-client-go => github.com/runner-mei/jaeger-client-go v2.19.1+incompatible

	xorm.io/xorm => github.com/runner-mei/xorm v0.8.1
)
