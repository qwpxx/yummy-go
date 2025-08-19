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

func (s *Omitter) OmitStatement(statement mir.Statement) ([]string, error) {
	switch statement := statement.(type) {
	case *mir.DeclareStatement:
		return []string{}, nil
	case *mir.AssignStatement:
		blockUuids := make([]string, 0)
		exprUuids, err := s.OmitExpression(statement.Value, &blockUuids)
		if err != nil {
			return nil, err
		}
		acessor, typeView, err := s.OmitAcessor(statement.Acessor, &blockUuids)
		if err != nil {
			return nil, err
		}
		acessorSize := typeView.Type.GetSize()
		if acessorSize == nil {
			return nil, fmt.Errorf("cannot assign to a dyn-sized array")
		}
		if uint(len(exprUuids)) != *acessorSize {
			return nil, fmt.Errorf("type not fit")
		}
		acessorNeedCopy := false
		for idx, exprUuid := range exprUuids {
			if acessorNeedCopy {
				acessor = s.scir.CopyBlocks(acessor)
			} else {
				acessorNeedCopy = true
			}
			block := scir.Block{
				Opcode: "data_replaceitemoflist",
				Inputs: make(map[string]scir.MaybeShadowedInput),
				Fields: map[string]scir.Field{
					"LIST": {
						Value: "_Stack",
						Id:    &s.stackUuid,
					},
				},
			}
			blockUuid := s.scir.InsertBlock(&block)
			substractBlock := scir.Block{
				Opcode: "operator_subtract",
				Fields: make(map[string]scir.Field),
				Inputs: map[string]scir.MaybeShadowedInput{
					"NUM2": {
						Type: scir.Shadow,
						ShadowedInput: &scir.NumberalInput{
							Type:  scir.InputNumber,
							Value: float64(idx),
						},
					},
				},
			}
			substractBlockUuid := s.scir.InsertBlock(&substractBlock)
			s.scir.SetInput(substractBlockUuid, "NUM1", s.scir.EditingTarget.Blocks[acessor[0]])
			s.scir.SetInput(blockUuid, "INDEX", &substractBlock)
			s.scir.SetInput(blockUuid, "ITEM", s.scir.EditingTarget.Blocks[exprUuid])
			blockUuids = append(blockUuids, blockUuid)
		}
		return blockUuids, nil
	case *mir.ReturnStatement:
		blockUuids := make([]string, 0)
		exprUuids, err := s.OmitExpression(statement.Value, &blockUuids)
		if err != nil {
			return nil, err
		}
		slots := s.omittingFunction.ReturnTypeView.Slots
		if len(slots) != len(exprUuids) {
			return nil, fmt.Errorf("type not fit")
		}
		for idx, exprUuid := range exprUuids {
			block := scir.Block{
				Opcode: "data_setvariableto",
				Inputs: make(map[string]scir.MaybeShadowedInput),
				Fields: map[string]scir.Field{
					"VARIABLE": {
						Value: slots[idx].Uuid,
						Id:    &slots[idx].Uuid,
					},
				},
			}
			blockUuid := s.scir.InsertBlock(&block)
			s.scir.SetInput(blockUuid, "VALUE", s.scir.EditingTarget.Blocks[exprUuid])
			blockUuids = append(blockUuids, blockUuid)
		}
		return blockUuids, nil
	}
	return nil, fmt.Errorf("not implemented yet")
}

func (s *Omitter) OmitAcessor(acessor mir.Acessor, blockUuids *[]string) ([]string, mir.TypeView, error) {
	switch acessor := acessor.(type) {
	case *mir.VariableAcessor:
		typeView := acessor.Declaration.GetTypeView()
		stackLengthBlock := scir.Block{
			Opcode: "data_lengthoflist",
			Fields: map[string]scir.Field{
				"LIST": {
					Value: "_Stack",
					Id:    &s.stackUuid,
				},
			},
			Inputs: make(map[string]scir.MaybeShadowedInput),
		}
		stackLengthBlockUuid := s.scir.InsertBlock(&stackLengthBlock)
		if typeView.Offset == 0 {
			return []string{stackLengthBlockUuid}, typeView, nil
		}
		substractBlock := scir.Block{
			Opcode: "operator_subtract",
			Fields: make(map[string]scir.Field),
			Inputs: map[string]scir.MaybeShadowedInput{
				"NUM2": {
					Type: scir.Shadow,
					ShadowedInput: &scir.NumberalInput{
						Type:  scir.InputNumber,
						Value: float64(typeView.Offset),
					},
				},
			},
		}
		substractBlockUuid := s.scir.InsertBlock(&substractBlock)
		s.scir.SetInput(substractBlockUuid, "NUM1", &stackLengthBlock)
	}
	return nil, mir.TypeView{}, fmt.Errorf("not implemented yet")
}

func (s *Omitter) OmitExpression(expression mir.Expression, blockUuids *[]string) ([]string, error) {
	switch expression := expression.(type) {
	case *mir.LiteralExpression:
		var exprUuid string
		switch literal := expression.Literal.(type) {
		case float64:
			exprUuid = s.scir.InsertBlock(&scir.Block{
				Opcode: "operator_add",
				Fields: make(map[string]scir.Field),
				Inputs: map[string]scir.MaybeShadowedInput{
					"NUM1": {
						Type: scir.Shadow,
						ShadowedInput: &scir.NumberalInput{
							Type:  scir.InputNumber,
							Value: literal,
						},
					},
					"NUM2": {
						Type: scir.Shadow,
						ShadowedInput: &scir.NumberalInput{
							Type:  scir.InputNumber,
							Value: 0,
						},
					},
				},
			})
		case string:
			exprUuid = s.scir.InsertBlock(&scir.Block{
				Opcode: "operator_join",
				Fields: make(map[string]scir.Field),
				Inputs: map[string]scir.MaybeShadowedInput{
					"STRING1": {
						Type: scir.Shadow,
						ShadowedInput: &scir.StringInput{
							Type:  scir.InputString,
							Value: literal,
						},
					},
					"STRING2": {
						Type: scir.Shadow,
						ShadowedInput: &scir.StringInput{
							Type:  scir.InputString,
							Value: "",
						},
					},
				},
			})
		case bool:
			if literal {
				exprUuid = s.scir.InsertBlock(&scir.Block{
					Opcode: "operator_not",
					Fields: make(map[string]scir.Field),
					Inputs: make(map[string]scir.MaybeShadowedInput),
				})
			} else {
				exprUuid = s.scir.InsertBlock(&scir.Block{
					Opcode: "operator_and",
					Fields: make(map[string]scir.Field),
					Inputs: make(map[string]scir.MaybeShadowedInput),
				})
			}
		default:
			return nil, fmt.Errorf("unknown value type")
		}
		return []string{exprUuid}, nil
	case *mir.AcessorExpression:
		exprUuids := make([]string, 0)
		acessor, typeView, err := s.OmitAcessor(expression.Acessor, blockUuids)
		if err != nil {
			return nil, err
		}
		acessorSize := typeView.Type.GetSize()
		if acessorSize == nil {
			return nil, fmt.Errorf("cannot deref a dyn-sized value")
		}
		acessorNeedCopy := false
		for idx := range *acessorSize {
			if acessorNeedCopy {
				acessor = s.scir.CopyBlocks(acessor)
			} else {
				acessorNeedCopy = true
			}
			block := scir.Block{
				Opcode: "data_itemoflist",
				Inputs: make(map[string]scir.MaybeShadowedInput),
				Fields: map[string]scir.Field{
					"LIST": {
						Value: "_Stack",
						Id:    &s.stackUuid,
					},
				},
			}
			blockUuid := s.scir.InsertBlock(&block)
			substractBlock := scir.Block{
				Opcode: "operator_subtract",
				Fields: make(map[string]scir.Field),
				Inputs: map[string]scir.MaybeShadowedInput{
					"NUM2": {
						Type: scir.Shadow,
						ShadowedInput: &scir.NumberalInput{
							Type:  scir.InputNumber,
							Value: float64(idx),
						},
					},
				},
			}
			substractBlockUuid := s.scir.InsertBlock(&substractBlock)
			s.scir.SetInput(substractBlockUuid, "NUM1", s.scir.EditingTarget.Blocks[acessor[0]])
			s.scir.SetInput(blockUuid, "INDEX", &substractBlock)
			exprUuids = append(exprUuids, blockUuid)
		}
		return exprUuids, nil
	case *mir.CallExpression:
		slots, err := s.OmitFunctionCall(expression, blockUuids)
		if err != nil {
			return nil, err
		}
		exprUuids := make([]string, 0)
		for _, slot := range slots {
			exprUuids = append(exprUuids, s.OmitVariable(slot))
		}
		return exprUuids, nil
	}
	return nil, fmt.Errorf("not implemented yet")
}

func (s *Omitter) OmitVariable(slot mir.Slot) string {
	return s.scir.InsertBlock(&scir.Block{
		Opcode: "data_variable",
		Inputs: make(map[string]scir.MaybeShadowedInput),
		Fields: map[string]scir.Field{
			"VARIABLE": {
				Value: slot.Uuid,
				Id:    &slot.Uuid,
			},
		},
	})
}

func (s *Omitter) OmitFunctionCall(call *mir.CallExpression, blockUuids *[]string) ([]mir.Slot, error) {
	warpString := strconv.FormatBool(call.Function.Warp)
	callBlock := scir.Block{
		Opcode: "procedures_call",
		Inputs: make(map[string]scir.MaybeShadowedInput),
		Fields: make(map[string]scir.Field),
		Mutation: &scir.Mutation{
			TagName:     "mutation",
			Children:    []any{},
			ProcCode:    &call.Function.ProcCode,
			ArgumentIds: &call.Function.ArgumentIds,
			Warp:        &warpString,
		},
	}
	if len(call.Function.Arguments) != len(call.Arguments) {
		return nil, fmt.Errorf("arguments mismatched")
	}
	callBlockUuid := s.scir.InsertBlock(&callBlock)
	idx2 := 0
	for _, argument := range call.Function.Arguments {
		exprUuids, err := s.OmitExpression(call.Arguments[idx2], blockUuids)
		if err != nil {
			return nil, err
		}
		if len(argument.TypeView.Slots) != len(exprUuids) {
			return nil, fmt.Errorf("type not fit")
		}
		for idx, slot := range argument.TypeView.Slots {
			exprBlock := s.scir.EditingTarget.Blocks[exprUuids[idx]]
			s.scir.SetInput(callBlockUuid, slot.Uuid, exprBlock)
		}
		idx2 += 1
	}
	*blockUuids = append(*blockUuids, callBlockUuid)
	return call.Function.ReturnTypeView.Slots, nil
}
