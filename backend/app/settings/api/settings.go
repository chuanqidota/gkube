package api

import (
	"encoding/json"
	"os"

	"github.com/gin-gonic/gin"
	"gkube/pkg/response"
)

type settings struct{}

var Settings = new(settings)

const settingsFile = "config/auth-settings.json"

type OIDCConfig struct {
	Enabled      bool   `json:"enabled"`
	Issuer       string `json:"issuer"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RedirectURI  string `json:"redirectUri"`
	Scopes       string `json:"scopes"`
	UsernameClaim string `json:"usernameClaim"`
	EmailClaim   string `json:"emailClaim"`
	GroupsClaim  string `json:"groupsClaim"`
}

type LDAPConfig struct {
	Enabled            bool   `json:"enabled"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	BindDN             string `json:"bindDN"`
	BindPassword       string `json:"bindPassword"`
	UserSearchBase     string `json:"userSearchBase"`
	UserSearchFilter   string `json:"userSearchFilter"`
	GroupSearchBase    string `json:"groupSearchBase"`
	GroupSearchFilter  string `json:"groupSearchFilter"`
	StartTLS           bool   `json:"startTLS"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify"`
}

type AuthSettings struct {
	OIDC *OIDCConfig `json:"oidc"`
	LDAP *LDAPConfig `json:"ldap"`
}

func (s *settings) GetAuthSettings(c *gin.Context) {
	settings := &AuthSettings{
		OIDC: &OIDCConfig{},
		LDAP: &LDAPConfig{},
	}

	data, err := os.ReadFile(settingsFile)
	if err == nil {
		json.Unmarshal(data, settings)
	}

	response.Success(c, "执行成功", settings)
}

func (s *settings) UpdateAuthSettings(c *gin.Context) {
	var settings AuthSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		response.Fail(c, "序列化配置失败: "+err.Error())
		return
	}

	if err := os.WriteFile(settingsFile, data, 0644); err != nil {
		response.Fail(c, "保存配置失败: "+err.Error())
		return
	}

	response.Success(c, "配置保存成功", nil)
}
