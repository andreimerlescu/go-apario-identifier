# Apario Identifier

This package aims to simplify the use of identifiers in the Project Apario workspace. This package is pretty generic
but was built specifically for the Project Apario database, but technically could be applied to any other type of
application that needs this specific use case.

The database for Project Apario's reader application is a filesystem directory based database that leverages the
directory tree structure of the operating system to use a breadth vs depth approach to storing unique identifiers
as directories within the tree of the filesystem.

An `Identifier` type in this package is a struct that contains a Year `int` and a Fragment `[]rune`. The fragment is
presented and stored in Base36, which is effectively a slice of possible char values ranging from A-Z or 0-9 (total of
36 elements in that, therefore base36 identifier fragment with the year).

The year is used to quarantine on the disk writes based on a given year. For example, all of 2024's data would be
inside a directory called 2024. Likewise, 2025 would be inside 2025. Once 2024 is written, new entries will not be
created by the application to store new data, however updates may still be written by the application to previous
year directories. If a previous year directory is a read-only filesystem, then proposals and updates are disabled.

```go
type Fragment []rune
type Identifier struct {
  Year int `json:"y"`
  Fragment Fragment `json:"f"`
}
```

There are methods associated to the `Identifier` type that are:

```go
func (i *Identifier) Identifier() *Identifier
func (i *Identifier) ID() string
func (i *Identifier) Path() string
func (i *Identifier) String() string
```

The methods `.ID()` and `.String()` are identical. The `.String()` method implements the strings.Stringer interface and
the `.Path()` will convert the `.String()` value into a fibonacci split up path based on the length of the identifier's
Fragment section. Ultimately, can the length of the Fragment fit inside n-fibonacci sums?

For example:

```go
fragment := CodeFragment("abcdef")
id, idErr := identifier := fragment.ToIdentifier()
if idErr == nil {
  log.Println(id.Path()) // prints 2024/a/b/cd/ef
}
```

Raw identifiers can be used, however its better to use wrapped identifiers or counter identifiers for a given
database directory. Given that this package is designed for an Apario database, its important to understand
how this database was written and why.

Most important concept: indexing the database should only take place against unique identifiers only as the
real way of interacting with the substance of the database should only be through the Textee interface, which
will use this identifier package to convert the 3 gematria values for a substring and then store those values
in their corresponding `[]rune()` representation of each of the gematria display digits. For example, the gematria
value for 1602 would be `[]rune{rune("1"), rune("6"), rune("0"), rune("2")}`.

## Cache

There is a cache that is maintained on a watched database directory that keeps track of each unique identifier
interacted with a given valet that gives quick and easy access to a semaphore and a RW mutex that can be used
by other parts of an application when actions are taking place inside the database itself.

```go
type Cache struct {
	ctx        context.Context
	Path       string                    `json:"-"`
	Mutexes    map[string]*sync.RWMutex  `json:"-"`
	Semaphores map[string]sema.Semaphore `json:"-"`
	muMu       *sync.RWMutex
	muSe       *sync.RWMutex
}
```

Methods attached to the `Cache` type:

```go
func (c *Cache) PathExists(path string) bool
func (c *Cache) LockIdentifier(identifier string) (err error)
func (c *Cache) EnsureIdentifierMutex(identifier string) (mu *sync.RWMutex)
func (c *Cache) EnsureIdentifierSemaphore(identifier string) (s sema.Semaphore)
func (c *Cache) EnsureIdentifier(identifier string)
func (c *Cache) EnsureIdentifierDirectory(identifier string) (*Identifier, string, error)
func (c *Cache) UnlockIdentifier(identifier string)
func (c *Cache) IdentifierCheck(identifier string) error
func (c *Cache) SafetyCheck()
func (c *Cache) Write(identifier string, limit int) (err error)
func (c *Cache) Semaphore(identifier string) sema.Semaphore
func (c *Cache) S(identifier string) sema.Semaphore
func (c *Cache) Mutex(identifier string) *sync.RWMutex
func (c *Cache) M(identifier string) *sync.RWMutex
func (c *Cache) LoadDatabase(databasePath string) error
```

Non-exporter functions are:

```go
func (c *Cache) readInt64File(identifier string, filename string) (int64, error)
func (c *Cache) writeInt64File(identifier string, filename string, value int64) error
func (c *Cache) readTimestampFile(identifier string, filename string) (time.Time, error)
func (c *Cache) writeTimestampFile(identifier string, filename string, timestamp time.Time) error
func (c *Cache) identifierLockFile(identifier string) string
func (c *Cache) removeLockFile(identifier string) bool
```

## Valet

The Valet is an interface that makes working with Identifiers and Caches easy to use. A Valet is also able
to keep track of multiple "databases", thus making each directory that is provided as a "database" in essence
a "table" but its important not to use those terms when discussing WHAT the entry is.

```go
type Valet struct {
	ctx       context.Context
	Databases map[string]*Cache `json:"-"`
	mu        *sync.RWMutex
	lim       int
}
```

Method attached to the Valet structure:

```go
func NewValetWithContext(ctx context.Context, databasePath string) *Valet
func NewValet(databasePath string) *Valet
func (v *Valet) GetCache(databasePrefix string) (*Cache, error)
func (v *Valet) SetCache(databasePrefix string, identifier string) (*Cache, error)
func (v *Valet) Lock(databasePrefix string, identifier string) (err error)
func (v *Valet) Unlock(databasePrefix string, identifier string)
func (v *Valet) Acquire(databasePrefix string, identifier string)
func (v *Valet) Release(databasePrefix string, identifier string)
func (v *Valet) SafetyCheck()
func (v *Valet) NewCountableDatabase(databasePath string) error
func (v *Valet) IsCountableDatabase(databasePath string) bool
func (v *Valet) LastID(databasePath string) (*Identifier, error)
func (v *Valet) NextID(databasePath string) (*Identifier, error)
func (v *Valet) NewID(databasePath string, length int) (*Identifier, error)
func (v *Valet) Scan() error
func (v *Valet) PathExists(path string) bool
```

## Testing

This package has nearly 100% code coverage associated with the functions offered throughout this package and the best
example of seeing the application is actually within the tests. To test the application:


```sh
mkdir -p ~/workspace
cd ~/workspace
git clone github.com/andreimerlescu/go-apario-identifier
cd go-apario-identifier
go test ./...
```

## Valet Example

The test has the same code, but for documentation purposes you can see what the Valet does with comments:

This test does the following:

1. Creates a temp directory called `users.db`
2. Creates a new `Valet` to track `users.db`
3. Ensures that Semaphores and Mutexes within Valet's Cache is defined
4. Establishes a new `.lastid` inside directory `users.db` with the value of `int64(1)` inside the file
5. Reference the Valet's Cache directly
6. Ask for a new `*Identifier` by running `.NextID()` and since its a countable database with a `.lastid`, it stores (value of .lastid)+=1 into .lastid and returns the identifier
7. Lock the identifier's directory (so something can happen inside the identifiers' directory, such as writing the identifier's json data)
8. Check that the `.locked` file exists in the identifiers' directory.
9. Try to lock a locked identifiers - this will return an error as expected
10. Unlock the identifier's directory.
11. Verify that .locked does not exist in an unlocked identifier directory
12. Get the next ID in the set reading from `.lastid` (in this case == 2) where nextID = now 2024[base36(2)]
13. Get the next ID in the set
14. Get the next ID in the set

Inside each directory where an identifier actually exists, a `.identifier` file will be created with the value of the
identifier inside it. This can be used for the purposes of taking the value of the `.identifier` file and checking it
against the current path of the current directory where `.identifier` exists. If the `.Path()` value of that
`.identifier` value through `ParseIdentifier()` does not match, then it means the directory is being used for another
identifier and THAT directory belongs to `.identifier` for its data.

```go
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
// handle err

cache, cacheErr := valet.GetCache(db)
// handle err

id, idErr := valet.NextID(db) // since valet.NewCountableDatabase; id == YYYY[base36(1)]
// handle err
// id == *Identifier

err = cache.LockIdentifier(id.String()) // places a .locked file inside db/id.Path()
// handle err

fileExists := cache.PathExists(filepath.Join(db, id.Path(), ".locked")) // does file exist as it should?
// do something with the bool

expectErr := cache.LockIdentifier(id.String()) // try to lock the file while already locked, returns err here
// handle error

cache.UnlockIdentifier(id.String()) // removes the .locked file and updates the RWMutex for the identifier
// no response of any type

fileExists = cache.PathExists(filepath.Join(db, id.Path(), ".locked")) // after Unlock no such file will exist here
// do something with the updated bool

nextId, nextIdErr := valet.NextID(db) // since countable database, .NextID adds 1 to the db/.lastid file ; therefore now = 2
// handle error
// nextId is an increment above id + 1

// reflect.DeepEqual(nextId.Fragment, CodeFragment("2")) // are they exact matches?
// validate the fragments match

nextId, nextIdErr = valet.NextID(db) // since countable database, .NextID adds 1 to the db/.lastid file ; therefore now = 3
// handle error
// nextId is an increment above nextId + 1

// reflect.DeepEqual(nextId.Fragment, CodeFragment("3")) // are they an exact match?
// validate the fragment match

nextId, nextIdErr = valet.NextID(db) // since countable database, .NextID adds 1 to db/.lastid file ; therefore now = 4
// handle error
// nextId is an increment above nextId + 1

// reflect.DeepEqual(nextId.Fragment, CodeFragment("4")) // are they an equal match
// validate the fragment match
```

In addition to using an incrementer database, a non-incremental database can be used that will not permit the use of
`.NextID`. Finally, you can use Valet with a context, so if you need to ensure that for/select statements properly exit
when the main context of your application cancels, then use `NewValetWithContext(ctx, db)`.