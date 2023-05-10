package tools

import (
	"bytes"
	model "captain/pkg/models/component"
	"captain/pkg/server/config"
	"captain/pkg/simple/client/helm"
	"captain/pkg/simple/server/errors"
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/release"
	"io/ioutil"
	v12 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	erros2 "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	//DefaultEcrCredentialServiceName = "ecr-credential-svc"

	DefaultEcrCredentialDeploymentName     = "ecr-credential-deployment"
	DefaultEcrCredentialNamespace          = "kube-system"
	DefaultEcrCredentialReNewJobName       = "ecr-credential-renew-job"
	DefaultEcrCredentialClearJobName       = "ecr-credential-clear-job"
	DefaultEcrCredentialServiceAccountName = "ecr-helper-sa"
	DefaultEcrCredentialConfigMap          = "ecr-helper-cm"

	DefaultEcrCredentialRgSecret = "ecr-helper-secret"

	createEcrUserUri = "/csk/createtemporaryuser"
	UpdateEcrUserUri = "/csk/updatetemporaryuser"
	deleteEcrUserUri = "/csk/deletetemporaryuser"

	HeaderContentType          = "Content-Type"
	HeaderJSONContentTypeValue = "application/json"

	DefaultEcrCredentialIngressName = "ecrCredential-ingress"
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
	ApiGateway    string `json:"apiGateway" yaml:"apiGateway"`
	AccessKey     string `json:"accessKey" yaml:"accessKey"`
	SecretKey     string `json:"secretKey" yaml:"secretKey"`
	ClearJobImage string `json:"clearJobImage" yaml:"clearJobImage"`
	ReNewJobImage string `json:"reNewJobImage" yaml:"reNewJobImage"`
}

func NewEcrCredentialOptions() *EcrCredentialOptions {
	return &EcrCredentialOptions{
		ApiGateway: "",
	}
}

type EcrCredential struct {
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
	Version        string `yaml:"ecr-api-version"`
	ServiceAccount string `yaml:"service-account"`
	Namespace      string `yaml:"namespace"`
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

func NewEcrCredential(client *helm.Client, kubeClient *kubernetes.Clientset, clusterComponent *model.ClusterComponent) (*EcrCredential, error) {
	ec := &EcrCredential{
		client:           client,
		kubeClient:       kubeClient,
		clusterComponent: clusterComponent,

		release: clusterComponent.ReleaseName,
		chart:   clusterComponent.ChartName,
		version: clusterComponent.ChartVersion,
	}
	return ec, nil
}

func (p *EcrCredential) setDefaultValue(clusterComponent *model.ClusterComponent, isInstall bool) {
	values := map[string]interface{}{}
	//根据不同版本EcrCredential填充  保留isInstall做控制

	switch clusterComponent.ChartVersion {
	case "0.1.0":
		values = p.valuse010Binding()

	}

	p.values = values
}

func (p *EcrCredential) valuse010Binding() map[string]interface{} {

	values := map[string]interface{}{}

	/* image根据 chart版本控制
	values["defaultImageRegistry"] = defaultIMageRegistry
	values["image.tag"] = "v1"
	values["initJob.initJobImage.tag"] = "v1.0.0"
	*/
	values["ingress.enabled"] = false
	values["autoscaling.enabled"] = false
	values["serviceAccount.create"] = true
	//cm 和 secret 由captain控制
	values["registrySecret.enable"] = false
	values["configMap.enabled"] = false

	values["resources.limits.cpu"] = "200m"
	values["resources.limits.memory"] = "512Mi"
	values["resources.requests.cpu"] = "100m"
	values["resources.requests.memory"] = "256Mi"

	values["service.type"] = "ClusterIP"

	return values
}

func (p *EcrCredential) Install() (*release.Release, error) {

	//抽离对Parameters操作  即configmap secret
	p.setDefaultValue(p.clusterComponent, true)

	// ecr创建临时用户
	err := p.setRegistrySecret(true)
	if err != nil {
		return nil, err
	}

	//更新configmap
	err = p.setCredentialCm(true)
	if err != nil {
		return nil, err
	}
	// init-job在helm 中通过钩子执行

	release, err := installChart(p.client, p.release, p.chart, p.version, p.values)
	if err != nil {
		return nil, err
	}

	if err = waitForRunning(p.clusterComponent.Namespace, DefaultEcrCredentialDeploymentName, 1, p.kubeClient); err != nil {
		return nil, err
	}

	return release, err
}

func (p *EcrCredential) Upgrade() (*release.Release, error) {

	upgrade := false
	emptyConfig := configInfo{}
	get, err := p.kubeClient.CoreV1().ConfigMaps(DefaultEcrCredentialNamespace).Get(context.TODO(), DefaultEcrCredentialConfigMap, v1.GetOptions{})
	if !isNotFound(err) {
		if err != nil {
			klog.Errorf("upgrade phase get configmap err %v", err)
			return nil, err
		}

		currentCf := get.Data["configmap-ep.yaml"]
		err = yaml.Unmarshal([]byte(currentCf), &emptyConfig)
		if err != nil {
			return nil, err
		}
		//版本不一致触发更新chart
		if emptyConfig.Version != p.version {
			upgrade = true
		}
	}

	if upgrade == true {
		p.setDefaultValue(p.clusterComponent, false)

		//clearJob
		job, err := genJob("clear")
		if err != nil {
			klog.Errorf("genJob error %v ", err)
			return nil, err
		}

		options := v1.ListOptions{
			LabelSelector: "app.kubernetes.io/name=ecr-credential",
		}
		list, err := p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).List(context.TODO(), options)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("list job error %v", err))
		}

		for _, i := range list.Items {
			if strings.Contains(i.Name, DefaultEcrCredentialClearJobName) {
				err = p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).Delete(context.TODO(), i.Name, v1.DeleteOptions{})
				if err != nil {
					return nil, errors.New(fmt.Sprintf("delete old  job error %v", err))
				}
			}
		}
		_, err = p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).Create(context.TODO(), &job, v1.CreateOptions{})
		if err != nil {
			return nil, errors.New(fmt.Sprintf("exec job error %v", err))
		}

		//重新创建临时用户 此时环境能铲除 但是ecr数据库依然存在
		err = p.setRegistrySecret(false)
		if err != nil {
			return nil, err
		}

		//更新configmap
		err = p.setCredentialCm(false)
		if err != nil {
			return nil, err
		}

		//更新chart 重新生成证书
		rel, err := upgradeChart(p.client, p.release, p.chart, p.version, p.values)
		return rel, err
	}

	// ecr检查临时用户 全量更新
	err = p.setRegistrySecret(false)
	if err != nil {
		return nil, err
	}

	//更新configmap
	err = p.setCredentialCm(false)
	if err != nil {
		return nil, err
	}

	//renewJob
	job, err := genJob("renew")
	if err != nil {
		klog.Errorf("genJob error %v ", err)
		return nil, err
	}
	options := v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=ecr-credential",
	}
	list, err := p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).List(context.TODO(), options)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("list job error %v", err))
	}

	for _, i := range list.Items {
		if strings.Contains(i.Name, DefaultEcrCredentialReNewJobName) {
			err = p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).Delete(context.TODO(), i.Name, v1.DeleteOptions{})
			if err != nil {
				return nil, errors.New(fmt.Sprintf("delete old  job error %v", err))
			}
		}
	}
	_, err = p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).Create(context.TODO(), &job, v1.CreateOptions{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("exec job error %v", err))
	}

	return nil, err
}

func (p *EcrCredential) Uninstall() (*release.UninstallReleaseResponse, error) {

	//delete ecr临时用户
	err := p.deleteEcrUser()
	if err != nil {
		return nil, err
	}

	// clear job执行
	job, err := genJob("clear")
	if err != nil {
		klog.Errorf("genJob error %v ", err)
		return nil, err
	}
	options := v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=ecr-credential",
	}
	list, err := p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).List(context.TODO(), options)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("list job error %v", err))
	}

	for _, i := range list.Items {
		if strings.Contains(i.Name, DefaultEcrCredentialClearJobName) {
			err = p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).Delete(context.TODO(), i.Name, v1.DeleteOptions{})
			if err != nil {
				return nil, errors.New(fmt.Sprintf("delete old  job error %v", err))
			}
		}
	}
	_, err = p.kubeClient.BatchV1().Jobs(DefaultEcrCredentialNamespace).Create(context.TODO(), &job, v1.CreateOptions{})
	if err != nil {
		return nil, errors.New(fmt.Sprintf("exec job error %v", err))
	}

	return uninstall(p.client, p.kubeClient, p.release, DefaultEcrCredentialIngressName, p.clusterComponent.Namespace)

}

func (p *EcrCredential) Status(release string) ([]model.ClusterComponentResStatus, error) {

	//获取组件状态
	return getReleaseStatus(p.client, release)
}

// setCredentialCm 处理前端传入的值
func (p *EcrCredential) setCredentialCm(install bool) error {
	i, ok := p.clusterComponent.Parameters["configmap"]
	if ok {
		err := p.createCm(i.(configInfo))
		if err != nil {
			return err
		}
	} else {
		err := p.createCm(configInfo{})
		if err != nil {
			return err
		}
	}

	return nil
}

// SetRegistrySecret 处理前端传入的值
func (p *EcrCredential) setRegistrySecret(isInstall bool) error {
	//install 和update时 用到
	//全量更新 后续可优化
	i, ok := p.clusterComponent.Parameters["auth"]
	if ok {
		err := p.createOrUpdateEcrUser(i.([]DbAuth), isInstall)
		if err != nil {
			return err
		}
	} else {
		err := p.createOrUpdateEcrUser([]DbAuth{}, isInstall)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *EcrCredential) createCm(configInfo configInfo) error {

	get, err := p.kubeClient.CoreV1().ConfigMaps(DefaultEcrCredentialNamespace).Get(context.TODO(), DefaultEcrCredentialConfigMap, v1.GetOptions{})
	if isNotFound(err) {
		configInfo = setcfDefault(configInfo, p.version)
		data := make(map[string]string)
		marshal, _ := yaml.Marshal(configInfo)

		data["configmap-ep.yaml"] = string(marshal)
		newConf := corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      DefaultEcrCredentialConfigMap,
				Namespace: DefaultEcrCredentialNamespace,
			},
			Data: data,
		}
		_, err = p.kubeClient.CoreV1().ConfigMaps(DefaultEcrCredentialNamespace).Create(context.TODO(), &newConf, v1.CreateOptions{})
		if err != nil {
			klog.Errorf("create EcrCredential Configmap error %v", err)
			return err
		}
	} else if err != nil {

		if err != nil {
			klog.Errorf("get EcrCredential Configmap error %v", err)
			return err
		}
	} else {

		configInfo = setcfDefault(configInfo, p.version)
		marshal, _ := yaml.Marshal(configInfo)
		get.Data["configmap-ep.yaml"] = string(marshal)

		_, err = p.kubeClient.CoreV1().ConfigMaps(DefaultEcrCredentialNamespace).Update(context.TODO(), get, v1.UpdateOptions{})
		if err != nil {
			klog.Errorf("update EcrCredential Configmap error %v", err)
			return err
		}
	}
	return err
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
		klog.Errorf("deleteEcrUser phase get secret error %v", err)
		return err

	} else {
		err = json.Unmarshal(get.Data["auth"], &user)
		if err != nil {
			klog.Errorf("Unmarshal secret  auth  error %v", err)
			return err
		}
	}
	opts, err := getEcrOpts()
	if err != nil {
		return err
	}

	url := genUrl(opts.ApiGateway, deleteEcrUserUri)
	httpClient := newHttpClient(url, opts.AccessKey, opts.SecretKey)

	//deleteBody
	//deleteRequest := GenUserDeleteRequest(user)
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
func (p *EcrCredential) createOrUpdateEcrUser(secretList []DbAuth, isinstall bool) error {

	//get检验是否已存在用户 仅在集群内  ecr侧 如果存在 直接覆盖

	get, err := p.kubeClient.CoreV1().Secrets(DefaultEcrCredentialNamespace).Get(context.TODO(), DefaultEcrCredentialRgSecret, v1.GetOptions{})
	if isNotFound(err) {
		//创建临时用户

		var url string
		var user ecrUser
		opts, err := getEcrOpts()
		if err != nil {
			return err
		}
		if isinstall {
			url = genUrl(opts.ApiGateway, createEcrUserUri)
			user = GenUser(p.clusterComponent.CkeClusterId, true)

		} else {
			//如果是升级 更新用户
			url = genUrl(opts.ApiGateway, UpdateEcrUserUri)
			user = GenUser(p.clusterComponent.CkeClusterId, false)
		}

		secret := setUser(user, secretList)
		data := make(map[string]string)
		marshal, _ := json.Marshal(user)
		data["auth"] = string(marshal)

		if len(secret) != 0 {
			rgs, _ := json.Marshal(secret)
			data["ecrRegistry"] = string(rgs)
		} else {
			return errors.New("auth is empty")
		}

		newSecret := corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Name:      DefaultEcrCredentialRgSecret,
				Namespace: DefaultEcrCredentialNamespace,
			},
			Type:       corev1.SecretTypeOpaque,
			StringData: data,
		}

		httpClient := newHttpClient(url, opts.AccessKey, opts.SecretKey)

		createRequest := GenUserCreateRequest(user, secretList)

		_, err = httpClient.Partner().WithRawBody(createRequest).Post()
		if err != nil {
			//重复创建 仍然为200
			return errors.New(fmt.Sprintf("init ecr user error %v", err))
		}

		_, err = p.kubeClient.CoreV1().Secrets(DefaultEcrCredentialNamespace).Create(context.TODO(), &newSecret, v1.CreateOptions{})

		if err != nil {
			klog.Errorf("create auth info secret error %v", err)
			return err
		}
		return err

	} else if err != nil {
		klog.Infof("get auth info secret error %v", err)
		return err
	} else {
		//临时账户已经存在  密码为空 不更新
		opts, err := getEcrOpts()
		if err != nil {
			return err
		}

		var user ecrUser
		err = json.Unmarshal(get.Data["auth"], &user)
		if err != nil {
			return err
		}

		//更新secret使用
		secret := setUser(user, secretList)

		if len(secret) != 0 {
			rgs, _ := json.Marshal(secret)
			get.Data["ecrRegistry"] = rgs
		} else {
			return errors.New("RgAuthInfo is empty")
		}

		reCreateUrl := genUrl(opts.ApiGateway, UpdateEcrUserUri)
		reCreateHttpClient := newHttpClient(reCreateUrl, opts.AccessKey, opts.SecretKey)

		updateRequest := GenUserUpdateRequest(user.Username, secretList)
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
					return errors.New(fmt.Sprintf("recreate ecr user error %v", err))
				}
				//创建成功 更新secret的auth值
				marshal, _ := json.Marshal(user)
				get.Data["auth"] = marshal

			} else {
				return errors.New(fmt.Sprintf("update ecr user error %v", err))
			}
		}

		_, err = p.kubeClient.CoreV1().Secrets(DefaultEcrCredentialNamespace).Update(context.TODO(), get, v1.UpdateOptions{})

		if err != nil {
			klog.Errorf("recreate auth info secret error %v", err)
			return err
		}
		return err
	}
}
func setcfDefault(cm configInfo, version string) configInfo {
	if reflect.DeepEqual(cm, configInfo{}) {
		cm.Namespace = "default"
		cm.ServiceAccount = "default"
	}
	cm.Version = version
	return cm
}

// setUser  创建生成的secret
func setUser(user ecrUser, rg []DbAuth) []ecrSecret {
	secrets := []ecrSecret{}
	//如果为设置仓库信息 返回空切片
	if len(rg) == 0 {
		return secrets
	}
	//生成集群中的secret
	for _, i := range rg {
		secret := ecrSecret{
			Server: i.HarborServer,
			User:   user.User,
		}
		secrets = append(secrets, secret)
	}
	return secrets
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

func getEcrOpts() (*EcrCredentialOptions, error) {
	//getApiGateWay  readConfig
	config, err := config.TryLoadFromDisk()
	if err != nil {
		klog.Errorf("Failed to load configuration from disk", err)
		return nil, err
	} else {

		return config.EcrCredentialOptions, err
	}
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
	specials := "~=+%^*/()[]{}/!@#$?|"
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

func genJob(jobType string) (v12.Job, error) {
	i := int32(3)
	var backoffLimit *int32
	backoffLimit = &i
	lables := make(map[string]string)
	lables["app.kubernetes.io/name"] = "ecr-credential"
	lables["app.kubernetes.io/instance"] = "ecr-helper"

	opts, err := getEcrOpts()
	if err != nil {
		return v12.Job{}, err
	}
	var command []string
	cons := []corev1.Container{}
	newJob := v12.Job{}
	switch jobType {
	case "renew":

		command = append(command, "/test/renewJob")
		con := corev1.Container{
			Name:            "renew-job",
			Image:           opts.ReNewJobImage,
			ImagePullPolicy: "IfNotPresent",
			Command:         command,
		}

		cons = append(cons, con)

		newJob = v12.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      DefaultEcrCredentialReNewJobName + fmt.Sprintf(time.Now().Format("200601021504")),
				Namespace: DefaultEcrCredentialNamespace,
			},
			Spec: v12.JobSpec{
				BackoffLimit: backoffLimit,
				Template: corev1.PodTemplateSpec{

					ObjectMeta: v1.ObjectMeta{
						Name:   DefaultEcrCredentialReNewJobName,
						Labels: lables,
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: DefaultEcrCredentialServiceAccountName,
						RestartPolicy:      "OnFailure",
						Containers:         cons,
					},
				},
			},
		}
		return newJob, nil
	case "clear":
		command = append(command, "/test/clearJob")
		con := corev1.Container{
			Name:            "clear-job",
			Image:           opts.ClearJobImage,
			ImagePullPolicy: "IfNotPresent",
			Command:         command,
		}

		cons = append(cons, con)

		newJob = v12.Job{
			ObjectMeta: v1.ObjectMeta{
				Name:      DefaultEcrCredentialClearJobName + fmt.Sprintf(time.Now().Format("200601021504")),
				Namespace: DefaultEcrCredentialNamespace,
			},
			Spec: v12.JobSpec{
				BackoffLimit: backoffLimit,
				Template: corev1.PodTemplateSpec{

					ObjectMeta: v1.ObjectMeta{
						Name:   DefaultEcrCredentialClearJobName,
						Labels: lables,
					},
					Spec: corev1.PodSpec{
						ServiceAccountName: DefaultEcrCredentialServiceAccountName,
						RestartPolicy:      "OnFailure",
						Containers:         cons,
					},
				},
			},
		}
		return newJob, nil
	default:
		return newJob, errors.New("job type error ")
	}
}

func GenUserCreateRequest(user ecrUser, dbAuth []DbAuth) ecrUser {
	marshal, _ := json.Marshal(dbAuth)

	return ecrUser{
		User: user.User,
		Auth: string(marshal),
	}

}
func GenUserUpdateRequest(username string, dbAuth []DbAuth) ecrUser {
	marshal, _ := json.Marshal(dbAuth)

	return ecrUser{
		User: User{Username: username},
		Auth: string(marshal),
	}
}

// GenUserDeleteRequest 预留
func GenUserDeleteRequest(user ecrUser) ecrUser {
	return user
}
