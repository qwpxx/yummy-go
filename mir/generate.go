package mir

import "yummy-go.com/m/v2/compiler"

func GenerateMir(ast compiler.Program) Program {
	return generateMirFromAstProgram(ast)
}

func generateMirFromAstProgram(_ast compiler.Program) Program {
	return Program{}
}
