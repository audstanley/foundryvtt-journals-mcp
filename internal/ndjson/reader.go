package ndjson

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Reader struct {
	path      string
	entities  map[string][]map[string]interface{}
	index     map[string]map[string]int // type→id→index
	whiteList []string
}

const (
	MaxEntitiesPerType = 10000
)

var defaultWhiteList = []string{
	"actors",
	"items",
	"journals",
}

func Open(worldsPath, worldName string) (*Reader, error) {
	dataPath := filepath.Join(worldsPath, worldName, "data")

	reader := &Reader{
		path:      dataPath,
		entities:  make(map[string][]map[string]interface{}),
		index:     make(map[string]map[string]int),
		whiteList: defaultWhiteList,
	}

	if err := reader.loadAll(); err != nil {
		return nil, fmt.Errorf("failed to load ndjson databases: %w", err)
	}

	return reader, nil
}

func (r *Reader) loadAll() error {
	for _, entityType := range r.whiteList {
		dbPath := filepath.Join(r.path, entityType+".db")
		if err := r.loadType(dbPath, entityType); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("failed to load %s: %w", entityType, err)
		}
	}
	return nil
}

func (r *Reader) loadType(dbPath, entityType string) error {
	file, err := os.Open(dbPath)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, exists := r.entities[entityType]; exists {
		r.entities[entityType] = make([]map[string]interface{}, 0)
	} else {
		r.entities[entityType] = []map[string]interface{}{}
	}

	r.index[entityType] = make(map[string]int)

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entity map[string]interface{}
		if err := json.Unmarshal([]byte(line), &entity); err != nil {
			continue
		}

		id, ok := entity["_id"].(string)
		if !ok || id == "" {
			continue
		}

		if len(r.entities[entityType]) >= MaxEntitiesPerType {
			break
		}

		r.entities[entityType] = append(r.entities[entityType], entity)
		r.index[entityType][id] = len(r.entities[entityType]) - 1
	}

	return scanner.Err()
}

func (r *Reader) GetByID(entityType, id string) (map[string]interface{}, bool) {
	typeIndex, exists := r.index[entityType]
	if !exists {
		return nil, false
	}

	entityIndex, exists := typeIndex[id]
	if !exists {
		return nil, false
	}

	entity := r.entities[entityType][entityIndex]
	return entity, true
}

func (r *Reader) Search(query string) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	queryLower := strings.ToLower(query)

	for _, entityType := range r.whiteList {
		typeEntities, exists := r.entities[entityType]
		if !exists {
			continue
		}

		for _, entity := range typeEntities {
			if r.entityMatches(entity, queryLower) {
				entityCopy := make(map[string]interface{})
				for k, v := range entity {
					entityCopy[k] = v
				}
				entityCopy["_source"] = entityType
				results = append(results, entityCopy)
			}
		}
	}

	return results, nil
}

func (r *Reader) entityMatches(entity map[string]interface{}, query string) bool {
	name, ok := entity["name"].(string)
	if ok && strings.Contains(strings.ToLower(name), query) {
		return true
	}

	entityType, ok := entity["type"].(string)
	if ok && strings.Contains(strings.ToLower(entityType), query) {
		return true
	}

	system, ok := entity["system"].(map[string]interface{})
	if ok {
		if r.searchSystem(system, query) {
			return true
		}
	}

	return false
}

func (r *Reader) searchSystem(system map[string]interface{}, query string) bool {
	for key, value := range system {
		if strings.Contains(strings.ToLower(key), query) {
			return true
		}

		switch v := value.(type) {
		case string:
			if strings.Contains(strings.ToLower(v), query) {
				return true
			}
		case map[string]interface{}:
			if r.searchSystem(v, query) {
				return true
			}
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(map[string]interface{}); ok {
					if r.searchSystem(s, query) {
						return true
					}
				} else if str, ok := item.(string); ok {
					if strings.Contains(strings.ToLower(str), query) {
						return true
					}
				}
			}
		}
	}

	return false
}

func (r *Reader) Close() error {
	r.entities = nil
	r.index = nil
	return nil
}

func (r *Reader) EntityCount(entityType string) int {
	if entities, exists := r.entities[entityType]; exists {
		return len(entities)
	}
	return 0
}

func (r *Reader) GetEntityTypes() []string {
	types := make([]string, 0, len(r.whiteList))
	for _, t := range r.whiteList {
		if count := r.EntityCount(t); count > 0 {
			types = append(types, t)
		}
	}
	return types
}
