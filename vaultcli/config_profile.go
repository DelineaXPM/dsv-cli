package vaultcli

import (
	"strconv"
)

type Profile struct {
	Name string

	data map[string]interface{}
}

func NewProfile(name string) *Profile {
	return &Profile{
		Name: name,
		data: make(map[string]interface{}),
	}
}

func (p *Profile) Get(path ...string) string {
	curr := p.data

	var val string
	var ok bool
	for i, key := range path {
		if i == len(path)-1 {
			val, _ = curr[key].(string)
			break
		}

		curr, ok = curr[key].(map[string]interface{})
		if !ok {
			break
		}
	}
	return val
}

func (p *Profile) Set(val string, path ...string) {
	curr := p.data

	for i, key := range path {
		if i == len(path)-1 {
			if i, err := strconv.Atoi(val); err == nil {
				curr[key] = i
			} else {
				curr[key] = val
			}
			break
		}

		_, ok := curr[key]
		if !ok {
			curr[key] = make(map[string]interface{})
		}

		_, ok = curr[key].(map[interface{}]interface{})
		if ok {
			tmp := make(map[string]interface{})
			for k, v := range curr[key].(map[interface{}]interface{}) {
				tmp[k.(string)] = v
			}
			curr[key] = tmp
		}

		curr = curr[key].(map[string]interface{})
	}
}

func (p *Profile) Del(path ...string) {
	curr := p.data

	var ok bool
	for i, key := range path {
		if i != len(path)-1 {
			_, ok = curr[key].(map[interface{}]interface{})
			if ok {
				tmp := make(map[string]interface{})
				for k, v := range curr[key].(map[interface{}]interface{}) {
					tmp[k.(string)] = v
				}
				curr[key] = tmp
			}
			curr = curr[key].(map[string]interface{})
		}
	}
	delete(curr, path[len(path)-1])
}
