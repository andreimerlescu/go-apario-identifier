package go_apario_identifier

import (
	`errors`
	`fmt`
	`log`
	`sync`

	sema `github.com/andreimerlescu/go-sema`
)

type Valet struct {
	Databases map[string]*Cache `json:"-"`
	mu        *sync.RWMutex
	lim       int
}

func NewValet(databasePath string) *Valet {
	return &Valet{
		Databases: map[string]*Cache{
			databasePath: {
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

func (v *Valet) NewID(databasePath string, length int) (*Identifier, error) {
	c, cErr := v.GetCache(databasePath)
	if cErr != nil {
		return nil, cErr
	}
	id, idErr := NewIdentifier(databasePath, length, 17, 30)
	if idErr != nil {
		return nil, idErr
	}
	s := c.S(id.String())
	m := c.M(id.String())

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
