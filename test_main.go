package main

import (
	"fmt"

	"yummy-go.com/m/v2/mir"
	"yummy-go.com/m/v2/omitter"
	"yummy-go.com/m/v2/scir"
)

func TestMain() {
	sb3file, err := scir.LoadSb3("./files/empty.sb3", nil)
	if err != nil {
		fmt.Println("load sb3 error:", err)
		return
	}

	allocator := mir.NewSlotAllocator()

	declaration := mir.DeclareStatement{
		Name: "Var",
		TypeView: mir.TypeView{
			Type: &mir.ArrayType{
				Inner: &mir.StringType{},
				N:     2,
			},
			Slots: allocator.AllocN(2),
		},
	}
	slots := allocator.AllocN(2)
	arg1 := mir.Argument{
		TypeView: mir.TypeView{
			Type: &mir.ArrayType{
				Inner: &mir.StringType{},
				N:     2,
			},
			Offset: 0,
			Slots:  slots,
		},
		Name: "world",
	}

	function := mir.FunctionDeclaration{
		Name: "Hello",
		Arguments: []mir.Argument{
			arg1,
		},
		ReturnTypeView: mir.TypeView{
			Type: &mir.ArrayType{
				Inner: &mir.StringType{},
				N:     2,
			},
			Offset: 0,
			Slots:  allocator.AllocN(2),
		},
		ProcCode:    "Hello(world: %s %s )",
		ArgumentIds: fmt.Sprintf("[\"%s\",\"%s\"]", slots[0].Uuid, slots[1].Uuid),
		StackSize:   2,
	}
	function.Body = mir.Block{
		Statements: []mir.Statement{
			&declaration,
			&mir.AssignStatement{
				Acessor: &mir.VariableAcessor{
					Declaration: &declaration,
				},
				Value: &mir.CallExpression{
					Function: &function,
					Arguments: []mir.Expression{
						&mir.AcessorExpression{
							Acessor: &mir.VariableAcessor{
								Declaration: &declaration,
							},
						},
					},
				},
			},
			&mir.ReturnStatement{
				Value: &mir.AcessorExpression{
					Acessor: &mir.VariableAcessor{
						Declaration: &declaration,
					},
				},
			},
		},
	}

	mir := mir.Program{
		Declarations: []mir.Declaration{
			&function,
		},
	}

	omitter := omitter.New(&sb3file)
	omitter.SetTarget("Stage")
	if err := omitter.Omit(mir); err != nil {
		fmt.Println("omit error:", err)
		return
	}

	//projectjson, _ := json.Marshal(sb3file.Ir)
	//fmt.Println("project.json:", string(projectjson))

	if err := scir.ExportSb3("./files/output.sb3", "./files/output.sb3.json", sb3file); err != nil {
		fmt.Println("export error:", err)
		return
	}
	fmt.Println("successfully exported!")
}
