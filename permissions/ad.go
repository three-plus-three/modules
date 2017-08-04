package permissions

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"gopkg.in/ldap.v2"
)

type ADConfig struct {
	LdapServer string
	LdapPort   string
	Username   string
	Password   string
	BaseDN     string
}

//读取AD的配置
func GetADConfig() (ADConfig, error) {
	ad, err := ioutil.ReadFile("D:/go/HughRevel/src/web_example/conf/AD.json")
	var adConfig ADConfig
	if err != nil {
		return adConfig, err
	}
	json.Unmarshal(ad, &adConfig)
	return adConfig, nil
}

func GetUser() ([]User, error) {
	var users []User
	ad, err := GetADConfig()
	if err != nil {
		return users, err
	}
	//连接活动目录
	var ldapServer = ad.LdapServer
	port, err := strconv.ParseInt(ad.LdapPort, 10, 16)
	var ldapPort = port

	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", ldapServer, ldapPort))

	if err != nil {
		return users, err
	}

	defer l.Close()

	err = l.Bind(ad.Username, ad.Password)

	if err != nil {

		return users, err
	}

	searchRequest := ldap.NewSearchRequest(
		ad.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalPerson))",
		[]string{},
		nil,
	)

	//获取数据
	sr, err := l.Search(searchRequest)

	if err != nil {
		return users, err
	}

	for i := 0; i < len(sr.Entries); i++ {
		var user User
		user.Name = sr.Entries[i].GetAttributeValue("name")
		user.Description = sr.Entries[i].GetAttributeValue("description")
		user.Attributes["streetAddress"] = sr.Entries[i].GetAttributeValue("streetAddress")
		user.Attributes["company"] = sr.Entries[i].GetAttributeValue("company")
		user.Attributes["telephoneNumber"] = sr.Entries[i].GetAttributeValue("telephoneNumber")
		user.Source = "AD"
		created := strings.Split(sr.Entries[i].GetAttributeValue("whenCreated"), ".")[0]
		changed := strings.Split(sr.Entries[i].GetAttributeValue("whenChanged"), ".")[0]
		user.CreatedAt = changeTime(created)
		user.UpdatedAt = changeTime(changed)
		users = append(users, user)
	}
	return users, nil
}

func changeTime(str string) time.Time {
	y := str[:4]
	m := str[4:6]
	d := str[6:8]
	h := str[8:10]
	mi := str[10:12]
	s := str[12:]
	time1 := strings.Join([]string{y, m, d}, "-")
	time2 := strings.Join([]string{h, mi, s}, ":")
	toBeCharge := strings.Join([]string{time1, time2}, " ")
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(timeLayout, toBeCharge, loc)
	hh, _ := time.ParseDuration("1h")
	return theTime.Add(8 * hh)
}
