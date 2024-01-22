package go_apario_identifier

import (
	`encoding/json`
	`os`
	`path/filepath`
	`testing`
)

func TestNewValet(t *testing.T) {
	db, err := os.MkdirTemp("", "users2.db")
	if err != nil {
		t.Errorf("os.MkdirTemp() received err %v", err)
		return
	}
	valet := NewValet(db)
	valet.SafetyCheck()

	err = valet.NewCountableDatabase(db)
	if err != nil {
		t.Errorf("valet.NewCountableDatabase received err %v", err)
		return
	}
	id, idErr := valet.NextID(db)
	if idErr != nil {
		t.Errorf("valet.NextID(db) returned err %v ", idErr)
		return
	}

	type user struct {
		id       *Identifier
		email    string
		username string
	}

	newUser := user{
		id:       id,
		email:    "user@example.com",
		username: "user",
	}

	userFile := filepath.Join(db, id.Path(), "user.json")
	userBytes, bytesErr := json.Marshal(newUser)
	if bytesErr != nil {
		t.Errorf("json.Marshal returned err %v", bytesErr)
		return
	}
	cache, cacheErr := valet.GetCache(db)
	if cacheErr != nil {
		t.Errorf("valet.GetCache() returned err %v", cacheErr)
		return
	}

	err = cache.LockIdentifier(id.String())
	if err != nil {
		t.Errorf("cache.LockIdentifier(%v) returned err %v", id.String(), err)
		return
	}
	writeErr := os.WriteFile(userFile, userBytes, 0600)
	if writeErr != nil {
		cache.UnlockIdentifier(id.String())
		t.Errorf("os.WriteFile received err %v", writeErr)
		return
	}
	cache.UnlockIdentifier(id.String())
}
