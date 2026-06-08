package loader

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"scoreador/internal/model"
)

func LoadConfig(path string) (model.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return model.Config{}, err
	}
	defer file.Close()

	var cfg model.Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return model.Config{}, err
	}
	cfg.ApplyDefaults()
	return cfg, nil
}

func LoadLambdaRules(path string) ([]model.LambdaRule, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}

	rules := make([]model.LambdaRule, 0, len(rows))
	for _, row := range rows {
		minVal, err := atoiField(row, "tiros_min")
		if err != nil {
			return nil, err
		}
		maxVal, err := atoiField(row, "tiros_max")
		if err != nil {
			return nil, err
		}
		mot, err := model.ParseMotivation(row["motivacion"])
		if err != nil {
			return nil, err
		}
		lambdaVal, err := atofField(row, "lambda")
		if err != nil {
			return nil, err
		}
		rules = append(rules, model.LambdaRule{
			ShotsMin:   minVal,
			ShotsMax:   maxVal,
			Motivation: mot,
			Lambda:     lambdaVal,
		})
	}

	sort.Slice(rules, func(i, j int) bool {
		if rules[i].ShotsMin != rules[j].ShotsMin {
			return rules[i].ShotsMin < rules[j].ShotsMin
		}
		if rules[i].ShotsMax != rules[j].ShotsMax {
			return rules[i].ShotsMax < rules[j].ShotsMax
		}
		return rules[i].Motivation < rules[j].Motivation
	})

	return rules, nil
}

func LoadMatches(path string) ([]model.MatchInput, error) {
	rows, err := readCSV(path)
	if err != nil {
		return nil, err
	}

	matches := make([]model.MatchInput, 0, len(rows))
	for _, row := range rows {
		id, err := atoiField(row, "match_id")
		if err != nil {
			return nil, err
		}
		motA, err := model.ParseMotivation(row["motivacion_a"])
		if err != nil {
			return nil, err
		}
		motB, err := model.ParseMotivation(row["motivacion_b"])
		if err != nil {
			return nil, err
		}
		shotsA, err := atoiField(row, "tiros_a")
		if err != nil {
			return nil, err
		}
		shotsB, err := atoiField(row, "tiros_b")
		if err != nil {
			return nil, err
		}

		matches = append(matches, model.MatchInput{
			MatchID:     id,
			Stage:       strings.ToLower(strings.TrimSpace(row["stage"])),
			Group:       strings.ToUpper(strings.TrimSpace(row["grupo"])),
			TeamA:       strings.TrimSpace(row["equipo_a"]),
			TeamB:       strings.TrimSpace(row["equipo_b"]),
			ShotsA:      shotsA,
			ShotsB:      shotsB,
			MotivationA: motA,
			MotivationB: motB,
		})
	}

	return matches, nil
}

func readCSV(path string) ([]map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	index := make(map[string]int, len(header))
	for i, h := range header {
		index[strings.ToLower(strings.TrimSpace(h))] = i
	}

	rows := make([]map[string]string, 0)
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		row := make(map[string]string, len(index))
		for key, idx := range index {
			if idx < len(record) {
				row[key] = strings.TrimSpace(record[idx])
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

func atoiField(row map[string]string, key string) (int, error) {
	value, ok := row[strings.ToLower(key)]
	if !ok {
		return 0, fmt.Errorf("falta columna %q", key)
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("valor invalido para %s: %w", key, err)
	}
	return n, nil
}

func atofField(row map[string]string, key string) (float64, error) {
	value, ok := row[strings.ToLower(key)]
	if !ok {
		return 0, fmt.Errorf("falta columna %q", key)
	}
	n, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("valor invalido para %s: %w", key, err)
	}
	return n, nil
}
