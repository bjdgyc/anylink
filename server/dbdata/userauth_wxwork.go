package dbdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"
)

type AuthWXwork struct {
	CorpId  string `json:"corp_id"`
	AgentId string `json:"agent_id"`
	Secret  string `json:"secret"`
}

func init() {
	authRegistry["wxwork"] = reflect.TypeOf(AuthWXwork{})
}

// 验证企微配置参数
func (auth AuthWXwork) checkData(authData map[string]interface{}) error {
	authType := authData["type"].(string)
	bodyBytes, err := json.Marshal(authData[authType])
	if err != nil {
		return errors.New("企微配置填写有误")
	}
	json.Unmarshal(bodyBytes, &auth)

	if auth.CorpId == "" {
		return errors.New("企微的企业ID不能为空")
	}
	if auth.AgentId == "" {
		return errors.New("企微的应用ID不能为空")
	}
	if auth.Secret == "" {
		return errors.New("企微的应用Secret不能为空")
	}
	return nil
}

// 企微用户验证逻辑
func (auth AuthWXwork) checkUser(name, pwd string, g *Group, ext map[string]interface{}) error {
	// 这里的 name 实际上是从企微回调中获取的 code
	// pwd 参数在企微认证中不使用
	// Todo: 占位函数，后续可优化完善统一用户验证逻辑
	// 该逻辑在当前SAML 认证中未使用！！！！
	authType := g.Auth["type"].(string)
	if _, ok := g.Auth[authType]; !ok {
		return fmt.Errorf("%s %s", name, "企微配置中不存在该类型")
	}

	body, err := json.Marshal(g.Auth[authType])
	if err != nil {
		return fmt.Errorf("%s %s", name, err.Error())
	}

	err = json.Unmarshal(body, &auth)
	if err != nil {
		return fmt.Errorf("%s %s", name, err.Error())
	}

	// 通过企微 API 获取用户信息
	userID, err := auth.GetWeworkUser(auth.CorpId, auth.Secret, name) // 这里的 name 实际上是从企微回调中获取的 code
	if err != nil {
		return fmt.Errorf("企微用户信息获取失败: %s", err.Error())
	}

	// 验证用户是否有效
	if userID == "" {
		return fmt.Errorf("企微用户ID为空")
	}

	return nil
}

type WXworkError struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// 企微获取AccessToken API 响应结构
type WXworkTokenResponse struct {
	WXworkError
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

// 企微用户ID信息
type WXworkUserResponse struct {
	WXworkError
	UserID string `json:"userid"`
}

// 获取企微访问令牌
func (auth AuthWXwork) getAccessToken(CorpID, Secret string) (string, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", CorpID, Secret)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	tokenResp := &WXworkTokenResponse{}
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return "", err
	}

	if tokenResp.ErrCode != 0 {
		return "", fmt.Errorf("获取访问令牌失败: %s", tokenResp.ErrMsg)
	}

	return tokenResp.AccessToken, nil
}

// 通过 code 获取企微用户信息
func (auth AuthWXwork) GetWeworkUser(CorpID, Secret, code string) (string, error) {
	// 获取访问令牌
	accessToken, err := auth.getAccessToken(CorpID, Secret)
	if err != nil {
		return "", err
	}

	// 通过 code 获取用户信息
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/auth/getuserinfo?access_token=%s&code=%s", accessToken, code)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	userInfo := &WXworkUserResponse{}
	err = json.Unmarshal(body, &userInfo)
	if err != nil {
		return "", err
	}

	if userInfo.ErrCode != 0 {
		return "", fmt.Errorf("获取用户信息失败: %s", userInfo.ErrMsg)
	}

	return userInfo.UserID, nil
}

// GetAuthWework 从组配置中获取企微认证配置
func GetAuthWework(groupName string) (*AuthWXwork, error) {
	// 获取组配置信息
	groupData := &Group{}
	if err := One("Name", groupName, groupData); err != nil {
		return nil, fmt.Errorf("用户组错误: %v", err)
	}

	// 检查认证类型
	authType, ok := groupData.Auth["type"].(string)
	if !ok || authType != "wxwork" {
		return nil, fmt.Errorf("该组未配置企微认证")
	}

	// 获取企微配置
	config, exists := groupData.Auth["wxwork"]
	if !exists {
		return nil, fmt.Errorf("企微配置不存在")
	}

	wxworkConfig, ok := config.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("企微配置格式错误")
	}

	// 解析配置到结构体
	authwxwork := &AuthWXwork{}
	body, err := json.Marshal(wxworkConfig)
	if err != nil {
		return nil, fmt.Errorf("企微配置序列化失败: %v", err)
	}

	if err := json.Unmarshal(body, authwxwork); err != nil {
		return nil, err
	}

	// 验证配置完整性
	if authwxwork.CorpId == "" || authwxwork.AgentId == "" || authwxwork.Secret == "" {
		return nil, fmt.Errorf("企微配置不完整: CorpId=%s, AgentId=%s, Secret=%s", authwxwork.CorpId, authwxwork.AgentId, authwxwork.Secret)
	}

	return authwxwork, nil
}
