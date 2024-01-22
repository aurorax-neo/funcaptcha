package funcaptcha

import (
	"encoding/json"
	"github.com/aurorax-neo/funcaptcha/logger"
	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Solver struct {
	initVer string
	initHex string
	Arks    map[string][]arkReq
	client  *tlsclient.HttpClient
	HarPath string
}

type SolverArg func(*Solver)

func NewSolver(args ...SolverArg) *Solver {
	var (
		jar     = tlsclient.NewCookieJar()
		options = []tlsclient.HttpClientOption{
			tlsclient.WithTimeoutSeconds(360),
			tlsclient.WithClientProfile(profiles.Chrome_117),
			tlsclient.WithRandomTLSExtensionOrder(),
			tlsclient.WithNotFollowRedirects(),
			tlsclient.WithCookieJar(jar),
		}
		client, _ = tlsclient.NewHttpClient(tlsclient.NewNoopLogger(), options...)
	)
	s := &Solver{
		Arks:    make(map[string][]arkReq),
		client:  &client,
		initVer: "1.5.4",
		initHex: "cd12da708fe6cbe6e068918c38de2ad9",
	}
	pwd, _ := os.Getwd()
	s.HarPath = filepath.Join(pwd, "harPool")
	for _, arg := range args {
		arg(s)
	}
	return s
}

func WithInitVer(ver string) SolverArg {
	return func(s *Solver) {
		s.initVer = ver
	}
}

func WithProxy(proxy string) SolverArg {
	return func(s *Solver) {
		_ = (*s.client).SetProxy(proxy)
	}
}

func WithInitHex(hex string) SolverArg {
	return func(s *Solver) {
		s.initHex = hex
	}
}

func WithClient(client *tlsclient.HttpClient) SolverArg {
	return func(s *Solver) {
		s.client = client
	}
}

func WithHarData(harData HARData) SolverArg {
	return func(s *Solver) {
		for _, v := range harData.Log.Entries {
			if strings.Contains(v.Request.URL, "/fc/gt2/public_key/") {
				// 临时ArkReq
				var tmpArkReq arkReq
				tmpArkReq.arkURL = v.Request.URL
				if v.StartedDateTime == "" {
					logger.Logger.Error("Error: no arkose request!")
					continue
				}
				// header
				t, _ := time.Parse(time.RFC3339, v.StartedDateTime)
				bw := getBw(t.Unix())
				fallbackBw := getBw(t.Unix() - 21600)
				tmpArkReq.arkHeader = make(http.Header)
				for _, h := range v.Request.Headers {
					if !strings.EqualFold(h.Name, "content-length") && !strings.EqualFold(h.Name, "cookie") && !strings.HasPrefix(h.Name, ":") {
						tmpArkReq.arkHeader.Set(h.Name, h.Value)
						if strings.EqualFold(h.Name, "user-agent") {
							tmpArkReq.userAgent = h.Value
						}
					}
				}
				// cookies
				tmpArkReq.arkCookies = []*http.Cookie{}
				for _, cookie := range v.Request.Cookies {
					expire, _ := time.Parse(time.RFC3339, cookie.Expires)
					if expire.After(time.Now()) {
						tmpArkReq.arkCookies = append(tmpArkReq.arkCookies, &http.Cookie{Name: cookie.Name, Value: cookie.Value, Expires: expire.UTC()})
					}
				}
				var arkType string
				tmpArkReq.arkBody = make(url.Values)
				for _, p := range v.Request.PostData.Params {
					// arkBody except bda & rnd
					if p.Name == "bda" {
						cipher, err := url.QueryUnescape(p.Value)
						if err != nil {
							panic(err)
						}
						tmpArkReq.arkBx = Decrypt(cipher, tmpArkReq.userAgent+bw, tmpArkReq.userAgent+fallbackBw)
					} else if p.Name != "rnd" {
						query, err := url.QueryUnescape(p.Value)
						if err != nil {
							panic(err)
						}
						tmpArkReq.arkBody.Set(p.Name, query)
					}
				}

				parts := strings.SplitN(v.Request.URL, "/fc/gt2/public_key/", 2)
				key := GetKey(parts[0], parts[1])
				s.Arks[key] = append(s.Arks[key], tmpArkReq)
				if tmpArkReq.arkBx != "" {
					logger.Logger.Info("success read " + arkType + " arkose from " + v.Request.URL)
					break
				} else {
					logger.Logger.Error("failed to decrypt HAR file")
				}
			}
		}

	}
}

func WithHarPool(s *Solver) {
	var harPath []string
	err := filepath.Walk(s.HarPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(info.Name())
			if ext == ".har" {
				harPath = append(harPath, path)
			}
		}
		return nil
	})
	if err != nil {
		logger.Logger.Error("Error: please put HAR files in" + s.HarPath + " directory!")
	}
	for _, path := range harPath {
		file, err := os.ReadFile(path)
		if err != nil {
			return
		}
		var harFile HARData
		err = json.Unmarshal(file, &harFile)
		if err != nil {
			logger.Logger.Error("Error: not a HAR file!")
			return
		}
		WithHarData(harFile)(s)
	}

}
