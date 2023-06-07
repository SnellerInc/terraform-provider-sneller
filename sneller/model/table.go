package model

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

var (
	ErrExpectedStartOfObject = errors.New("expected start-of-object")
	ErrExpectedFieldName     = errors.New("expected field-name")
	ErrInvalidHint           = errors.New("hints should be a single string or an array of strings")
)

type TableInputModel struct {
	Pattern   string                    `tfsdk:"pattern"`
	Format    *string                   `tfsdk:"format"`
	JSONHints []TableInputJSONHintModel `tfsdk:"json_hints"`
	CSVHints  *TableInputCSVHintModel   `tfsdk:"csv_hints"`
	TSVHints  *TableInputTSVHintModel   `tfsdk:"tsv_hints"`
}

func (m *TableInputModel) MarshalJSON() ([]byte, error) {
	sb := bytes.Buffer{}
	sb.Write([]byte(`{"pattern":`))
	patternValue, err := json.Marshal(m.Pattern)
	if err != nil {
		return nil, err
	}
	sb.Write(patternValue)
	if m.Format != nil {
		sb.Write([]byte(`,"format":`))
		formatValue, err := json.Marshal(*m.Format)
		if err != nil {
			return nil, err
		}
		sb.Write(formatValue)
	}
	if len(m.JSONHints) > 0 {
		sb.Write([]byte(`,"hints":`))
		hintsValue, err := json.Marshal(&TableInputJSONHintsModel{m.JSONHints})
		if err != nil {
			return nil, err
		}
		sb.Write(hintsValue)
	}
	if m.CSVHints != nil {
		sb.Write([]byte(`,"hints":`))
		hintsValue, err := json.Marshal(m.CSVHints)
		if err != nil {
			return nil, err
		}
		sb.Write(hintsValue)
	}
	if m.TSVHints != nil {
		sb.Write([]byte(`,"hints":`))
		hintsValue, err := json.Marshal(m.TSVHints)
		if err != nil {
			return nil, err
		}
		sb.Write(hintsValue)
	}
	sb.WriteRune('}')
	return sb.Bytes(), nil
}

func (m *TableInputModel) UnmarshalJSON(data []byte) error {
	var shadow struct {
		Pattern string  `json:"pattern"`
		Format  *string `json:"format,omitempty"`
	}
	err := json.Unmarshal(data, &shadow)
	if err != nil {
		return err
	}

	m.Pattern = shadow.Pattern
	m.Format = shadow.Format

	m.JSONHints = nil
	m.CSVHints = nil
	m.TSVHints = nil

	format := "json"
	if shadow.Format != nil {
		format = strings.TrimPrefix(*shadow.Format, ".")
	}
	switch format {
	case "json", "json.gz", "json.zst":
		var shadow struct {
			Hints *TableInputJSONHintsModel `json:"hints,omitempty"`
		}
		err := json.Unmarshal(data, &shadow)
		if err != nil {
			return err
		}
		if shadow.Hints != nil {
			m.JSONHints = shadow.Hints.Rules
		}

	case "csv", "csv.gz", "csv.zst":
		var shadow struct {
			Hints *TableInputCSVHintModel `json:"hints,omitempty"`
		}
		err := json.Unmarshal(data, &shadow)
		if err != nil {
			return err
		}
		if shadow.Hints != nil {
			m.CSVHints = shadow.Hints
		}

	case "tsv", "tsv.gz", "tsv.zst":
		var shadow struct {
			Hints *TableInputTSVHintModel `json:"hints,omitempty"`
		}
		err := json.Unmarshal(data, &shadow)
		if err != nil {
			return err
		}
		if shadow.Hints != nil {
			m.TSVHints = shadow.Hints
		}
	}

	return nil
}

type TableInputJSONHintsModel struct {
	Rules []TableInputJSONHintModel
}

func (h *TableInputJSONHintsModel) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.Rules)
}

func (h *TableInputJSONHintsModel) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	t, err := d.Token()
	if err != nil {
		return err
	}

	// TODO: Deprecate and remove object encoding
	// TODO: Use high-level unmarshalling functions afterwards

	switch t {
	case json.Delim('{'):
		return h.decodeRulesFromObject(d)
	case json.Delim('['):
		return h.decodeRulesFromArray(d)
	}

	return errors.New("unsupported type; expected 'object' or 'array'")
}

func (h *TableInputJSONHintsModel) decodeRulesFromObject(d *json.Decoder) error {
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}
		if t == json.Delim('}') {
			// End of main json object -> done
			return nil
		}

		path := t.(string)
		hints, err := decodeHints(d)
		if err != nil {
			return err
		}

		h.Rules = append(h.Rules, TableInputJSONHintModel{
			Path:  path,
			Hints: hints,
		})
	}
}

func (h *TableInputJSONHintsModel) decodeRulesFromArray(d *json.Decoder) error {
	for {
		t, err := d.Token()
		if err != nil {
			return err
		}
		if t == json.Delim(']') {
			// End of main json array -> done
			return nil
		}

		if t == json.Delim('{') {
			path, hints, err := decodeRuleObject(d)
			if err != nil {
				return err
			}

			h.Rules = append(h.Rules, TableInputJSONHintModel{
				Path:  path,
				Hints: hints,
			})

			continue
		}

		return errors.New("unsupported type; expected 'object'")
	}
}

func decodeRuleObject(d *json.Decoder) (path string, hints Hints, err error) {
	for {
		t, err := d.Token()
		if err != nil {
			return "", nil, err
		}
		if t == json.Delim('}') {
			// End of rule json object -> done
			break
		}

		label := strings.ToLower(t.(string))
		switch label {
		case "path":
			t, err = d.Token()
			if err != nil {
				return "", nil, err
			}
			value, ok := t.(string)
			if !ok {
				return "", nil, errors.New("unsupported type; expected 'string'")
			}
			path = value
		case "hints":
			value, err := decodeHints(d)
			if err != nil {
				return "", nil, err
			}
			hints = value
		default:
			// Ignore all extra fields..
			if err = skipValue(d); err != nil {
				return "", nil, err
			}
		}
	}
	return
}

func decodeHints(d *json.Decoder) (Hints, error) {
	t, err := d.Token()
	if err != nil {
		return nil, err
	}

	value, ok := t.(string)
	if ok {
		return []string{value}, nil
	}

	if t != json.Delim('[') {
		return nil, errors.New("unsupported type; expected 'string' or '[]string'")
	}

	var result Hints
	for {
		t, err := d.Token()
		if err != nil {
			return nil, err
		}
		if t == json.Delim(']') {
			return result, nil
		}
		value, ok := t.(string)
		if !ok {
			return nil, errors.New("unsupported type; expected 'string'")
		}

		result = append(result, value)
	}
}

func skipValue(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	switch t {
	case json.Delim('['), json.Delim('{'):
		for {
			if err := skipValue(d); err != nil {
				if err == errErrDelim {
					break
				}
				return err
			}
		}
	case json.Delim(']'), json.Delim('}'):
		return errErrDelim
	}
	return nil
}

var errErrDelim = errors.New("invalid end of array or object")

type Hints []string

func (h *Hints) MarshallJSON() ([]byte, error) {
	if len(*h) == 0 {
		return json.Marshal("default")
	}
	if len(*h) == 1 {
		return json.Marshal((*h)[0])
	}

	sort.Strings(*h)

	return json.Marshal(*h)
}

func (h *Hints) UnmarshallJSON(data []byte) (err error) {
	*h = (*h)[0:0]

	switch data[0] {
	case '"':
		var s string
		if err = json.Unmarshal(data, &s); err != nil {
			return err
		}
		*h = append(*h, s)
	case '[':
		if err = json.Unmarshal(data, h); err != nil {
			return err
		}
	default:
		return errors.New("unsupported type; expected string or list of strings")
	}

	return err
}

type TableInputJSONHintModel struct {
	Path  string `tfsdk:"path"  json:"path"`
	Hints Hints  `tfsdk:"hints" json:"hints"`
}

type TableInputCSVHintModel struct {
	Separator     *string                        `tfsdk:"separator" json:"separator,omitempty"`
	SkipRecords   *int64                         `tfsdk:"skip_records" json:"skip_records,omitempty"`
	MissingValues []string                       `tfsdk:"missing_values" json:"missing_values,omitempty"`
	Fields        []TableInputXSVHintsFieldModel `tfsdk:"fields" json:"fields,omitempty"`
}

type TableInputTSVHintModel struct {
	SkipRecords   *int64                         `tfsdk:"skip_records" json:"skip_records,omitempty"`
	MissingValues []string                       `tfsdk:"missing_values" json:"missing_values,omitempty"`
	Fields        []TableInputXSVHintsFieldModel `tfsdk:"fields" json:"fields,omitempty"`
}

type TableInputXSVHintsFieldModel struct {
	Name          *string  `tfsdk:"name" json:"name,omitempty"`
	Type          *string  `tfsdk:"type" json:"type,omitempty"`
	Default       *string  `tfsdk:"default" json:"default,omitempty"`
	Format        *string  `tfsdk:"format" json:"format,omitempty"`
	AllowEmpty    *bool    `tfsdk:"allow_empty" json:"allowEmpty,omitempty"`
	NoIndex       *bool    `tfsdk:"no_index" json:"noIndex,omitempty"`
	TrueValues    []string `tfsdk:"true_values" json:"trueValues,omitempty"`
	FalseValues   []string `tfsdk:"false_values" json:"falseValues,omitempty"`
	MissingValues []string `tfsdk:"missing_values" json:"missingValues,omitempty"`
}

type TablePartitionModel struct {
	Field string  `tfsdk:"field" json:"field"`
	Type  *string `tfsdk:"type" json:"type,omitempty"`
	Value *string `tfsdk:"value" json:"value,omitempty"`
}

type TableRetentionModel struct {
	Field    string `tfsdk:"field" json:"field"`
	ValidFor string `tfsdk:"valid_for" json:"valid_for"`
}
