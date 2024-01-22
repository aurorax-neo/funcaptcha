package funcaptcha

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
)

type arkReq struct {
	arkURL     string
	arkBx      string
	arkHeader  http.Header
	arkBody    url.Values
	arkCookies []*http.Cookie
	userAgent  string
}

type kvPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type cookie struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Expires string `json:"expires"`
}
type postBody struct {
	Params []kvPair `json:"params"`
}
type request struct {
	URL      string   `json:"url"`
	Headers  []kvPair `json:"headers,omitempty"`
	PostData postBody `json:"postData,omitempty"`
	Cookies  []cookie `json:"cookies,omitempty"`
}
type entry struct {
	StartedDateTime string  `json:"startedDateTime"`
	Request         request `json:"request"`
}
type logData struct {
	Entries []entry `json:"entries"`
}
type HARData struct {
	Log logData `json:"log"`
}

func (s *Solver) GetOpenAIToken(key string, puid string) (string, error) {
	token, err := s.sendRequest(key, "", puid)
	return token, err
}

func (s *Solver) GetOpenAITokenWithBx(key string, bx string, puid string) (string, error) {
	token, err := s.sendRequest(key, getBdaWitBx(bx), puid)
	return token, err
}

func (s *Solver) sendRequest(key string, bda string, puid string) (string, error) {
	if len(s.Arks[key]) == 0 {
		return "", errors.New("a valid HAR file with key " + key + " required")
	}
	var tmpArk = &s.Arks[key][0]
	s.Arks[key] = append(s.Arks[key][1:], s.Arks[key][0])
	if tmpArk == nil || tmpArk.arkBx == "" || len(tmpArk.arkBody) == 0 || len(tmpArk.arkHeader) == 0 {
		return "", errors.New("a valid HAR file required")
	}
	if bda == "" {
		bda = s.getBDA(tmpArk)
	}
	tmpArk.arkBody.Set("bda", base64.StdEncoding.EncodeToString([]byte(bda)))
	tmpArk.arkBody.Set("rnd", strconv.FormatFloat(rand.Float64(), 'f', -1, 64))
	req, _ := http.NewRequest(http.MethodPost, tmpArk.arkURL, strings.NewReader(tmpArk.arkBody.Encode()))
	req.Header = tmpArk.arkHeader.Clone()
	arkURLIns, _ := url.Parse(tmpArk.arkURL)
	(*s.client).GetCookieJar().SetCookies(arkURLIns, tmpArk.arkCookies)
	if puid != "" {
		req.Header.Set("cookie", "_puid="+puid+";")
	}
	resp, err := (*s.client).Do(req)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	if resp.StatusCode != 200 {
		return "", errors.New("status code " + resp.Status)
	}

	type arkoseResponse struct {
		Token string `json:"token"`
	}
	var arkose arkoseResponse
	err = json.NewDecoder(resp.Body).Decode(&arkose)
	if err != nil {
		return "", err
	}
	// Check if rid is empty
	if !strings.Contains(arkose.Token, "sup=1|rid=") {
		return arkose.Token, errors.New("captcha required")
	}
	return arkose.Token, nil
}

//goland:noinspection SpellCheckingInspection
func (s *Solver) getBDA(arkReq *arkReq) string {
	var bx = arkReq.arkBx
	if bx == "" {
		bx = fmt.Sprintf(bxTemplate,
			getF(),
			getN(),
			getWh(),
			webglExtensions,
			getWebglExtensionsHash(),
			webglRenderer,
			webglVendor,
			webglVersion,
			webglShadingLanguageVersion,
			webglAliasedLineWidthRange,
			webglAliasedPointSizeRange,
			webglAntialiasing,
			webglBits,
			webglMaxParams,
			webglMaxViewportDims,
			webglUnmaskedVendor,
			webglUnmaskedRenderer,
			webglVsfParams,
			webglVsiParams,
			webglFsfParams,
			webglFsiParams,
			getWebglHashWebgl(),
			s.initVer,
			s.initHex,
			getFe(),
			getIfeHash(),
		)
	} else {
		re := regexp.MustCompile(`"key":"n","value":"\S*?"`)
		bx = re.ReplaceAllString(bx, `"key":"n","value":"`+getN()+`"`)
	}
	bt := getBt()
	bw := getBw(bt)
	return Encrypt(bx, arkReq.userAgent+bw)
}

func getBt() int64 {
	return time.Now().UnixMicro() / 1000000
}

func getBw(bt int64) string {
	return strconv.FormatInt(bt-(bt%21600), 10)
}

func getBdaWitBx(bx string) string {
	bt := getBt()
	bw := getBw(bt)
	return Encrypt(bx, bv+bw)
}