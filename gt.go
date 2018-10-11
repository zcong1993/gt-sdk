package gt

import (
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Config struct {
	Protocol     string
	ApiServer    string
	ValidatePath string
	RegisterPath string
	Timeout      time.Duration
	NewCaptcha   bool
	JsonFormat   string

	GeeTestID  string
	GeeTestKey string
}

var DefaultConfig = &Config{
	Protocol:     "http://",
	ApiServer:    "api.geetest.com",
	ValidatePath: "/validate.php",
	RegisterPath: "/register.php",
	Timeout:      time.Second * 2,
	NewCaptcha:   true,
	JsonFormat:   "1",
	// GeeTestID test id, replace it to yourself
	GeeTestID:  "683898b124098fd661657f731db857aa",
	GeeTestKey: "3f204e4be8c779614b2ad5caf5a6de8e",
}

type Gt struct {
	client *http.Client
	config Config
}

type RegisterResp struct {
	Success    int    `json:"success"`
	Challenge  string `json:"challenge"`
	Gt         string `json:"gt"`
	NewCaptcha bool   `json:"new_captcha"`
}

type ValidateForm struct {
	Challenge string `json:"challenge"`
	Validate  string `json:"validate"`
	Seccode   string `json:"seccode"`
}

func getMd5(in string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(in)))
}

func NewCt(config Config) *Gt {
	client := &http.Client{Timeout: config.Timeout}

	return &Gt{client: client, config: config}
}

func (gt *Gt) genChallenge() string {
	return getMd5(fmt.Sprintf("%d", rand.Intn(90))) + getMd5(fmt.Sprintf("%d", rand.Intn(90)))[:2]
}

func (gt *Gt) Register(clientType, ipAddress string) (*RegisterResp, error) {
	u := gt.config.Protocol + gt.config.ApiServer + gt.config.RegisterPath

	if clientType == "" {
		clientType = "unknown"
	}
	if ipAddress == "" {
		ipAddress = "unknown"
	}

	qs := url.Values{}
	qs.Add("gt", gt.config.GeeTestID)
	qs.Add("json_format", gt.config.JsonFormat)
	qs.Add("client_type", clientType)
	qs.Add("ip_address", ipAddress)

	u += "?" + qs.Encode()

	req, err := http.NewRequest(http.MethodGet, u, nil)

	if err != nil {
		return nil, err
	}

	resp, err := gt.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	ch := gjson.GetBytes(data, "challenge").String()

	success := 1

	if ch == "" {
		ch = gt.genChallenge()
		success = 0
	} else {
		ch = getMd5(ch + gt.config.GeeTestKey)
	}

	return &RegisterResp{
		Success:    success,
		Challenge:  ch,
		Gt:         gt.config.GeeTestID,
		NewCaptcha: gt.config.NewCaptcha,
	}, nil
}

func (gt *Gt) Validate(f *ValidateForm, fallback bool) (bool, error) {
	if fallback {
		if getMd5(f.Challenge) == f.Validate {
			return true, nil
		}
		return false, nil
	}

	hash := gt.config.GeeTestKey + "geetest" + f.Challenge
	if f.Validate != getMd5(hash) {
		return false, nil
	}

	u := gt.config.Protocol + gt.config.ApiServer + gt.config.ValidatePath

	form := url.Values{}
	form.Add("gt", gt.config.GeeTestID)
	form.Add("seccode", f.Seccode)
	form.Add("json_format", gt.config.JsonFormat)

	resp, err := gt.client.PostForm(u, form)

	if err != nil {
		return false, err
	}

	defer resp.Body.Close()

	d, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return false, err
	}

	code := gjson.GetBytes(d, "seccode").String()

	if code == "" {
		return false, errors.New("api server error")
	}

	return getMd5(f.Seccode) == code, nil
}
