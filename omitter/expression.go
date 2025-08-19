package omitter

import (
	"fmt"
	"strconv"

	"yummy-go.com/m/v2/mir"
	"yummy-go.com/m/v2/scir"
)

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
