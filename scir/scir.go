package scir

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/google/uuid"
)

type Scir struct {
	Assets        map[string][]byte
	Ir            Project
	IdTable       IdTable
	EditingTarget *Target
	StageTarget   *Target
}

func (s *Scir) SetEditingTarget(name string) {
	for _, target := range s.Ir.Targets {
		if target.Name == name {
			s.EditingTarget = &target
			return
		}
	}
	newTarget := NewTarget(name, []Costume{
		// use the first costume of stage by default
		s.StageTarget.Costumes[0],
	})
	s.Ir.Targets = append(s.Ir.Targets, newTarget)
	s.EditingTarget = &newTarget
}

func (s *Scir) InsertBlock(block *Block) string {
	blockUuid := uuid.NewString()
	s.EditingTarget.Blocks[blockUuid] = block
	return blockUuid
}

func (s *Scir) InsertBlockWithUuid(uuid string, block *Block) {
	s.EditingTarget.Blocks[uuid] = block
}

func (s *Scir) ConnectBlocks(blockUuid, nextBlockUuid string) {
	s.EditingTarget.Blocks[blockUuid].Next = &nextBlockUuid
	s.EditingTarget.Blocks[nextBlockUuid].Parent = &blockUuid
}

func (s *Scir) CopyBlocks(blockUuids []string) []string {
	newBlockUuids := make([]string, 0)
	for _, uuid := range blockUuids {
		block := s.EditingTarget.Blocks[uuid]
		newBlockUuid := s.InsertBlock(block)
		newBlockUuids = append(newBlockUuids, newBlockUuid)
	}
	return newBlockUuids
}

func (s *Scir) SetInput(blockUUid, input string, inputBlock *Block) {
	inputUuid := s.InsertBlock(inputBlock)
	blockInput := BlockInput(inputUuid)
	inputBlock.Parent = &blockUUid
	s.EditingTarget.Blocks[blockUUid].Inputs[input] = MaybeShadowedInput{
		Type:          Nonshadow,
		ObscuredInput: &blockInput,
	}
}

func (s *Scir) SetShadowInput(blockUUid, input string, inputBlock *Block) {
	inputUuid := s.InsertBlock(inputBlock)
	blockInput := BlockInput(inputUuid)
	inputBlock.Parent = &blockUUid
	s.EditingTarget.Blocks[blockUUid].Inputs[input] = MaybeShadowedInput{
		Type:          Shadow,
		ShadowedInput: &blockInput,
	}
}

func parseProjectJson(project []byte) (Project, error) {
	var info Project
	err := json.Unmarshal(project, &info)
	return info, err
}

func LoadSb3(path string, idTablePath *string) (Scir, error) {
	zipReader, err := zip.OpenReader(path)
	if err != nil {
		return Scir{}, err
	}
	assets := make(map[string][]byte)
	var ir *Project = nil
	for _, file := range zipReader.File {
		fileCloser, err := file.Open()
		if err != nil {
			return Scir{}, err
		}
		defer fileCloser.Close()
		content, err := io.ReadAll(fileCloser)
		if err != nil {
			return Scir{}, err
		}
		if file.Name == "project.json" {
			info, err := parseProjectJson(content)
			if err != nil {
				return Scir{}, err
			}
			ir = &info
			continue
		}
		assets[file.Name] = content
	}
	idTable := NewIdTable()
	if idTablePath != nil {
		theIdTable, err := OpenIdTable(*idTablePath)
		if err == nil {
			idTable = theIdTable
		}
	}
	for _, target := range ir.Targets {
		if target.IsStage {
			return Scir{
				Assets:        assets,
				Ir:            *ir,
				IdTable:       idTable,
				EditingTarget: nil,
				StageTarget:   &target,
			}, nil
		}
	}
	return Scir{}, fmt.Errorf("project.json: missing target `stage`")
}

func ExportSb3(path, idTablePath string, sb3 Scir) error {
	idTableFile, err := os.Create(idTablePath)
	if err != nil {
		return err
	}
	defer idTableFile.Close()
	idTableContent, err := json.Marshal(sb3.IdTable)
	if err != nil {
		return err
	}
	idTableFile.Write(idTableContent)
	zipFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()
	for assetName, asset := range sb3.Assets {
		assetFile, err := writer.Create(assetName)
		if err != nil {
			return err
		}
		assetFile.Write(asset)
	}
	projectJsonFile, err := writer.Create("project.json")
	if err != nil {
		return err
	}
	jsonContent, err := json.Marshal(sb3.Ir)
	if err != nil {
		return err
	}
	projectJsonFile.Write(jsonContent)
	return nil
}
