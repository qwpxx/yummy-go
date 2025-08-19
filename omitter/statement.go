package omitter

import (
	"fmt"

	"yummy-go.com/m/v2/mir"
	"yummy-go.com/m/v2/scir"
)

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
