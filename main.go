package main

import (
	"encoding/json"
	"fmt"

	"yummy-go.com/m/v2/mir"
	"yummy-go.com/m/v2/omitter"
	"yummy-go.com/m/v2/scir"
)

func main() {
	sb3file, err := scir.LoadSb3("./files/empty.sb3", nil)
	if err != nil {
		fmt.Println("load sb3 error:", err)
		return
	}

	declaration := mir.DeclareStatement{}
	allocator := mir.NewSlotAllocator()

	mir := mir.Program{
		Declarations: []mir.Declaration{
			&mir.FunctionDeclaration{
				Name:      "Hello",
				Arguments: map[string]mir.Argument{},
				ReturnTypeView: mir.TypeView{
					Type:   &mir.StringType{},
					Offset: 0,
					Slots:  allocator.AllocN(1),
				},
				ProcCode:    "Hello()",
				ArgumentIds: "[]",
				Body: mir.Block{
					Statements: []mir.Statement{
						&declaration,
						&mir.AssignStatement{
							Declaration: &declaration,
							Value: &mir.LiteralExpression{
								Literal:     "World",
								LiteralType: &mir.StringType{},
							},
						},
						&mir.ReturnStatement{
							Value: &mir.VariableExpression{
								Declaration: &declaration,
							},
						},
					},
				},
			},
		},
	}

	omitter := omitter.New(&sb3file)
	omitter.SetTarget("Stage")
	if err := omitter.Omit(mir); err != nil {
		fmt.Println("omit error:", err)
		return
	}

	projectjson, _ := json.Marshal(sb3file.Ir)
	fmt.Println("project.json:", string(projectjson))

	if err := scir.ExportSb3("./files/output.sb3", "./files/output.sb3.json", sb3file); err != nil {
		fmt.Println("export error:", err)
		return
	}
	fmt.Println("successfully exported!")
}
