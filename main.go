package main

import (
	"encoding/json"
	"fmt"
	"os"

	"yummy-go.com/m/v2/scir"
)

func main() {
	projectjson, err := os.ReadFile("./files/project.json")
	if err != nil {
		fmt.Printf("open file error: %v\n", err)
	}
	fmt.Println("loaded file:", string(projectjson))
	var project scir.Project
	if err := json.Unmarshal(projectjson, &project); err != nil {
		fmt.Printf("unmarshal error: %v\n", err)
		return
	}
	fmt.Println("project:", project)
	marshaled, err := json.Marshal(project)
	if err != nil {
		fmt.Printf("marshal error: %v\n", err)
		return
	}
	fmt.Println("marshaled:", string(marshaled))
}
