package store_test

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"thy/store"
	"time"

	"github.com/stretchr/testify/assert"
)

type tokenData struct {
	Token    []byte `json:"token"`
	Password []byte `json:"password"`
}

var storeTypes []store.StoreType = []store.StoreType{
	store.File,
}

func TestStores(t *testing.T) {
	isWindows := runtime.GOOS == "windows"
	if isWindows {
		storeTypes = append(storeTypes, store.WinCred)
	}

	// TODO : get pass installation working in ci-cd
	// isLinux := runtime.GOOS == "linux"
	// if isLinux {
	// 	storeTypes = append(storeTypes, store.PassLinux)
	// }

	for i, st := range storeTypes {
		t.Run(fmt.Sprintf("case=%d:(%s)", i, st), func(t *testing.T) {
			testStore(t, st)
		})
	}
}

func testStore(t *testing.T, st store.StoreType) {
	// arrange
	rand.Seed(time.Now().UTC().UnixNano())
	sf := store.StoreFactory{}
	s := sf.CreateStore(st)
	token := make([]byte, 4)
	rand.Read(token)
	password := make([]byte, 4)
	rand.Read(password)
	obj := tokenData{
		token,
		password,
	}
	err := s.Store("token", obj)
	assert.Nil(t, err)

	// act
	var obj2 tokenData
	err = s.Get("token", &obj2)
	assert.Nil(t, err)
	assert.Equal(t, obj, obj2)
	_ = s.Delete("token")

	// assert
	obj3 := tokenData{}
	err = s.Get("token", &obj2)
	assert.Empty(t, obj3)

	// arrange
	err = s.Store("token", obj)
	assert.Nil(t, err)

	// act
	_ = s.Wipe("")

	//assert
	obj2 = tokenData{}
	err = s.Get("token", &obj2)
	assert.Empty(t, obj2)
}
