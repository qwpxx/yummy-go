package scir

import (
	"encoding/json"
	"os"
)

type IdTable struct {
	Ids map[string]IdUsage
}

type IdUsage struct {
	For            string
	Uuid           string
	RawDeclaration string
}

func (s *IdTable) UpdateId(id string, usage IdUsage) {
	s.Ids[id] = usage
}

func (s *IdTable) LookupId(id string) *IdUsage {
	if usage, ok := s.Ids[id]; ok {
		return &usage
	}
	return nil
}

func NewIdTable() IdTable {
	return IdTable{
		Ids: make(map[string]IdUsage),
	}
}

func OpenIdTable(path string) (IdTable, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return IdTable{}, err
	}
	var idTable IdTable
	if err := json.Unmarshal(content, &idTable); err != nil {
		return IdTable{}, err
	}
	return idTable, nil
}
