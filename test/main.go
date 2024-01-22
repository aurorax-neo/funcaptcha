package main

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
