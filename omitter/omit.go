package omitter

import (
	"encoding/json"
	"fmt"
	"strconv"

	"yummy-go.com/m/v2/mir"
	"yummy-go.com/m/v2/scir"
)

type Omitter struct {
	scir *scir.Scir
}

func New(scir *scir.Scir) Omitter {
	return Omitter{
		scir,
	}
}

func (s *Omitter) SetTarget(name string) {
	s.scir.SetEditingTarget(name)
}

func (s *Omitter) Omit(mir mir.Program) error {
	for _, declaration := range mir.Declarations {
		if err := s.OmitDeclaration(declaration); err != nil {
			return err
		}
	}
	return nil
}

func (s *Omitter) OmitDeclaration(declaration mir.Declaration) error {
	switch declaration := declaration.(type) {
	case *mir.GlobalDeclaration:
	case *mir.FunctionDeclaration:
		s.OmitFunction(declaration)
	}
	return nil
}

func (s *Omitter) OmitFunction(function *mir.FunctionDeclaration) error {
	procedureHead := scir.Block{
		Opcode:   "procedures_definition",
		Fields:   make(map[string]scir.Field),
		Inputs:   make(map[string]scir.MaybeShadowedInput),
		Shadow:   false,
		TopLevel: true,
	}
	procedureHeadUuid := s.scir.InsertBlock(&procedureHead)
	warpString := strconv.FormatBool(function.Warp)
	procedurePrototype := scir.Block{
		Opcode:   "procedures_prototype",
		Inputs:   make(map[string]scir.MaybeShadowedInput),
		Fields:   make(map[string]scir.Field),
		Shadow:   true,
		TopLevel: false,
		Mutation: &scir.Mutation{
			TagName:     "mutation",
			Children:    make([]any, 0),
			ProcCode:    &function.ProcCode,
			ArgumentIds: &function.ArgumentIds,
			Warp:        &warpString,
		},
	}
	procedurePrototypeUuid := s.scir.InsertBlock(&procedurePrototype)
	s.scir.SetShadowInput(procedureHeadUuid, "custom_block", &procedurePrototype)
	argumentNames := make([]string, 0)
	argumentDefaults := make([]string, 0)
	for argumentName, argumentDeclaration := range function.Arguments {
		for _, slot := range argumentDeclaration.TypeView.Slots {
			argName := fmt.Sprintf("(%s)%d", argumentName, slot.Index)
			argument := scir.Block{
				Opcode: "argument_reporter_string_number",
				Inputs: make(map[string]scir.MaybeShadowedInput),
				Fields: map[string]scir.Field{
					"VALUE": {
						Value: argName,
					},
				},
				Shadow:   true,
				TopLevel: false,
			}
			s.scir.InsertBlock(&argument)
			s.scir.SetShadowInput(
				procedurePrototypeUuid,
				slot.Uuid,
				&argument,
			)
			argumentNames = append(argumentNames, argName)
			argumentDefaults = append(argumentDefaults, "")
		}
	}
	argumentNamesBytes, _ := json.Marshal(argumentNames)
	argumentDefaultsBytes, _ := json.Marshal(argumentDefaults)
	argumentNamesString := string(argumentNamesBytes)
	argumentDefaultsString := string(argumentDefaultsBytes)
	procedurePrototype.Mutation.ArgumentNames = &argumentNamesString
	procedurePrototype.Mutation.ArgumentDefaults = &argumentDefaultsString
	bodyStartUuid := procedureHeadUuid
	bodyUuids, err := s.OmitBlock(function.Body)
	if err != nil {
		return err
	}
	if len(bodyUuids) > 0 {
		s.scir.ConnectBlocks(bodyStartUuid, bodyUuids[0])
	}
	return nil
}

func (s *Omitter) OmitBlock(block mir.Block) ([]string, error) {
	blockUuids := make([]string, 0)
	for _, statement := range block.Statements {
		statementUuids, err := s.OmitStatement(statement)
		if err != nil {
			return nil, err
		}
		blockUuids = append(blockUuids, statementUuids...)
	}
	for i := 0; i < len(blockUuids)-1; i += 1 {
		s.scir.ConnectBlocks(blockUuids[i], blockUuids[i+1])
	}
	return blockUuids, nil
}

func (s *Omitter) OmitStatement(statement mir.Statement) ([]string, error) {
	return []string{
		s.scir.InsertBlock(&scir.Block{
			Opcode:   "looks_show",
			Inputs:   make(map[string]scir.MaybeShadowedInput),
			Fields:   make(map[string]scir.Field),
			TopLevel: false,
			Shadow:   false,
		}),
	}, nil
}
