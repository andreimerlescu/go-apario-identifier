package go_apario_identifier

import (
	`log`
	`os`
	`path/filepath`
	`reflect`
	`testing`
)

func TestCache_LockIdentifier(t *testing.T) {
	db, err := os.MkdirTemp("", "users.db")
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Printf("os.RemoveAll(%v) returned err %v", path, err)
		}
	}(db)
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

	cache, cacheErr := valet.GetCache(db)
	if cacheErr != nil {
		t.Errorf("valet.GetCache() returned err %v", cacheErr)
		return
	}

	id, idErr := valet.NextID(db)
	if idErr != nil {
		t.Errorf("valet.NextID(db) returned err %v ", idErr)
		return
	}

	err = cache.LockIdentifier(id.String())
	if err != nil {
		t.Errorf("cache.LockIdentifier() received err %v", err)
		return
	}

	if !cache.PathExists(filepath.Join(db, id.Path(), ".locked")) {
		t.Errorf("expected .locked file to exist inside path")
		return
	}

	expectErr := cache.LockIdentifier(id.String())
	if expectErr == nil {
		t.Errorf("expected error to trigger for %v", expectErr)
		return
	} else {
		log.Printf("successfully verified that id %v is locked since expectErr = %v", id.String(), expectErr)
	}

	cache.UnlockIdentifier(id.String())

	if cache.PathExists(filepath.Join(db, id.Path(), ".locked")) {
		t.Errorf("expected .locked file to not exist inside path")
		return
	}

	nextId, nextIdErr := valet.NextID(db)
	if nextIdErr != nil {
		t.Errorf("valet.NextID(db) returned unexpected err %v", nextIdErr)
		return
	}

	if reflect.DeepEqual(nextId.Fragment, CodeFragment("2")) {
		t.Errorf("nextID is not equal ; %v == %v", CodeFragment("2"), nextId.Fragment)
	}

	nextId, nextIdErr = valet.NextID(db)
	if nextIdErr != nil {
		t.Errorf("valet.NextID(db) returned unexpected err %v", nextIdErr)
		return
	}

	if reflect.DeepEqual(nextId.Fragment, CodeFragment("3")) {
		t.Errorf("nextID is not equal ; %v == %v", CodeFragment("3"), nextId.Fragment)
	}

	nextId, nextIdErr = valet.NextID(db)
	if nextIdErr != nil {
		t.Errorf("valet.NextID(db) returned unexpected err %v", nextIdErr)
		return
	}

	if reflect.DeepEqual(nextId.Fragment, CodeFragment("4")) {
		t.Errorf("nextID is not equal ; %v == %v", CodeFragment("4"), nextId.Fragment)
	}
}
