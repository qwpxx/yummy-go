package scir

import (
	"archive/zip"
	"encoding/json"
	"io"
)

type Scir struct {
	assets map[string][]byte
	ir     Project
}

func parseProjectJson(project []byte) (Project, error) {
	var info Project
	err := json.Unmarshal(project, &info)
	return info, err
}

func LoadSb3(path string) (Scir, error) {
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
	return Scir{
		assets: assets,
		ir:     *ir,
	}, nil
}
