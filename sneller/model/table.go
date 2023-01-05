package model

import (
	"bytes"
	"encoding/json"
	"errors"
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
	CSVHints  []TableInputCSVHintModel  `tfsdk:"csv_hints"`
	TSVHints  []TableInputTSVHintModel  `tfsdk:"tsv_hints"`
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
	if len(m.CSVHints) > 0 {
		sb.Write([]byte(`,"hints":`))
		hintsValue, err := json.Marshal(m.CSVHints)
		if err != nil {
			return nil, err
		}
		sb.Write(hintsValue)
	}
	if len(m.TSVHints) > 0 {
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
			m.JSONHints = shadow.Hints.Hints
		}

	case "csv", "csv.gz", "csv.zst":
		var shadow struct {
			Hints *[]TableInputCSVHintModel `json:"hints,omitempty"`
		}
		err := json.Unmarshal(data, &shadow)
		if err != nil {
			return err
		}
		if shadow.Hints != nil {
			m.CSVHints = *shadow.Hints
		}

	case "tsv", "tsv.gz", "tsv.zst":
		var shadow struct {
			Hints *[]TableInputTSVHintModel `json:"hints,omitempty"`
		}
		err := json.Unmarshal(data, &shadow)
		if err != nil {
			return err
		}
		if shadow.Hints != nil {
			m.TSVHints = *shadow.Hints
		}
	}

	return nil
}

type TableInputJSONHintsModel struct {
	Hints []TableInputJSONHintModel
}

func (h *TableInputJSONHintsModel) MarshalJSON() ([]byte, error) {
	// We need custom marshalling here, because the order of
	// the hash-map is order-sensitive.
	sb := bytes.Buffer{}
	sb.WriteRune('{')
	totalFields := 0
	for _, fh := range h.Hints {
		if len(fh.Hints) > 0 {
			if totalFields > 0 {
				sb.WriteRune(',')
			}
			fieldName, err := json.Marshal(fh.Field)
			if err != nil {
				return nil, err
			}
			sb.WriteString(string(fieldName))
			sb.WriteRune(':')
			if len(fh.Hints) == 1 {
				err = json.NewEncoder(&sb).Encode(fh.Hints[0])
			} else {
				err = json.NewEncoder(&sb).Encode(fh.Hints)
			}
			if err != nil {
				return nil, err
			}
			totalFields++
		}
	}
	sb.WriteRune('}')
	return sb.Bytes(), nil
}

func (h *TableInputJSONHintsModel) UnmarshalJSON(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	token, err := d.Token()
	if err != nil {
		return err
	}
	if token != json.Delim('{') {
		return ErrExpectedStartOfObject
	}
	h.Hints = nil
	for {
		token, err = d.Token()
		if err != nil {
			return err
		}
		if token == json.Delim('}') {
			break
		}
		field, ok := token.(string)
		if !ok {
			return ErrExpectedFieldName
		}
		var value any
		if err = d.Decode(&value); err != nil {
			return err
		}
		hint := TableInputJSONHintModel{Field: field}
		switch v := value.(type) {
		case string:
			hint.Hints = []string{v}
		case []any:
			hints := make([]string, 0, len(v))
			for _, h := range v {
				if hh, ok := h.(string); ok {
					hints = append(hints, hh)
				} else {
					return ErrInvalidHint
				}
			}
			hint.Hints = hints
		default:
			return ErrInvalidHint
		}
		h.Hints = append(h.Hints, hint)
	}
	return nil
}

type TableInputJSONHintModel struct {
	Field string   `tfsdk:"field"`
	Hints []string `tfsdk:"hints"`
}

type TableInputCSVHintModel []struct {
	Separator     *string                        `tfsdk:"separator" json:"separator,omitempty"`
	SkipRecords   *int64                         `tfsdk:"skip_records" json:"skipRecords,omitempty"`
	MissingValues []string                       `tfsdk:"missing_values" json:"missingValues,omitempty"`
	Fields        []TableInputXSVHintsFieldModel `tfsdk:"fields" json:"fields,omitempty"`
}

type TableInputTSVHintModel []struct {
	SkipRecords   *int64                         `tfsdk:"skip_records" json:"skipRecords,omitempty"`
	MissingValues []string                       `tfsdk:"missing_values" json:"missingValues,omitempty"`
	Fields        []TableInputXSVHintsFieldModel `tfsdk:"fields" json:"fields,omitempty"`
}

type TableInputXSVHintsFieldModel struct {
	Name          *string  `tfsdk:"name" json:"name,omitempty"`
	Type          *string  `tfsdk:"type" json:"type,omitempty"`
	Default       *string  `tfsdk:"default" json:"default,omitempty"`
	Format        *string  `tfsdk:"format" json:"format,omitempty"`
	AllowEmpty    *bool    `tfsdk:"allow_empty" json:"allowEmpty,omitempty"`
	NoIndex       *bool    `tfsdk:"no_index" json:"noIndex,omitempty"`
	TrueValues    []string `tfsdk:"true_values" json:"trueValues"`
	FalseValues   []string `tfsdk:"false_values" json:"falseValues"`
	MissingValues []string `tfsdk:"missing_values" json:"missingValues"`
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
