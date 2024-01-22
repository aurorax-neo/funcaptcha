package funcaptcha

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/aurorax-neo/funcaptcha/logger"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
)

func toJSON(data interface{}) string {
	str, _ := json.Marshal(data)
	return string(str)
}

func jsonToForm(data string) string {
	// Unmarshal into map
	var formData map[string]interface{}
	_ = json.Unmarshal([]byte(data), &formData)
	// Use reflection to convert to form data
	var form = url.Values{}
	for k, v := range formData {
		form.Add(k, fmt.Sprintf("%v", v))
	}
	return form.Encode()
}

func (sn *Session) DownloadChallenge(urls []string, b64 bool) ([]string, error) {
	var b64Images = make([]string, len(urls))
	for i, link := range urls {
		req, _ := http.NewRequest(http.MethodGet, link, nil)
		req.Header = headers
		resp, err := (*sn.Client).Do(req)
		if err != nil {
			return nil, err
		}
		_ = resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("status code %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		// Figure out filename from URL
		urlPaths := strings.Split(link, "/")
		if !b64 {
			filename := strings.Split(urlPaths[len(urlPaths)-1], "?")[0]
			if filename == "image" {
				filename = fmt.Sprintf("image_%sn.png", getTimeStamp())
			}
			err = os.WriteFile(filename, body, 0644)
			if err != nil {
				return nil, err
			}
		} else {
			// base64 encode body
			b64Images[i] = base64.StdEncoding.EncodeToString(body)
		}
	}
	return b64Images, nil
}

func getTimeStamp() string {
	return fmt.Sprintf("%d", time.Now().UnixNano()/int64(time.Millisecond))
}

func getRequestId(sessionId string) string {
	pwd := fmt.Sprintf("REQUESTED%sID", sessionId)
	return Encrypt(`{"sc":[147,307]}`, pwd)
}

// GetHashStr 生成指定长度的哈希字符串
func GetHashStr(str string, len int) string {
	// 创建一个 SHA-256 哈希对象
	hasher := sha256.New()
	if str == "" {
		logger.Logger.Error("str is nil")
		panic("str is nil")
	}
	// 写入要哈希的数据
	hasher.Write([]byte(str))
	// 计算哈希值并将其转换为字符串形式
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	// 生成固定长度的哈希值（假设要截取前len个字符）
	fixedLengthHash := hashString[:len]
	return fixedLengthHash
}

func GetKey(sUrl string, publicKey string) string {
	return GetHashStr(sUrl+publicKey, 32)
}
