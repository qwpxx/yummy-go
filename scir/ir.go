package scir

import (
	"encoding/json"
	"strconv"
)

type Project struct {
	Targets    []Target  `json:"targets"`
	Monitors   []Monitor `json:"monitors"`
	Extensions []string  `json:"extensions"`
	Meta       Meta      `json:"meta"`
}

type Target struct {
	IsStage        bool                `json:"isStage"`
	Name           string              `json:"name"`
	Variables      map[string]Variable `json:"variables"`
	Lists          map[string]List     `json:"lists"`
	Broadcasts     map[string]string   `json:"broadcasts,omitempty"`
	Blocks         map[string]*Block   `json:"blocks"`
	Comments       map[string]Comment  `json:"comments"`
	CurrentCostume uint                `json:"currentCostume"`
	Costumes       []Costume           `json:"costumes"`
	Sounds         []Sound             `json:"sounds"`
	LayerOrder     float64             `json:"layerOrder"`
	Volume         float64             `json:"volume"`
	// for the stage only
	Tempo                *float64 `json:"tempo,omitempty"`
	VideoState           *string  `json:"videoState,omitempty"`
	VideoTransparency    *float64 `json:"videoTransparency,omitempty"`
	TextToSpeechLanguage *string  `json:"textToSpeechLanguage,omitempty"`
	// for sprites only
	Visible       *bool    `json:"visible,omitempty"`
	X             *float64 `json:"x,omitempty"`
	Y             *float64 `json:"y,omitempty"`
	Size          *float64 `json:"size,omitempty"`
	Direction     *float64 `json:"direction,omitempty"`
	Draggable     *bool    `json:"draggable,omitempty"`
	RotationStyle *string  `json:"rotationStyle,omitempty"`
}

func NewTarget(name string, costumes []Costume) Target {
	visible := true
	var x, y float64 = 0, 0
	var size float64 = 100
	var direction float64 = 90
	dragable := false
	rotationStyle := "all around"
	return Target{
		IsStage:        false,
		Name:           name,
		Variables:      make(map[string]Variable),
		Lists:          make(map[string]List),
		Broadcasts:     make(map[string]string),
		Blocks:         make(map[string]*Block),
		Comments:       make(map[string]Comment),
		CurrentCostume: 0,
		Costumes:       costumes,
		Sounds:         make([]Sound, 0),
		LayerOrder:     0,
		Volume:         100,
		Visible:        &visible,
		X:              &x,
		Y:              &y,
		Size:           &size,
		Direction:      &direction,
		Draggable:      &dragable,
		RotationStyle:  &rotationStyle,
	}
}

type Costume struct {
	AssetId          string  `json:"assetId"`
	Name             string  `json:"name"`
	Md5ext           string  `json:"md5ext"`
	DataFormat       string  `json:"dataFormat"`
	BitmapResolution float64 `json:"bitmapResolution,omitempty"`
	RotationCenterX  float64 `json:"rotationCenterX,omitempty"`
	RotationCenterY  float64 `json:"rotationCenterY,omitempty"`
}

type Sound struct {
	AssetId     string  `json:"assetId"`
	Name        string  `json:"name"`
	Md5ext      string  `json:"md5ext"`
	DataFormat  string  `json:"dataFormat"`
	Rate        float64 `json:"rate,omitempty"`
	SampleCount float64 `json:"sampleCount,omitempty"`
}

type Block struct {
	Opcode   string                        `json:"opcode"`
	Fields   map[string]Field              `json:"fields"`
	Inputs   map[string]MaybeShadowedInput `json:"inputs"`
	Parent   *string                       `json:"parent"`
	Next     *string                       `json:"next"`
	Shadow   bool                          `json:"shadow"`
	TopLevel bool                          `json:"topLevel"`
	X        *float64                      `json:"x,omitempty"`
	Y        *float64                      `json:"y,omitempty"`
	Comment  *string                       `json:"comment,omitempty"`
	Mutation *Mutation                     `json:"mutation,omitempty"`
}

type Input interface {
	InputType() InputType
}

type InputType uint8

const (
	InputBlock           InputType = 0
	InputNumber          InputType = 4
	InputPositiveNumber  InputType = 5
	InputPositiveInteger InputType = 6
	InputInteger         InputType = 7
	InputAngle           InputType = 8
	InputColor           InputType = 9
	InputString          InputType = 10
	InputBroadcast       InputType = 11
	InputVariable        InputType = 12
	InputList            InputType = 13
)

type InputShadowType uint8

const (
	Shadow    InputShadowType = 1
	Nonshadow InputShadowType = 2
	Shadowed  InputShadowType = 3
)

type MaybeShadowedInput struct {
	Type          InputShadowType
	ObscuredInput Input
	ShadowedInput Input
}

func (s *MaybeShadowedInput) UnmarshalJSON(data []byte) error {
	var array []json.RawMessage
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	var inputShadowType uint
	if err := json.Unmarshal(array[0], &inputShadowType); err != nil {
		return err
	}
	switch inputShadowType {
	case 1:
		s.Type = Shadow
		shadowedInput, err := unmarshalInput(array[1])
		if err != nil {
			return err
		}
		s.ShadowedInput = shadowedInput
	case 2:
		s.Type = Nonshadow
		obscuredInput, err := unmarshalInput(array[1])
		if err != nil {
			return err
		}
		s.ObscuredInput = obscuredInput
	case 3:
		s.Type = Shadowed
		obscuredInput, err := unmarshalInput(array[1])
		if err != nil {
			return err
		}
		s.ObscuredInput = obscuredInput
		shadowedInput, err := unmarshalInput(array[2])
		if err != nil {
			return err
		}
		s.ShadowedInput = shadowedInput
	default:
		return &json.UnmarshalTypeError{}
	}
	return nil
}

func (s MaybeShadowedInput) MarshalJSON() ([]byte, error) {
	bytes := make([]byte, 0)
	bytes = append(bytes, '[')
	switch s.Type {
	case Shadow:
		bytes = append(bytes, "1,"...)
		shadowedInput, err := json.Marshal(s.ShadowedInput)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, shadowedInput...)
	case Nonshadow:
		bytes = append(bytes, "2,"...)
		obscuredInput, err := json.Marshal(s.ObscuredInput)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, obscuredInput...)
	case Shadowed:
		bytes = append(bytes, "3,"...)
		obscuredInput, err := json.Marshal(s.ObscuredInput)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, obscuredInput...)
		bytes = append(bytes, ',')
		shadowedInput, err := json.Marshal(s.ShadowedInput)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, shadowedInput...)
	}
	bytes = append(bytes, ']')
	return bytes, nil
}

func unmarshalInput(data json.RawMessage) (Input, error) {
	var blockInput string
	if err := json.Unmarshal(data, &blockInput); err == nil {
		return (*BlockInput)(&blockInput), nil
	}
	var array []any
	if err := json.Unmarshal(data, &array); err != nil {
		return nil, err
	}
	inputType, _ := array[0].(float64)
	switch uint(inputType) {
	case 4:
		fallthrough
	case 5:
		fallthrough
	case 6:
		fallthrough
	case 7:
		fallthrough
	case 8:
		value, _ := array[1].(float64)
		return &NumberalInput{
			Type:  InputType(inputType),
			Value: value,
		}, nil
	case 9:
		fallthrough
	case 10:
		value, _ := array[1].(string)
		return &StringInput{
			Type:  InputType(inputType),
			Value: value,
		}, nil
	case 11:
		value, _ := array[1].(string)
		id, _ := array[2].(string)
		return &BroadcastInput{
			Value: value,
			Id:    id,
		}, nil
	case 12:
		fallthrough
	case 13:
		value, _ := array[1].(string)
		id, _ := array[2].(string)
		var x, y *float64
		if len(array) == 5 {
			theX, _ := array[3].(float64)
			theY, _ := array[4].(float64)
			x = &theX
			y = &theY
		}
		return &VariableOrListInput{
			Value: value,
			Id:    id,
			X:     x,
			Y:     y,
		}, nil
	}
	return nil, &json.UnmarshalTypeError{}
}

type NumberalInput struct {
	Type  InputType
	Value float64
}

func (s *NumberalInput) InputType() InputType {
	return s.Type
}

func (s NumberalInput) MarshalJSON() ([]byte, error) {
	bytes := make([]byte, 0)
	bytes = append(bytes, '[')
	inputType, err := json.Marshal(s.Type)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, inputType...)
	bytes = append(bytes, ',')
	value, err := json.Marshal(s.Value)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, value...)
	bytes = append(bytes, ']')
	return bytes, nil
}

type StringInput struct {
	Type  InputType
	Value string
}

func (s *StringInput) InputType() InputType {
	return s.Type
}

func (s StringInput) MarshalJSON() ([]byte, error) {
	bytes := make([]byte, 0)
	bytes = append(bytes, '[')
	inputType, err := json.Marshal(s.Type)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, inputType...)
	bytes = append(bytes, ',')
	value, err := json.Marshal(s.Value)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, value...)
	bytes = append(bytes, ']')
	return bytes, nil
}

type BroadcastInput struct {
	Value string
	Id    string
}

func (s *BroadcastInput) InputType() InputType {
	return InputBroadcast
}

func (s BroadcastInput) MarshalJSON() ([]byte, error) {
	bytes := make([]byte, 0)
	bytes = append(bytes, '[')
	bytes = append(bytes, []byte(strconv.Itoa(int(InputBroadcast)))...)
	bytes = append(bytes, ',')
	value, err := json.Marshal(s.Value)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, value...)
	bytes = append(bytes, ',')
	id, err := json.Marshal(s.Id)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, id...)
	bytes = append(bytes, ']')
	return bytes, nil
}

type VariableOrListInput struct {
	Type  InputType
	Value string
	Id    string
	X     *float64
	Y     *float64
}

func (s *VariableOrListInput) InputType() InputType {
	return s.Type
}

func (s VariableOrListInput) MarshalJSON() ([]byte, error) {
	bytes := make([]byte, 0)
	bytes = append(bytes, '[')
	inputType, err := json.Marshal(s.Type)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, inputType...)
	bytes = append(bytes, ',')
	value, err := json.Marshal(s.Value)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, value...)
	bytes = append(bytes, ',')
	id, err := json.Marshal(s.Id)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, id...)
	if s.X != nil && s.Y != nil {
		bytes = append(bytes, ',')
		x, err := json.Marshal(s.X)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, x...)
		bytes = append(bytes, ',')
		y, err := json.Marshal(s.Y)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, y...)
	}
	bytes = append(bytes, ']')
	return bytes, nil
}

type BlockInput string

func (s *BlockInput) InputType() InputType {
	return InputBlock
}

type Field struct {
	Value string
	Id    *string
}

func (s *Field) UnmarshalJSON(data []byte) error {
	var array []any
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	if len(array) < 1 {
		return &json.UnmarshalTypeError{}
	}
	s.Value, _ = array[0].(string)
	if len(array) > 1 {
		id, _ := array[1].(string)
		s.Id = &id
	}
	return nil
}

func (s Field) MarshalJSON() ([]byte, error) {
	bytes := make([]byte, 0)
	bytes = append(bytes, '[')
	inputType, err := json.Marshal(s.Value)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, inputType...)
	if s.Id != nil {
		bytes = append(bytes, ',')
		id, err := json.Marshal(s.Id)
		if err != nil {
			return nil, err
		}
		bytes = append(bytes, id...)
	}
	bytes = append(bytes, ']')
	return bytes, nil
}

type Mutation struct {
	TagName  string `json:"tagName"`
	Children []any  `json:"children"`
	// for procedures_prototype and procedures_call only
	ProcCode    *string `json:"proccode,omitempty"`
	ArgumentIds *string `json:"argumentids,omitempty"`
	Warp        *string `json:"warp,omitempty"`
	// for procedures_prototype only
	ArgumentNames    *string `json:"argumentnames,omitempty"`
	ArgumentDefaults *string `json:"argumentdefaults,omitempty"`
	// for control_stop only
	HasNext *string `json:"hasnext,omitempty"`
}

type Comment struct {
	BlockId   string  `json:"blockId"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
	Width     float64 `json:"width"`
	Height    float64 `json:"height"`
	Minimized bool    `json:"minimized"`
	Text      string  `json:"text"`
}

type Variable struct {
	Name    string
	Value   string
	IsCloud bool
}

func (s *Variable) UnmarshalJSON(data []byte) error {
	var array []any
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	if len(array) < 2 {
		return &json.UnmarshalTypeError{}
	}
	if len(array) > 2 {
		s.IsCloud, _ = array[2].(bool)
	}
	s.Name, _ = array[0].(string)
	s.Value, _ = array[1].(string)
	return nil
}

func (s Variable) MarshalJSON() ([]byte, error) {
	bytes := make([]byte, 0)
	bytes = append(bytes, '[')
	name, err := json.Marshal(s.Name)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, name...)
	bytes = append(bytes, ',')
	value, err := json.Marshal(s.Value)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, value...)
	if s.IsCloud {
		bytes = append(bytes, ",true"...)
	}
	bytes = append(bytes, ']')
	return bytes, nil
}

type List struct {
	Name  string
	Value []string
}

func (s *List) UnmarshalJSON(data []byte) error {
	var array []any
	if err := json.Unmarshal(data, &array); err != nil {
		return err
	}
	if len(array) < 2 {
		return &json.UnmarshalTypeError{}
	}
	s.Name, _ = array[0].(string)
	s.Value, _ = array[1].([]string)
	return nil
}

func (s List) MarshalJSON() ([]byte, error) {
	bytes := make([]byte, 0)
	bytes = append(bytes, '[')
	name, err := json.Marshal(s.Name)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, name...)
	bytes = append(bytes, ',')
	value, err := json.Marshal(s.Value)
	if err != nil {
		return nil, err
	}
	bytes = append(bytes, value...)
	bytes = append(bytes, ']')
	return bytes, nil
}

type Monitor struct {
	Id         string   `json:"id"`
	Mode       string   `json:"mode"`
	Opcode     string   `json:"opcode"`
	SpriteName string   `json:"spriteName"`
	Value      string   `json:"value"`
	Width      float64  `json:"width"`
	Height     float64  `json:"height"`
	X          float64  `json:"x"`
	Y          float64  `json:"y"`
	Visible    bool     `json:"visible"`
	SliderMin  *float64 `json:"sliderMin,omitempty"`
	SliderMax  *float64 `json:"sliderMax,omitempty"`
	IsDiscrete *bool    `json:"isDiscrete,omitempty"`
}

type Meta struct {
	Semver string `json:"semver"`
	Vm     string `json:"vm"`
	Agent  string `json:"agent"`
}
