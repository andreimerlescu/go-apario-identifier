package go_apario_identifier

import (
	`errors`
	`fmt`
	`io/fs`
	`log`
	`os`
	`path/filepath`
	`strings`
	`sync`

	sema `github.com/andreimerlescu/go-sema`
)

type Cache struct {
	Mutexes    map[string]*sync.RWMutex  `json:"-"`
	Semaphores map[string]sema.Semaphore `json:"-"`
	muMu       *sync.RWMutex
	muSe       *sync.RWMutex
}

func (c *Cache) LockPath(identifier string) (err error) {
	defer func() {
		r := recover()
		var wasErr bool
		err, wasErr = r.(error)
		if !wasErr {
			err = fmt.Errorf("cache Lock recovered from err %v", r)
		}
		return
	}()
	c.SafetyCheck()
	c.S(identifier).Acquire()
	c.M(identifier).Lock()
	return nil
}

func (c *Cache) UnlockPath(identifier string) {
	c.SafetyCheck()
	c.M(identifier).Unlock()
	c.S(identifier).Release()
}

func (c *Cache) SafetyCheck() {
	if c.muMu == nil {
		c.muMu = &sync.RWMutex{}
	}
	if c.muSe == nil {
		c.muSe = &sync.RWMutex{}
	}
	if c.Mutexes == nil {
		c.muMu.Lock()
		c.Mutexes = make(map[string]*sync.RWMutex)
		c.muMu.Unlock()
	}
	if c.Semaphores == nil {
		c.muSe.Lock()
		c.Semaphores = make(map[string]sema.Semaphore)
		c.muSe.Unlock()
	}
}

func (c *Cache) Write(identifier string, limit int) (err error) {
	go func() {
		r := recover()
		var wasErr bool
		err, wasErr = r.(error)
		if !wasErr {
			err = fmt.Errorf("recovered from %v", r)
		}
	}()
	c.SafetyCheck()

	id, idErr := ParseIdentifier(identifier)
	if idErr != nil {
		err = idErr
		return
	}

	if id.String() != identifier {
		err = errors.New("mismatching identifiers on Cache Write")
		return
	}

	c.muMu.RLock()
	_, mutexExists := c.Mutexes[identifier]
	c.muMu.RUnlock()
	if !mutexExists {
		c.muMu.Lock()
		c.Mutexes[identifier] = &sync.RWMutex{}
		c.muMu.Unlock()
	}

	c.muSe.RLock()
	_, semaExists := c.Semaphores[identifier]
	c.muSe.RUnlock()
	if !semaExists {
		if limit == 0 {
			limit = 1
		}
		c.muSe.Lock()
		c.Semaphores[identifier] = sema.New(limit)
		c.muSe.Unlock()
	}
	return
}

func (c *Cache) S(identifier string) sema.Semaphore {
	c.SafetyCheck()
	return c.Semaphores[identifier]
}

func (c *Cache) M(identifier string) *sync.RWMutex {
	c.SafetyCheck()
	return c.Mutexes[identifier]
}

func (c *Cache) LoadDatabase(databasePath string) error {
	c.SafetyCheck()
	return filepath.Walk(databasePath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil // skip over files
		}
		path = strings.ReplaceAll(path, databasePath, ``)
		if strings.Contains(path, `.`) {
			return nil // skip over dot directories like .git or ..
		}
		maybeIdentifier := strings.ReplaceAll(path, string(os.PathSeparator), ``)
		if len(maybeIdentifier) < 6 || len(maybeIdentifier) > 29 {
			return nil // invalid range for an identifier; 6 is recommended, no more than 29 though!
		}
		identifier, idErr := ParseIdentifier(maybeIdentifier)
		if idErr != nil {
			log.Printf("LoadDatabase() failed ParseIdentifier(%v) resulted in %v", maybeIdentifier, idErr)
			return nil // skip over invalid identifiers
		}
		c.muMu.Lock()
		c.Mutexes[identifier.String()] = &sync.RWMutex{}
		c.muMu.Unlock()

		c.muSe.Lock()
		c.Semaphores[identifier.String()] = sema.New(1)
		c.muSe.Unlock()

		return nil
	})
}
