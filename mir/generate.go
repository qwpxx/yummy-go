package mir

import "yummy-go.com/m/v2/frontend"

func GenerateMir(ast frontend.Program) Program {
	return generateMirFromAstProgram(ast)
}

func generateMirFromAstProgram(_ast frontend.Program) Program {
	return Program{}
}
