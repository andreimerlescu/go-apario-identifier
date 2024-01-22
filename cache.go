package go_apario_identifier

import (
	`context`
	`errors`
	`fmt`
	`io/fs`
	`log`
	`os`
	`path/filepath`
	`strconv`
	`strings`
	`sync`
	`sync/atomic`
	`time`

	sema `github.com/andreimerlescu/go-sema`
)

type Cache struct {
	ctx        context.Context
	Path       string                    `json:"-"`
	Mutexes    map[string]*sync.RWMutex  `json:"-"`
	Semaphores map[string]sema.Semaphore `json:"-"`
	muMu       *sync.RWMutex
	muSe       *sync.RWMutex
}

func (c *Cache) PathExists(path string) bool {
	return pathExists(path)
}

// LockIdentifier will place a .locked file inside of the directory that belongs to the identifier argument
func (c *Cache) LockIdentifier(identifier string) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		var wasErr bool
		err, wasErr = r.(error)
		if !wasErr {
			err = fmt.Errorf("cache Lock recovered from err %v", r)
		}
		return
	}()
	err = c.IdentifierCheck(identifier)
	if err != nil {
		log.Printf("c.LockIdentifier(%v) received err %v", identifier, err)
		return err
	}

	var breakMe bool = false
	lockerChecker := time.NewTicker(30 * time.Millisecond)
	attempts := atomic.Int64{}
	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-lockerChecker.C: // every 30ms retry
			attempts.Add(1)
			i, readErr := c.readInt64File(identifier, ".locked")
			if readErr != nil {
				// not locked
				breakMe = true
				break
			}
			if i > 0 {
				if attempts.Load() > 30 {
					err = errors.New(".locked file present and failed to unlock within timeout")
					return
				}
			}

		}
		if breakMe {
			break
		}
	}

	c.Semaphore(identifier).Acquire()
	c.Mutex(identifier).Lock()
	return c.writeTimestampFile(identifier, ".locked", time.Now().UTC())
}

// EnsureIdentifierMutex protects against nil dereference errors and locks the mutex if .locked exists in the identifiers' directory
func (c *Cache) EnsureIdentifierMutex(identifier string) (mu *sync.RWMutex) {
	c.SafetyCheck()

	c.muMu.RLock()
	_, mExists := c.Mutexes[identifier]
	c.muMu.RUnlock()
	if !mExists {
		c.muMu.Lock()
		c.Mutexes[identifier] = &sync.RWMutex{}
		c.muMu.Unlock()
	}

	c.muMu.RLock()
	mu = c.Mutexes[identifier]
	c.muMu.RUnlock()

	lockPath := filepath.Join(c.Path, IdentifierPath(identifier), ".locked")
	if c.PathExists(lockPath) {
		lockedAt, readErr := c.readTimestampFile(identifier, ".locked")
		if readErr == nil {
			if time.Now().UTC().After(lockedAt) {
				// locked in the past, aka still actively locked
				if mu.TryLock() {
					mu.Lock()
				}
			}
		}
	}

	return
}

// EnsureIdentifierSemaphore protects against nil dereference errors and loads the identifiers' .sema file into the Cache
func (c *Cache) EnsureIdentifierSemaphore(identifier string) (s sema.Semaphore) {
	c.SafetyCheck()
	c.muSe.RLock()
	_, sExists := c.Semaphores[identifier]
	c.muSe.RUnlock()
	if !sExists {
		semaLimit, semaErr := c.readInt64File(identifier, ".sema")
		if semaErr == nil && semaLimit > 0 {
			c.muSe.Lock()
			c.Semaphores[identifier] = sema.New(int(semaLimit))
			c.muSe.Unlock()
		} else {
			c.muSe.Lock()
			c.Semaphores[identifier] = sema.New(1)
			c.muSe.Unlock()
		}
	}
	c.muSe.RLock()
	s = c.Semaphores[identifier]
	c.muSe.RUnlock()
	return
}

// EnsureIdentifier ensures non-nil assignments to Cache's Mutexes and Semaphores data for identifier
func (c *Cache) EnsureIdentifier(identifier string) {
	c.EnsureIdentifierMutex(identifier)
	c.EnsureIdentifierSemaphore(identifier)
}

func (c *Cache) EnsureIdentifierDirectory(identifier string) (*Identifier, string, error) {
	id, idErr := ParseIdentifier(identifier)
	if idErr != nil {
		return nil, "", idErr
	}
	c.EnsureIdentifier(identifier)

	identifierPath := filepath.Join(c.Path, id.Path())
	if !c.PathExists(identifierPath) {
		mkdirErr := os.MkdirAll(identifierPath, 0700)
		if mkdirErr != nil {
			return nil, "", mkdirErr
		}
	}
	return id, identifierPath, nil
}

func (c *Cache) readInt64File(identifier string, filename string) (int64, error) {
	path := filepath.Join(c.Path, IdentifierPath(identifier), filename)
	if !c.PathExists(path) {
		return 0, errors.New("no such file exists")
	}

	lockedThenBytes, readErr := os.ReadFile(path)
	if readErr != nil {
		return 0, readErr
	}
	thenStr := string(lockedThenBytes)
	return strconv.ParseInt(thenStr, 10, 64)
}

func (c *Cache) writeInt64File(identifier string, filename string, value int64) error {
	_, dir, idErr := c.EnsureIdentifierDirectory(identifier)
	if idErr != nil {
		return idErr
	}
	path := filepath.Join(dir, filename)
	return os.WriteFile(path, []byte(fmt.Sprintf("%d", value)), 0600)
}

func (c *Cache) readTimestampFile(identifier string, filename string) (time.Time, error) {
	then, thenErr := c.readInt64File(identifier, filename)
	if thenErr != nil {
		return time.Now().UTC(), thenErr
	}
	when := time.Unix(int64(then), 0)
	return when, nil
}

func (c *Cache) writeTimestampFile(identifier string, filename string, timestamp time.Time) error {
	return c.writeInt64File(identifier, filename, timestamp.UTC().Unix())
}

func (c *Cache) identifierLockFile(identifier string) string {
	_, path, idErr := c.EnsureIdentifierDirectory(identifier)
	if idErr != nil {
		log.Printf("c.identifierLockFile() raised error at c.EnsureIdentifierDirectory() called idErr = %v", idErr)
	}
	return filepath.Join(path, ".locked")
}

func (c *Cache) removeLockFile(identifier string) bool {
	err := c.IdentifierCheck(identifier)
	if err != nil {
		log.Printf("c.removeLockFile(%v) received err %v", identifier, err)
		return false
	}
	lockPath := c.identifierLockFile(identifier)
	info, infoErr := os.Stat(lockPath)
	if infoErr == nil && !info.IsDir() && info.Name() == ".locked" && info.Size() > 0 {
		rmErr := os.RemoveAll(lockPath)
		if rmErr != nil {
			log.Printf("c.removeLockFile() raised error at os.RemoveAll() called rmErr = %v", rmErr)
			return false
		}
	}
	return true
}

func (c *Cache) UnlockIdentifier(identifier string) {
	err := c.IdentifierCheck(identifier)
	if err != nil {
		log.Printf("c.UnlockIdentifier(%v) received err %v", identifier, err)
		return
	}
	c.removeLockFile(identifier)
	c.Mutex(identifier).Unlock()
	c.Semaphore(identifier).Release()
}

func (c *Cache) IdentifierCheck(identifier string) error {
	c.SafetyCheck()
	c.EnsureIdentifier(identifier)
	return nil

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
		if r == nil {
			err = nil
			return
		}
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

	writeErr := c.writeInt64File(identifier, ".sema", int64(limit))
	if writeErr != nil {
		return writeErr
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

// Semaphore is an alias to EnsureIdentifierSemaphore
func (c *Cache) Semaphore(identifier string) sema.Semaphore {
	return c.EnsureIdentifierSemaphore(identifier)
}

// S is an alias to EnsureIdentifierSemaphore
func (c *Cache) S(identifier string) sema.Semaphore {
	return c.EnsureIdentifierSemaphore(identifier)
}

// Mutex is an alias to EnsureIdentifierMutex
func (c *Cache) Mutex(identifier string) *sync.RWMutex {
	return c.EnsureIdentifierMutex(identifier)
}

// M is an alias to EnsureIdentifierMutex
func (c *Cache) M(identifier string) *sync.RWMutex {
	return c.EnsureIdentifierMutex(identifier)
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
