package go_apario_identifier

import (
	`context`
	`errors`
	`fmt`
	`log`
	`os`
	`path/filepath`
	`strconv`
	`sync`

	sema `github.com/andreimerlescu/go-sema`
)

type Valet struct {
	ctx       context.Context
	Databases map[string]*Cache `json:"-"`
	mu        *sync.RWMutex
	lim       int
}

func NewValetWithContext(ctx context.Context, databasePath string) *Valet {
	return &Valet{
		ctx: ctx,
		Databases: map[string]*Cache{
			databasePath: {
				ctx:        context.WithoutCancel(ctx),
				Path:       databasePath,
				Mutexes:    make(map[string]*sync.RWMutex),
				Semaphores: make(map[string]sema.Semaphore),
			},
		},
		mu: &sync.RWMutex{},
	}
}

func NewValet(databasePath string) *Valet {
	ctx := context.Background()
	return &Valet{
		ctx: ctx,
		Databases: map[string]*Cache{
			databasePath: {
				ctx:        context.WithoutCancel(ctx),
				Path:       databasePath,
				Mutexes:    make(map[string]*sync.RWMutex),
				Semaphores: make(map[string]sema.Semaphore),
			},
		},
		mu: &sync.RWMutex{},
	}
}

func (v *Valet) GetCache(databasePrefix string) (*Cache, error) {
	v.SafetyCheck()
	return v.Databases[databasePrefix], nil
}

func (v *Valet) SetCache(databasePrefix string, identifier string) (*Cache, error) {
	id, idErr := ParseIdentifier(identifier)
	if idErr != nil {
		return nil, idErr
	}
	cache, cacheErr := v.GetCache(databasePrefix)
	if cacheErr != nil {
		return nil, cacheErr
	}
	err := cache.Write(id.String(), 1)
	if err != nil {
		return nil, err
	}
	return cache, nil
}

func (v *Valet) Lock(databasePrefix string, identifier string) (err error) {
	defer func() {
		r := recover()
		var wasErr bool
		err, wasErr = r.(error)
		if !wasErr {
			err = fmt.Errorf("valet .Lock() received unexpected error %v", r)
		}
	}()
	v.SafetyCheck()
	c := v.Databases[databasePrefix]
	c.SafetyCheck()
	m, exists := c.Mutexes[identifier]
	if !exists {
		return errors.New("no such identifier found")
	}
	m.Lock()
	return nil
}

func (v *Valet) Unlock(databasePrefix string, identifier string) {
	defer func() {
		r := recover()
		log.Printf("valet .Unlock() recovered from panic %v", r)
	}()
	v.SafetyCheck()
	c := v.Databases[databasePrefix]
	c.SafetyCheck()
	m, exists := c.Mutexes[identifier]
	if !exists {
		return
	}
	m.Unlock()
}

func (v *Valet) Acquire(databasePrefix string, identifier string) {
	defer func() {
		r := recover()
		log.Printf("valet .Acquire() recovered from panic %v", r)
	}()
	v.SafetyCheck()
	c := v.Databases[databasePrefix]
	c.SafetyCheck()
	s, exists := c.Semaphores[identifier]
	if !exists {
		return
	}
	s.Acquire()
}

func (v *Valet) Release(databasePrefix string, identifier string) {
	defer func() {
		r := recover()
		log.Printf("valet .Release() recovered from panic %v", r)
	}()
	v.SafetyCheck()
	c := v.Databases[databasePrefix]
	c.SafetyCheck()
	s, exists := c.Semaphores[identifier]
	if !exists {
		return
	}
	s.Release()
}

func (v *Valet) SafetyCheck() {
	if v.mu == nil {
		v.mu = &sync.RWMutex{}
	}
	if v.Databases == nil {
		v.Databases = make(map[string]*Cache)
	}
	if v.lim == 0 {
		v.lim = 3
	}
}

func (v *Valet) NewCountableDatabase(databasePath string) error {
	mkdirErr := os.MkdirAll(databasePath, 0700)
	if mkdirErr != nil {
		return mkdirErr
	}

	firstId := int64(1)
	lastIdPath := filepath.Join(databasePath, ".lastid")
	idStr := fmt.Sprintf("%d", firstId)
	writeErr := os.WriteFile(lastIdPath, []byte(idStr), 0600)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func (v *Valet) IsCountableDatabase(databasePath string) bool {
	_, lastIdErr := v.LastID(databasePath)
	return lastIdErr == nil
}

func (v *Valet) LastID(databasePath string) (*Identifier, error) {
	// assume that database is using incremental base36 for its storage needs
	lastIdPath := filepath.Join(databasePath, ".lastid")
	c, cacheErr := v.GetCache(databasePath)
	if cacheErr != nil {
		// failed to get cache for valet database
		return nil, errors.New("no such cache exists for databasePath")
	}

	if !c.PathExists(lastIdPath) {
		return nil, errors.New("no .lastid found in databasePath")
	}

	lockedThenBytes, readErr := os.ReadFile(lastIdPath)
	if readErr != nil {
		return v.NewID(databasePath, 6)
	}
	thenStr := string(lockedThenBytes)
	lastId, convErr := strconv.Atoi(thenStr)
	if convErr != nil {
		return v.NewID(databasePath, 6)
	}

	identifier, identifierErr := IntegerFragment(lastId).ToIdentifier()
	if identifierErr != nil {
		// invalid identifier generated
		return v.NewID(databasePath, 6)
	}

	return identifier, nil
}

func (v *Valet) NextID(databasePath string) (*Identifier, error) {
	if !v.IsCountableDatabase(databasePath) {
		return v.NewID(databasePath, 6)
	}

	// assume that database is using incremental base36 for its storage needs
	lastIdPath := filepath.Join(databasePath, ".lastid")
	c, cacheErr := v.GetCache(databasePath)
	if cacheErr != nil {
		// failed to get cache for valet database
		return v.NewID(databasePath, 6)
	}

	if !c.PathExists(lastIdPath) {
		return v.NewID(databasePath, 6)
	}

	lockedThenBytes, readErr := os.ReadFile(lastIdPath)
	if readErr != nil {
		return v.NewID(databasePath, 6)
	}
	thenStr := string(lockedThenBytes)
	lastId, convErr := strconv.Atoi(thenStr)
	if convErr != nil {
		return v.NewID(databasePath, 6)
	}

	nextId := lastId + 1
	nextIdStr := fmt.Sprintf("%d", nextId)
	writeErr := os.WriteFile(lastIdPath, []byte(nextIdStr), 0600)
	if writeErr != nil {
		// failed to write to the file
		return v.NewID(databasePath, 6)
	}

	identifier, identifierErr := IntegerFragment(nextId).ToIdentifier()
	if identifierErr != nil {
		// invalid identifier generated
		return v.NewID(databasePath, 6)
	}

	identifierDir := filepath.Join(c.Path, identifier.Path())
	if !c.PathExists(identifierDir) {
		mkdirErr := os.MkdirAll(identifierDir, 0700)
		if mkdirErr != nil {
			return nil, mkdirErr
		}
	}

	identifierPath := filepath.Join(identifierDir, ".identifier")
	idPathLockErr := c.LockIdentifier(identifier.String())
	defer c.UnlockIdentifier(identifier.String())
	if idPathLockErr != nil {
		return v.NewID(databasePath, 6)
	}
	writeErr = os.WriteFile(identifierPath, []byte(identifier.String()), 0600)
	if writeErr != nil {
		return v.NewID(databasePath, 6)
	}

	return identifier, nil
}

func (v *Valet) NewID(databasePath string, length int) (*Identifier, error) {
	c, cErr := v.GetCache(databasePath)
	if cErr != nil {
		return nil, cErr
	}
	id, idErr := NewIdentifier(databasePath, length, 17, 30)
	if idErr != nil {
		return nil, idErr
	}
	s := c.Semaphore(id.String())
	m := c.Mutex(id.String())

	// perform a flush on the semaphore and mutex wrapped with the semaphore first
	s.Acquire()
	m.RLock()
	m.RUnlock()
	s.Release()
	return id, nil
}

func (v *Valet) Scan() error {
	v.mu.Lock() // prevent more than 1 scan at a time from running
	defer v.mu.Unlock()
	v.SafetyCheck()
	wg := &sync.WaitGroup{}
	sem := sema.New(v.lim)
	for name, cache := range v.Databases {
		wg.Add(1)     // worker started
		sem.Acquire() // concurrency protection
		go func(name string, cache *Cache) {
			defer func() { // work completed
				sem.Release()
				wg.Done()
			}()
			err := cache.LoadDatabase(name)
			if err != nil {
				log.Printf("err received cache.LoadDatabase(%v) returned %v", name, err)
			}
		}(name, cache)
	}
	wg.Wait()  // wait for scan to complete
	return nil // no error
}

func (v *Valet) PathExists(path string) bool {
	return pathExists(path)
}
