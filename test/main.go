package main

import (
	"fmt"
	"github.com/aurorax-neo/funcaptcha"
)

func main() {
	solver := funcaptcha.NewSolver()
	funcaptcha.WithHarPool(solver)
	key := funcaptcha.GetKey("https://share.wendaalpha.net", "35536E1E-65B4-4D96-9D97-6ADB7EFF8147")
	token, _ := solver.GetOpenAIToken(key, "")
	fmt.Println(token)
}
