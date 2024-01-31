package go_apario_identifier

import (
	`errors`
	`net/url`
	`os`
	`sync`
	`sync/atomic`
	`time`

	sema `github.com/andreimerlescu/go-sema`
)

type File struct {
	Name    string  `json:"n"`
	Path    string  `json:"p"`
	Remote  bool    `json:"r"`
	Size    int64   `json:"s"`
	Version Version `json:"v"`
	Kind    string  `json:"k"`
	b       []byte
	locked  *atomic.Bool
	mu      *sync.RWMutex
	o       *sync.Once
	ai      *atomic.Int64
	s       sema.Semaphore
	e       error
	eat     time.Time
}

// FileFromPath returns a *File populated with the Path from the argument and all non-exported properties initialized
func FileFromPath(path string) *File {
	return &File{
		Path: path,
		// now set the musoai ; each file has one musoai
		mu:  &sync.RWMutex{},
		s:   sema.New(1),
		o:   &sync.Once{},
		ai:  &atomic.Int64{},
		e:   nil,
		eat: time.Time{},
	}
}

// Exists runs an os.Stat on the f.Path
func (f *File) Exists() bool {
	if len(f.Path) == 0 {
		return false
	}
	_, statErr := os.Stat(f.Path)
	if statErr == nil {
		return true
	}
	f.e = errors.Join(f.e, statErr)
	f.eat = time.Now().UTC()
	return false
}

// HasBytes determines if len(f.b) > 0
func (f *File) HasBytes() bool {
	return len(f.b) > 0
}

// ReadBytes wraps the os.ReadFile using the f.Path and consuming the error in f.e/f.eat by returning the f when loaded
func (f *File) ReadBytes() *File {
	b, e := os.ReadFile(f.Path)
	if e != nil {
		f.e = errors.Join(f.e, e)
		f.eat = time.Now().UTC()
		return f
	}
	f.b = b
	return f
}

// GetBytes uses downloadBytes(f.Path) to populate the f.b []byte property. In order to use GetBytes you must be able to
// respond to the errChan that it returns.
//
//   err := <-f.GetBytes()
//   if err != nil {
//     log.Errorf(err)
//   }
func (f *File) GetBytes() (errChan chan error) {
	// return channel stuff
	errChan = make(chan error)
	defer close(errChan)

	if f.HasBytes() { // already in memory
		return
	}
	if !f.IsRemote() { // its local
		f.ReadBytes()
		return
	}

	// its remote, lets download it
	var body []byte
	tryCounter := atomic.Int32{}
	fibDepth := atomic.Int32{}
	fibDepth.Store(0)

	for {
		downloadErrCh := make(chan error)
		body, downloadErrCh = downloadBytes(f.Path)
		err := <-downloadErrCh // wait to receive an err here from func
		if err != nil {
			fib := fibonacci(int(fibDepth.Add(1)))
			<-time.Tick(time.Duration(fib*100) * time.Millisecond) // 0ms, 200ms = max 200ms = 2 tries
			if tryCounter.Add(1) > 1 {
				errChan <- err
				return
			}
			continue // retry the downloadBytes()
		}
		break
	}

	if len(body) == 0 {
		errChan <- errors.New("body is 0 bytes; nothing downloaded")
		return
	}
	f.b = body
	body = nil // reset body to save memory
	return
}

// Load will use GetBytes if f.b is empty
func (f *File) Load() *File {
	if len(f.b) == 0 {
		err := <-f.GetBytes()
		if err != nil {
			f.e = errors.Join(f.e, err)
			f.eat = time.Now().UTC()
		}
	}
	return f
}

func (f *File) IsRemote() bool {
	if errors.Is(f.e, ErrNotRemote) {
		return false
	}

	_, parseErr := url.Parse(f.Path)
	if parseErr != nil {
		f.e = errors.Join(ErrNotRemote, parseErr, f.e)
		f.eat = time.Now().UTC()
		return false
	}
	return true
}
