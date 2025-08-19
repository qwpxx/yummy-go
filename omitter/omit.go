package omitter

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"yummy-go.com/m/v2/mir"
	"yummy-go.com/m/v2/scir"
)

const OmitMaxStackSize uint = 20

type Omitter struct {
	scir             *scir.Scir
	stackUuid        string
	omittingFunction *mir.FunctionDeclaration
}

func New(ctx *scir.Scir) Omitter {
	var stackUuid string
	theStackUuid := ctx.IdTable.LookupId("_Stack")
	if theStackUuid == nil {
		stackUuid = uuid.NewString()
		ctx.StageTarget.Lists[stackUuid] = scir.List{
			Name:  "_Stack",
			Value: make([]string, 0),
		}
	} else {
		stackUuid = theStackUuid.Uuid
	}
	return Omitter{
		scir:             ctx,
		stackUuid:        stackUuid,
		omittingFunction: nil,
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
		return s.OmitFunction(declaration)
	}
	return nil
}

func (s *Omitter) OmitFunction(function *mir.FunctionDeclaration) error {
	s.omittingFunction = function
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
	var procedurePrototypeUuid string
	usage := s.scir.IdTable.LookupId(function.ProcCode)
	if usage == nil {
		procedurePrototypeUuid = s.scir.InsertBlock(&procedurePrototype)
	} else {
		s.scir.InsertBlockWithUuid(usage.Uuid, &procedurePrototype)
		procedurePrototypeUuid = usage.Uuid
	}
	s.scir.IdTable.UpdateId(function.ProcCode, scir.IdUsage{
		For:            function.Name,
		Uuid:           procedurePrototypeUuid,
		RawDeclaration: function.Span.String(),
	})
	s.scir.SetShadowInput(procedureHeadUuid, "custom_block", &procedurePrototype)
	argumentNames := make([]string, 0)
	argumentDefaults := make([]string, 0)
	for _, argumentDeclaration := range function.Arguments {
		for _, slot := range argumentDeclaration.TypeView.Slots {
			argName := fmt.Sprintf("(%s)%d", argumentDeclaration.Name, slot.Index)
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
	bodyUuids, err := s.OmitBlock(function.Body)
	if err != nil {
		return err
	}
	bodyStartUuid := procedureHeadUuid
	if s.omittingFunction.StackSize > OmitMaxStackSize {
		return fmt.Errorf("reaches OmitMaxStackSize")
	}
	for range function.StackSize {
		blockUuid := s.scir.InsertBlock(&scir.Block{
			Opcode: "data_addtolist",
			Inputs: map[string]scir.MaybeShadowedInput{
				"ITEM": {
					Type: scir.Shadow,
					ShadowedInput: &scir.StringInput{
						Type:  scir.InputString,
						Value: "",
					},
				},
			},
			Fields: map[string]scir.Field{
				"LIST": {
					Value: "_Stack",
					Id:    &s.stackUuid,
				},
			},
		})
		s.scir.ConnectBlocks(bodyStartUuid, blockUuid)
		bodyStartUuid = blockUuid
	}
	if len(bodyUuids) > 0 {
		s.scir.ConnectBlocks(bodyStartUuid, bodyUuids[0])
		bodyStartUuid = bodyUuids[len(bodyUuids)-1]
	}
	blockUuids, _ := s.OmitFunctionCleanup()
	if len(blockUuids) > 0 {
		s.scir.ConnectBlocks(bodyStartUuid, blockUuids[0])
	}
	s.omittingFunction = nil
	return nil
}

func (s *Omitter) OmitFunctionCleanup() ([]string, error) {
	if s.omittingFunction == nil {
		return []string{}, fmt.Errorf("cannot omit cleanup blocks outside a function")
	}
	blockUuids := make([]string, 0)
	if s.omittingFunction.StackSize > OmitMaxStackSize {
		return []string{}, fmt.Errorf("reaches OmitMaxStackSize")
	}
	for range s.omittingFunction.StackSize {
		blockUuid := s.scir.InsertBlock(&scir.Block{
			Opcode: "data_deleteoflist",
			Inputs: map[string]scir.MaybeShadowedInput{
				"INDEX": {
					Type: scir.Shadow,
					ShadowedInput: &scir.StringInput{
						Type:  scir.InputString,
						Value: "last",
					},
				},
			},
			Fields: map[string]scir.Field{
				"LIST": {
					Value: "_Stack",
					Id:    &s.stackUuid,
				},
			},
		})
		blockUuids = append(blockUuids, blockUuid)
	}
	for i := 0; i < len(blockUuids)-1; i += 1 {
		s.scir.ConnectBlocks(blockUuids[i], blockUuids[i+1])
	}
	return blockUuids, nil
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
