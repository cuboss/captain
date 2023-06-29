package tools

import (
	"bytes"
	model "captain/pkg/models/component"
	"captain/pkg/simple/client/helm"
	"captain/pkg/simple/server/errors"
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/release"
	"io/ioutil"
	erros2 "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DefaultEcrCredentialDeploymentName = "ecr-credential"
	DefaultEcrCredentialNamespace      = "kube-system"
	DefaultEcrCredentialRgSecret       = "ecr-helper-secret"
	createEcrUserUri                   = "/csk/createtemporaryuser"
	UpdateEcrUserUri                   = "/csk/updatetemporaryuser"
	deleteEcrUserUri                   = "/csk/deletetemporaryuser"
	HeaderContentType                  = "Content-Type"
	HeaderJSONContentTypeValue         = "application/json"
	DefaultEcrCredentialIngressName    = "ecrCredential-ingress"
)

type auth struct {
	AccessKey string
	SecretKey string
}

type HttpClient struct {
	url        string
	httpClient *http.Client
	auth       *auth
	mu         sync.Mutex
	headers    map[string]string
	body       map[string]interface{}
	partner    bool
}

type Replay struct {
	Code    interface{} `json:"code"`
	Data    interface{} `json:"data,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
}

// EcrCredentialOptions viper configmap读取的参数值
type EcrCredentialOptions struct {
	ApiGateway          string `json:"apiGateway" yaml:"apiGateway"`
	AccessKey           string `json:"accessKey" yaml:"accessKey"`
	SecretKey           string `json:"secretKey" yaml:"secretKey"`
	ValuesImageRegistry string `json:"valuesImageRegistry" yaml:"valuesImageRegistry"`
	ValuesTag           string `json:"valuesTag" yaml:"valuesTag"`
	ValuesArchitecture  string `json:"valuesArchitecture" yaml:"valuesArchitecture"`
}

func NewEcrCredentialOptions() *EcrCredentialOptions {
	return &EcrCredentialOptions{
		ApiGateway: "",
	}
}

type EcrCredential struct {
	options          *EcrCredentialOptions
	client           *helm.Client
	clusterComponent *model.ClusterComponent
	kubeClient       *kubernetes.Clientset
	release          string
	chart            string
	version          string
	values           map[string]interface{}
}

type configInfo struct {
	//version从helm里面取
	Version        string `yaml:"ecr-api-version" json:"version"`
	ServiceAccount string `yaml:"service-account" json:"service-account"`
	Namespace      string `yaml:"namespace" json:"namespace"`
	EcrRegistry    string `yaml:"ecr-registry,omitempty" json:"ecr-registry,omitempty"`
}

// secret存储
type ecrSecret struct {
	User
	Server string `json:"server,omitempty"`
	Email  string `json:"email,omitempty"`
	Auth   string `json:"auth,omitempty" `
}

type User struct {
	Username string `json:"user_name,omitempty" yaml:"user_name,omitempty" `
	Passwd   string `json:"user_password,omitempty" yaml:"user_password,omitempty"`
}

// ecrUser  request Body
type ecrUser struct {
	User
	Auth string `json:"auth,omitempty" yaml:"auth,omitempty"`
}

type DbAuth struct {
	HarborServer string     `json:"harbor_id"`
	DbAuthInfo   DbAuthInfo `json:"auth_info"`
}
type DbAuthInfo struct {
	All      bool     `json:"all"`
	Projects []string `json:"projects"`
}

func NewEcrCredential(options *EcrCredentialOptions, client *helm.Client, kubeClient *kubernetes.Clientset, clusterComponent *model.ClusterComponent) (*EcrCredential, error) {
	ec := &EcrCredential{
		options:          options,
		client:           client,
		kubeClient:       kubeClient,
		clusterComponent: clusterComponent,
		release:          clusterComponent.ReleaseName,
		chart:            clusterComponent.ChartName,
		version:          clusterComponent.ChartVersion,
	}
	return ec, nil
}

func (p *EcrCredential) setDefaultValue(clusterComponent *model.ClusterComponent, config configInfo, user ecrUser, isInstall bool) error {
	values := map[string]interface{}{}
	var err error
	//根据不同版本EcrCredential填充  保留isInstall做控制
	switch clusterComponent.ChartVersion {
	case "1.0.0":
		values, err = p.valuse010Binding(config, user)
	default:
		return err
	}
	p.values = values
	return err
}

func (p *EcrCredential) valuse010Binding(config configInfo, user ecrUser) (map[string]interface{}, error) {

	values := map[string]interface{}{}
	configMap := map[string]interface{}{}
	secret := map[string]interface{}{}
	opts := p.options
	//增加校验 校验
	err2 := checkOptions(opts)
	if err2 != nil {
		return nil, err2
	}

	//灵活控制helm values的仓库值 单云池唯一

	var defaultImageRegistry string
	lastIndex := len(opts.ValuesImageRegistry) - 1
	if string(opts.ValuesImageRegistry[lastIndex]) == "/" {
		defaultImageRegistry = opts.ValuesImageRegistry[:lastIndex]
	} else {
		defaultImageRegistry = opts.ValuesImageRegistry
	}
	values["defaultImageRegistry"] = defaultImageRegistry
	values["architecture"] = opts.ValuesArchitecture
	//如果需要调整镜像版本 修改漏洞的话 修改tag
	values["image.tag"] = opts.ValuesTag
	values["initJob.initJobImage.tag"] = opts.ValuesTag
	values["clearJob.clearJobImage.tag"] = opts.ValuesTag
	values["ingress.enabled"] = false
	values["autoscaling.enabled"] = false
	values["serviceAccount.create"] = true
	values["resources.limits.cpu"] = "200m"
	values["resources.limits.memory"] = "512Mi"
	values["resources.requests.cpu"] = "100m"
	values["resources.requests.memory"] = "256Mi"
	values["service.type"] = "ClusterIP"

	values, err := MergeValueMap(values)
	if err != nil {
		return nil, err
	}
	var rg []string
	err = yaml.Unmarshal([]byte(config.EcrRegistry), &rg)
	if err != nil {
		return nil, err
	}
	//cm 和 secret 由captain控制
	configMap["apiVersion"] = opts.ValuesTag
	configMap["enabled"] = true
	configMap["namespace"] = config.Namespace
	configMap["serviceAccount"] = config.ServiceAccount
	configMap["ecrRegistry"] = rg
	secret["userName"] = user.Username
	secret["password"] = user.Passwd
	secret["enable"] = true
	values["configMap"] = configMap
	values["registrySecret"] = secret

	return values, nil
}

func (p *EcrCredential) Install() (*release.Release, error) {

	// ecr创建临时用户
	user, rg, err := p.setEcrUser(true)
	if err != nil {
		return nil, err
	}
	//获取configmap
	conf, err := p.setCredentialCm(rg, true)
	if err != nil {
		return nil, err
	}

	err = p.setDefaultValue(p.clusterComponent, conf, user, true)
	if err != nil {
		return nil, err
	}

	// init-job在helm 中通过钩子执行

	release, err := installChart(p.client, p.release, p.chart, p.version, p.values)
	if err != nil {
		return nil, err
	}

	if err = waitForRunning(DefaultEcrCredentialNamespace, DefaultEcrCredentialDeploymentName, 1, p.kubeClient); err != nil {
		return nil, err
	}

	return release, err
}

func (p *EcrCredential) Upgrade() (*release.Release, error) {
	// ecr创建临时用户
	user, rg, err := p.setEcrUser(false)
	if err != nil {
		return nil, err
	}
	//获取configmap
	conf, err := p.setCredentialCm(rg, false)
	if err != nil {
		return nil, err
	}
	err = p.setDefaultValue(p.clusterComponent, conf, user, false)
	if err != nil {
		return nil, err
	}
	//更新chart 重新生成证书
	rel, err := upgradeChart(p.client, p.release, p.chart, p.version, p.values)
	return rel, err

}

func (p *EcrCredential) Uninstall() (*release.UninstallReleaseResponse, error) {
	//delete ecr临时用户
	err := p.deleteEcrUser()
	if err != nil {
		return nil, err
	}

	//clusterRole clusterRoleBinding serviceAccount helm卸载的时候 自动删除
	return uninstall(p.client, p.kubeClient, p.release, DefaultEcrCredentialIngressName, DefaultEcrCredentialNamespace)

}

func (p *EcrCredential) Status(release string) ([]model.ClusterComponentResStatus, error) {

	//获取组件状态
	return getReleaseStatus(p.client, release)
}

// setCredentialCm 处理前端传入的值
func (p *EcrCredential) setCredentialCm(rg []string, install bool) (configInfo, error) {
	i, ok := p.clusterComponent.Parameters["configmap"]
	config := configInfo{}
	//将unkkown map 转化为configInfo 借助json转换 直接struct转化存在问题
	marshal, err := json.Marshal(i)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(marshal, &config)
	if err != nil {
		return config, err
	}
	if reflect.DeepEqual(config, configInfo{}) {
		return config, errors.New("ecrCredential parameters auth is empty")
	}

	if ok {
		out, _ := yaml.Marshal(rg)
		config.EcrRegistry = string(out)
		return config, nil
	} else {
		return configInfo{}, errors.New("Parameters configmap error")
	}
}

// SetRegistrySecret 处理前端传入的值
func (p *EcrCredential) setEcrUser(isInstall bool) (ecrUser, []string, error) {
	//install 和update时 用到
	//全量更新 后续可优化
	rg := []string{}
	authList := []DbAuth{}
	i, ok := p.clusterComponent.Parameters["auth"]
	//将unkkown map 转化为[]dbAuth 借助json转换 直接struct转化存在问题
	marshal, err := json.Marshal(i)
	if err != nil {
		return ecrUser{}, rg, err
	}
	err = json.Unmarshal(marshal, &authList)
	if err != nil {
		return ecrUser{}, rg, err
	}
	if len(authList) == 0 {
		return ecrUser{}, rg, errors.New("ecrCredential parameters auth is empty")

	}
	if ok {

		user, err := p.createOrUpdateEcrUser(authList, isInstall)
		if err != nil {
			return ecrUser{}, rg, err
		}
		for _, v := range authList {
			rg = append(rg, v.HarborServer)
		}
		return user, rg, err

	} else {
		return ecrUser{}, rg, errors.New("Parameters auth error")

	}
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return erros2.IsNotFound(err)
}

func (p *EcrCredential) deleteEcrUser() error {

	user := ecrUser{}
	get, err := p.kubeClient.CoreV1().Secrets(DefaultEcrCredentialNamespace).Get(context.TODO(), DefaultEcrCredentialRgSecret, v1.GetOptions{})
	if isNotFound(err) {
		user = GenUser(p.clusterComponent.CkeClusterId, false)
	} else if err != nil {
		klog.Infof("deleteEcrUser phase get secret error %v", err)
		return err
	} else {
		user.Username = string(get.Data["user-name"])
	}
	opts := p.options
	err2 := checkOptions(opts)
	if err2 != nil {
		return err2
	}

	url := genUrl(opts.ApiGateway, deleteEcrUserUri)
	httpClient := newHttpClient(url, opts.AccessKey, opts.SecretKey)

	r, err := httpClient.Partner().WithRawBody(user).DEL()
	if err != nil {
		//如果用户不存在 认为删除成功
		if r.GetCode() == 400 {
			return nil
		} else {
			return errors.New(fmt.Sprintf("delete ecr user error %v", err))
		}
	}
	return err
}

// createOrUpdateEcrUser 请求ECR创建和更新临时用户
func (p *EcrCredential) createOrUpdateEcrUser(secretList []DbAuth, isinstall bool) (ecrUser, error) {
	var url string
	var user ecrUser
	opts := p.options
	err2 := checkOptions(opts)
	if err2 != nil {
		return ecrUser{}, err2
	}

	if isinstall {
		url = genUrl(opts.ApiGateway, createEcrUserUri)
		user = GenUser(p.clusterComponent.CkeClusterId, true)
		httpClient := newHttpClient(url, opts.AccessKey, opts.SecretKey)
		createRequest := GenUserCreateRequest(user, secretList)
		_, err := httpClient.Partner().WithRawBody(createRequest).Post()
		if err != nil {
			//重复创建 仍然为200
			return user, errors.New(fmt.Sprintf("init ecr user error %v", err))
		}

		return user, err
	} else {
		//如果是升级 更新用户
		url = genUrl(opts.ApiGateway, UpdateEcrUserUri)
		user = GenUser(p.clusterComponent.CkeClusterId, false)
		//先去校验集群内的密码
		get, err := p.kubeClient.CoreV1().Secrets(DefaultEcrCredentialNamespace).Get(context.TODO(), DefaultEcrCredentialRgSecret, v1.GetOptions{})
		if isNotFound(err) {
			//集群内没有存储用户名和密码 更新密码
			user.Passwd = genRandomPassword(10)
		} else {
			if err != nil {
				return ecrUser{}, err
			}
			//使用旧密码
			user.Passwd = string(get.Data["user-passwd"])
		}
		reCreateUrl := genUrl(opts.ApiGateway, UpdateEcrUserUri)
		reCreateHttpClient := newHttpClient(reCreateUrl, opts.AccessKey, opts.SecretKey)
		updateRequest := GenUserUpdateRequest(user, secretList)
		r, err := reCreateHttpClient.Partner().WithRawBody(updateRequest).Put()
		if err != nil {
			//如果用户不存在 改为创建user
			if r.GetCode() == 400 {
				//重新生成密码
				user.Passwd = genRandomPassword(10)
				createRequest := GenUserCreateRequest(user, secretList)
				createUrl := genUrl(opts.ApiGateway, createEcrUserUri)
				reCreateHttpClient.url = createUrl
				r, err = reCreateHttpClient.Partner().WithRawBody(createRequest).Post()
				if r == nil {
					return user, errors.New(fmt.Sprintf("recreate ecr user error %v", err))
				}
				return user, err
			} else {
				return user, errors.New(fmt.Sprintf("update ecr user error %v", err))
			}
		}
		return user, err
	}
}

func GenUser(clusterid string, updatePasswd bool) ecrUser {
	//安装组件的时候 初始化用户名和密码 写入secret
	//更新用户 可以设置密码为空
	newUser := ecrUser{}
	if updatePasswd == true {
		user := User{
			Username: fmt.Sprintf("ecr-helper-%s", gen20UserName(clusterid)),
			Passwd:   genRandomPassword(10),
		}
		newUser = ecrUser{
			User: user,
		}
	} else {
		user := User{
			Username: fmt.Sprintf("ecr-helper-%s", gen20UserName(clusterid)),
		}
		newUser = ecrUser{
			User: user,
		}

	}
	return newUser
}

func gen20UserName(clusterId string) string {
	id := clusterId[4:24]
	//clusterId 24位Id   目标生成固定 全剧唯一  可溯源 4-20
	col := 5
	times := 4
	translate := ""
	for i := 0; i < col; i++ {
		for j := 0; j < times; j++ {
			translate = translate + id[j*5+i:j*5+i+1]
		}
	}
	return translate
}

func genUrl(endpoint, path string) string {
	u, _ := url.Parse(endpoint)
	rel, _ := url.Parse(path)
	u = u.ResolveReference(rel)
	return u.String()
}

func genRandomPassword(length int) string {
	rand.Seed(time.Now().UnixNano())
	digits := "0123456789"
	specials := "~%^*!@#$?"
	all := "ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		digits + specials
	buf := make([]byte, length)
	buf[0] = digits[rand.Intn(len(digits))]
	buf[1] = specials[rand.Intn(len(specials))]
	for i := 2; i < length; i++ {
		buf[i] = all[rand.Intn(len(all))]
	}
	rand.Shuffle(len(buf), func(i, j int) {
		buf[i], buf[j] = buf[j], buf[i]
	})
	return string(buf)
}

func newHttpClient(url, AccessKey, SecretKey string) *HttpClient {
	client := &HttpClient{
		url:     url,
		headers: make(map[string]string),
		body:    make(map[string]interface{}),
	}
	client.httpClient = http.DefaultClient
	return client
}

func (c *HttpClient) Partner() *HttpClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.partner = true
	return c
}
func (c *HttpClient) DEL() (*Replay, error) {
	c.headers[HeaderContentType] = HeaderJSONContentTypeValue
	return c.do(http.MethodDelete)
}
func (c *HttpClient) Post() (*Replay, error) {
	c.headers[HeaderContentType] = HeaderJSONContentTypeValue
	return c.do(http.MethodPost)
}
func (c *HttpClient) Put() (*Replay, error) {
	c.headers[HeaderContentType] = HeaderJSONContentTypeValue
	return c.do(http.MethodPut)
}

func (c *HttpClient) do(method string) (*Replay, error) {
	bodyBuf := &bytes.Buffer{}
	err := json.NewEncoder(bodyBuf).Encode(c.body)
	if err != nil {
		return nil, err
	}
	url := c.url
	index := strings.Index(url, "?")
	if index > 0 {
		// 请求地址中包含参数，统一放在body中
		urlParam := url[index+1:]
		split := strings.Split(urlParam, "&")
		for _, s := range split {
			p := strings.Split(s, "=")
			c.body[p[0]] = p[1]
		}
	}
	request, err := http.NewRequest(method, c.url, bodyBuf)
	if err != nil {
		return nil, err
	}
	if c.partner {

	}

	// 自定义请求头
	if c.headers != nil {
		for k, v := range c.headers {
			request.Header[k] = []string{v}
		}
	}

	resp, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	return checkResponse(resp)
}

func (r *Replay) GetCode() int {
	s := fmt.Sprintf("%v", r.Code)
	code, _ := strconv.Atoi(s)
	return code
}

func checkResponse(resp *http.Response) (*Replay, error) {
	const MinValidStatusCode = 200
	const MaxValidStatusCode = 500

	defer resp.Body.Close()
	r := Replay{}
	if resp.StatusCode >= MinValidStatusCode && resp.StatusCode <= MaxValidStatusCode {
		err := json.NewDecoder(resp.Body).Decode(&r)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", r.Message, err)
		}
		switch r.GetCode() {
		case 400:
			return &r, fmt.Errorf("user not found: %s", r.Message)
		case 403:
			return &r, fmt.Errorf("user forbidden: %s", r.Message)
		case 200:
			return &r, nil
		default:
			return nil, fmt.Errorf("server error: %s", r.Message)
		}
	}
	resBodyReads, _ := ioutil.ReadAll(resp.Body)
	return nil, fmt.Errorf("error to parse response: %s", string(resBodyReads))
}
func (c *HttpClient) WithRawBody(raw interface{}) *HttpClient {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, _ := json.Marshal(raw)
	_ = json.Unmarshal(b, &c.body)
	return c
}

func GenUserCreateRequest(user ecrUser, dbAuth []DbAuth) ecrUser {
	marshal, _ := json.Marshal(dbAuth)

	return ecrUser{
		User: user.User,
		Auth: string(marshal),
	}

}
func GenUserUpdateRequest(user ecrUser, dbAuth []DbAuth) ecrUser {
	marshal, _ := json.Marshal(dbAuth)

	return ecrUser{
		User: User{Username: user.Username, Passwd: user.Passwd},
		Auth: string(marshal),
	}
}

// GenUserDeleteRequest 预留
func GenUserDeleteRequest(user ecrUser) ecrUser {
	return user
}

func checkOptions(opts *EcrCredentialOptions) error {
	if opts.ApiGateway == "" {
		return errors.New("ecr apiGateway is empty")
	} else {

		url, err := url.Parse(opts.ApiGateway)
		if err != nil {
			klog.Infof("tcp ping ecr apiGateway error ,url is illegal ", err)
			return errors.New(fmt.Sprintf("tcp ping addr %s error,url is illegal: %v", opts.ApiGateway, err))
		}

		//tcp ping  ecr apiGateway
		conect, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", url.Hostname(), url.Port()), 5*time.Second)
		if err != nil {
			klog.Infof("tcp ping ecr apiGateway", err)
			return errors.New(fmt.Sprintf("tcp ping addr %s failed: %v", opts.ApiGateway, err))
		}
		defer conect.Close()
	}
	return nil
}
