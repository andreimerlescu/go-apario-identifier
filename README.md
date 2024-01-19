# Apario Identifier

This package is responsible for interacting with Apario Identifiers for the Reader/Writer/Database apps.

## Valet

### Functions

This is `Valet`...

```go

func NewValet(databasePath string) *Valet
func (v *Valet) GetCache(databasePrefix string) (*Cache, error)
func (v *Valet) SetCache(databasePrefix string, identifier string) (*Cache, error)
func (v *Valet) Lock(databasePrefix string, identifier string) (err error)
func (v *Valet) Unlock(databasePrefix string, identifier string)
func (v *Valet) Acquire(databasePrefix string, identifier string)
func (v *Valet) Release(databasePrefix string, identifier string)
func (v *Valet) SafetyCheck()
func (v *Valet) NewID(databasePath string, length int) (*Identifier, error)
func (v *Valet) Scan() error
```

The `Valet` structure is very simple.

```go
type Valet struct {
	Databases map[string]*Cache `json:"-"`
	mu        *sync.RWMutex
	lim       int
}
```

This creates a map of paths that are called databasePrefix have many identifiers inside of them. The Cache is the
referenced structure. Inside this code:


```go
type Cache struct {
	Mutexes    map[string]*sync.RWMutex  `json:"-"`
	Semaphores map[string]sema.Semaphore `json:"-"`
	muMu       *sync.RWMutex
	muSe       *sync.RWMutex
}
```

This gives us an interface to use the `string` value in the `map` for `Mutexes` to be the `identifier` of the record
belonging to the valeted' database. The mutexes are used to provide thread safe maps. But the maps connect `identifier`
types of data to a locker mutex and a semaphore that can be customized with the

### Using the Valet

```go
package main

import (
  `fmt`

  ai `github.com/andreimerlescu/go-apario-identifier`
)

func main(){
  db, _ := os.MkdirTemp("", "users.db")
	v := NewValet(db)
	err := v.Scan()
	if err != nil {
		log.Printf("failed to scan valet database directory %v due to err %v", db, err)
		return
	}

	c, cErr := v.GetCache(db)
	if cErr != nil {
		log.Printf("failed to scan valet cache directory %v due to err %v", db, cErr)
		return
	}

	// new id
	id, idErr := v.NewID(db, 6)
	if idErr != nil {
		log.Printf("failed to get new ID for database %v due to err %v", db, idErr)
		return
	}

	// account.json inside id path
	identifierDataDir := filepath.Join(db, id.Path())
	accountFile := filepath.Join(identifierDataDir, "account.json")

	// read the file
	c.S(identifierDataDir).Acquire()                   // allowed to access identifier directory?
	c.M(identifierDataDir).Lock()                      // can access?
	c.S(accountFile).Acquire()                         // allowed to access file?
	c.M(accountFile).RLock()                           // read lock
	accountBytes, bytesErr := os.ReadFile(accountFile) // read json file
	c.M(accountFile).RUnlock()                         // read unlock
	c.S(accountFile).Release()                         // done with accountFile
	c.M(identifierDataDir).Unlock()                    // done with directory
	c.S(identifierDataDir).Release()                   // allow next task
	if bytesErr != nil {
		if errors.Is(bytesErr, os.ErrNotExist) {
			// no such file
		}
		// other issue
	}

	// write to a file
	err = c.LockPath(identifierDataDir)
	if err != nil {
		log.Printf("error %v locking path %v", err, accountFile)
	}
	err = c.LockPath(accountFile)
	if err != nil {
		log.Printf("error %v locking path %v", err, accountFile)
	}
	err = os.WriteFile(accountFile, accountBytes, 0600)
	c.UnlockPath(accountFile)
	c.UnlockPath(identifierDataDir)
	if err != nil {
		return
	}

	// if you didnt want to muck with the directory at all, but just the file itself
	err = c.LockPath(accountFile)
	if err != nil {
		log.Printf("received err %v", err)
	}
	err = os.WriteFile(accountFile, accountBytes, 0600)
	c.UnlockPath(accountFile)
	if err != nil {
		log.Printf("failed to write to file %v due to err %v", accountFile, err)
	}
}

```


## Functions

0. `NewID(databasePath string, length int) (ID, error)`

This func will provide an interface for the Identifier. The Identifier will have a Year + Code associated to it. This
allows identifiers to have a prefix that allows for better time-series archival purposes.

```go
package main

import (
  `fmt`
  `os`

  id `github.com/andreimerlescu/go-apario-identifier`
)

func main() {
  dir, dirErr := os.MkdirTemp("", "db")
  if dirErr != nil {
    panic(dirErr)
  }
  id, err := NewID(dir, 6)
  if err != nil {
    panic(err)
  }
  fmt.Printf(id)
}
```

This will create a new temp directory that contains a fibonacci sequence of subdirectories that correspond to the unique
identifier that was provided. Inside this folder, you can put things that need to be indexed as such. You can create
many interfaces into the data from within this directory using this structure.


1. `IdentifierPath(identifier string) string`

This func will convert a identifier into a path that uses fibonacci to introduce depth from the breadth of the path.
Additionally it uses the fibonacci sequence of the identifier's char positions. A 6 digit long code will have chars
between A-Z0-9 only. An example of that could be `ABC123` as the identifier. This identifier has a special interface{}
attached to it called ID that implements Identification. The `Path()` function will take the identifier and generate a
path `2023/a/b/c1/23` for the `ABC123` identifier of 6 chars long.

Example:

```go
package main

import (
  `fmt`
  `os`

   id `github.com/andreimerlescu/go-apario-identifier`
)

func main() {
  identifier := "2023ABCDEFG"
  path, err := id.IdentifierPath(identifier)
  if err != nil {
    fmt.Printf("failed to calculate identifier %v due to err %v\n", identifier, err)
    os.Exit(1)
  }
  fmt.Printf("done processing\n\nidentifier = %v\npath = %v\n", identifier, path)
  // path = 2023/A/B/CD/EFG
}
```

2. `NewIdentifier(databasePrefixPath string, identifierLength int, attemptsCounter int, timeoutSeconds int) (*Identifier, error)`

This func allows you to bypass the interface called Identification and the NewID response type by getting an instance of
`*Identifier` back instead of an `ID` which is an `Identification` interface implementation. This distinction means the
accessors of private functions can be acquired without first calling `.Instance()` on the `ID` type.

3. `ParseIdentifier(identifier string) (*Identifier, error)`

This func will take a string value of an identifier and return the *Identifier type. If the input string fails to
validate as an identifier, an error is returned.

## Index Uses

This package can manage the `&sync.RWMutex{}` values for each identifier that is ingested into the package. For instance,
you can do this:

```go
package main

import (
  `fmt`
  `os`

  ai `github.com/andreimerlescu/go-apario-identifier`
)

func main() {
  db, err := os.MkdirTemp("", "users.db")
  if err != nil {
    panic(err)
  }

  id, idErr := ai.NewID(db, 6) // 6 chars long for the code
  if idErr != nil {
    panic(idErr)
  }

  err = id.Lock()
  if err != nil {
    panic(err)
  }
  time.Sleep(100*time.Millisecond) // perform a task against this indexed value, such as modifying a file inside the dir
  id.Unlock()

  id.RLock()
  time.Sleep(time.Millisecond)
  id.RUnlock()
}
```

This lock/unlock capability can be utilized throughout the application.

## Semaphores

Each database path provided as the first argument to many functions within this package will each have a 1 length
buffered channel called a semaphore that will be used when Locking and Unlocking an identifier's access. By making
it a semaphore, its a first in first out action on any unique identifier ensuring that the system does not change
during write operations to the INDEX of the database.

```go
package main

import (
  `fmt`
  `os`
  `encoding/json`

  ai `github.com/andreimerlescu/go-apario-identifier`
)

func main() {
  dir1, _ := os.MkdirTemp("", "db1")
  dir2, _ := os.MkdirTemp("", "db2")

  id1, _ := ai.NewID(dir1, 6)

  for {
    select {
      case <-time.Tick(30*time.Second):
        fmt.Printf("ran out of time")
        return
      case <-time.Tick(10*time.Second):
        ai.DatabaseSemaphores[dir1].Acquire() // this will succeed
        id2, _ := ai.NewID(dir1, 6) // this will fail because ai.NewID uses DatabaseSemaphores
        ai.DatabaseSemaphores[dir1].Release()
    }
  }

  usersDb, _ := os.MkdirTemp("", "users.db")
  username := "admin"
  identifier, _ := ai.NewID(usersDb, 6)
  type account struct {
    Username string
    Identifier string
  }
  user := &account{
    Username: username,
    Identifier: identifier.String()
  }
  user_bytes, bytes_err := json.Marshal(user)
  if bytes_err != nil {
    panic(bytes_err)
  }
  path := filepath.Join(usersDb, identifier.Path(), "account.json")
  identifier.Lock()
  writeErr := os.WriteFile(path, user_bytes, 0600)
  identifier.Unlock()
  if writeErr != nil {
    panic(writeErr)
  }

}

```
