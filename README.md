# Arkose funcaptcha

本项目为改自于
[funcaptcha](https://github.com/acheong08/funcaptcha)
[GenerateArkose](https://github.com/Ink-Osier/GenerateArkose)

### Warning

本项目不保证生成的Arkose的可用性以及使用本项目不会导致封号等一系列问题，使用本项目造成的一切后果由使用者自行承担。

### 安装

```
go get -u github.com/aurorax-neo/funcaptcha
```

### 使用方法

1. 新建`harPool`文件夹
2. 根据Ninja项目中的[Har获取说明](https://github.com/gngpp/ninja/blob/main/doc/readme_zh.md#arkoselabs)
   下载Har文件至`harPool`文件夹下

### 使用示例

```go
import (
	"fmt"
	"github.com/aurorax-neo/funcaptcha"
)

func main() {
	solver := funcaptcha.NewSolver()
	funcaptcha.WithHarPool(solver)
	//https://tcr9i.chat.openai.com/fc/gt2/public_key/35536E1E-65B4-4D96-9D97-6ADB7EFF8147
	// surl 对应har文件中如上包含 /fc/gt2/public_key/ 前面的部分
	sUrl := "https://share.wendaalpha.net"
	// publicKey 对应har文件中如上包含 /fc/gt2/public_key/ 后面的部分
	publicKey := "35536E1E-65B4-4D96-9D97-6ADB7EFF8147"
	key := funcaptcha.GetKey(sUrl, publicKey)
	token, _ := solver.GetOpenAIToken(key, "")
	fmt.Println(token)
}
```
