package go_apario_identifier

import (
	`crypto/tls`
	`encoding/json`
	`errors`
	`fmt`
	`io`
	`io/fs`
	`log`
	`net/http`
	`net/url`
	`os`
	`path/filepath`
	`strconv`
	`strings`
	`sync/atomic`
	`time`
	`unicode`
)

type Identifier struct {
	Instance  []rune   `json:"i"` // the instance of apario-reader serving the identifier
	Concierge []rune   `json:"c"` // the path where valet is served on, the concierge is the valet + cache duo aka combo
	Table     []rune   `json:"t"` // the table of the database needed to access
	Year      int16    `json:"y"` // the year the record was created/added to the database
	Fragment  Fragment `json:"f"` // the base36 random token of n-char length
	Version   *Version `json:"v"` // the version of the record
	files     []*File
	e         Err
	eat       time.Time
}

// pingResponse is the structured data returned by any instance of idoread.com on .Any("/ping") path
type pingResponse struct {
	Message string `json:"message"`
}

// downloadBytes uses insecure TLS to download the bytes from a the path specified and returns the bytes and an error
// channel. When invoking this method, use this convention:
//
//   bytes, err := <-downloadBytes(myPath)
//   if err == nil {
//     fmt.Println(len(bytes))
//   }
//
func downloadBytes(path string) (bytes []byte, errChan chan error) {
	errChan = make(chan error)
	defer close(errChan)
	bytes = []byte{}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	resp, err := client.Get(path)
	if err != nil {
		errChan <- err
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			errChan <- err
			return
		}
	}(resp.Body)

	var ioErr error
	bytes, ioErr = io.ReadAll(resp.Body)
	if ioErr != nil {
		errChan <- ioErr
		return
	}

	return
}

// tryOnline will attempt an HTTPS with TLS verify OFF to validate the concierge endpoint on the instance.
func (i *Identifier) tryOnline() (errChan chan error) {
	errChan = make(chan error)
	defer close(errChan)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}

	resp, err := client.Get(i.Path())
	if err != nil {
		errChan <- err
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			errChan <- err
			return
		}
	}(resp.Body)

	body, ioErr := io.ReadAll(resp.Body)
	if ioErr != nil {
		errChan <- ioErr
		return
	}

	var response pingResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		errChan <- err
		return
	}

	if response.Message == "pong" {
		fmt.Println("Successful PING: received 'pong'")
	} else {
		errChan <- errors.New("did not receive pong")
		return
	}
	return
}

// Online will use tryOnline with up to 4 retries to verify if an identifier's instance and concierge is available.
func (i *Identifier) Online() bool {
	if !i.IsRemote() {
		return false
	}
	tryCounter := atomic.Int32{}
	fibDepth := atomic.Int32{}
	fibDepth.Store(0)
	for {
		hasErr := i.tryOnline()
		err := <-hasErr // wait to receive an err here from func
		if err != nil {
			fib := fibonacci(int(fibDepth.Add(1)))
			<-time.Tick(time.Duration(fib*10) * time.Millisecond) // 0ms, 20ms, 20ms, 30ms = max 70ms = 4 tries
			if tryCounter.Add(1) > 3 {
				return false
			}
			continue // retry the i.tryOnline()
		}
		return true
	}
}

const UUIDSeparator = `-`

// IdentifierForUUID parses a 2-6 part UUID in the form of part-part-part-part-part-part to part-part into an Identifier
func IdentifierForUUID(uuid string) *Identifier {
	i := &Identifier{}
	parts := strings.Split(uuid, UUIDSeparator)
	partsLen := len(parts)
	iInts, cInts, tInts, yInts, fInts, vInts := "", "", "", "", "", ""
	if partsLen == 6 { // UUID returns a i-c-t-y-f-v
		iInts = parts[0]
		cInts = parts[1]
		tInts = parts[2]
		yInts = parts[3]
		fInts = parts[4]
		vInts = parts[5]
	} else if partsLen == 5 { // UUID returns a i-c-t-y-f
		iInts = parts[0]
		cInts = parts[1]
		tInts = parts[2]
		yInts = parts[3]
		fInts = parts[4]
	} else if partsLen == 4 { // UUID returns a t-y-f-v
		tInts = parts[0]
		yInts = parts[1]
		fInts = parts[2]
		vInts = parts[3]
	} else if partsLen == 3 { // UUID returns a t-y-f
		tInts = parts[0]
		yInts = parts[1]
		fInts = parts[2]
	} else if partsLen == 2 { // UUID returns a t-f
		tInts = parts[0]
		fInts = parts[1]
	} else {
		i.e = errors.New("UUID format invalid")
		i.eat = time.Now().UTC()
	}

	x := func(s string) []rune {
		var result []rune
		parts := strings.Split(s, ".")
		partsLen := len(parts)
		if partsLen == 0 {
			return []rune{}
		}
		for j := 0; j < partsLen; j++ {
			if j > partsLen {
				break
			}
			k := parts[j]
			if len(k) == 0 {
				continue
			}
			possibleInt, intErr := strconv.Atoi(k)
			if intErr == nil {
				result = append(result, rune(possibleInt))
			}
		}
		return result
	}

	// UUID returns a i-c-t-y-f-v
	if len(iInts) > 0 {
		i.Instance = x(iInts)
	}
	if len(cInts) > 0 {
		i.Concierge = x(cInts)
	}
	if len(tInts) > 0 {
		i.Table = x(tInts)
	}
	if len(yInts) > 0 {
		year, intErr := strconv.Atoi(yInts)
		if intErr == nil {
			i.Year = int16(year)
		}
	}
	if len(fInts) > 0 {
		i.Fragment = x(fInts)
	}
	if len(vInts) > 0 {
		i.Version = ParseVersion(string(x(vInts)))
	}

	return i
}

// HasFile will iterate over f.files []*File to determine if the File.Path meets strings.HasSuffix in f.Path of filename
func (i *Identifier) HasFile(filename string) bool {
	for _, f := range i.files {
		if strings.HasSuffix(f.Path, filename) {
			return true
		}
	}
	return false
}

func (i *Identifier) GetFile(filename string) *File {
	return FileFromPath(filepath.Join(i.Path(), filename))
}

func (i *Identifier) GetFiles() {
	if i.IsRemote() {
		fileUrl := i.Path() + i.sPs() + ".files"
		file := FileFromPath(fileUrl)
		err := <-file.GetBytes()
		if err != nil {
			i.e = errors.Join(i.e, err)
			i.eat = time.Now().UTC()
		}

	} else if i.LocalExists() {
		walkErr := filepath.WalkDir(i.Path(), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() && len(d.Name()) > 2 {
				i.files = append(i.files, FileFromPath(path))
			}

			return nil
		})
		if walkErr != nil {
			i.e = errors.Join(i.e, walkErr)
			i.eat = time.Now().UTC()
		}
	}
}

// UUID returns a i-c-t-y-f-v where each component is a dot separated list of rune int32 strconv.Iota string values
func (i *Identifier) UUID() string {
	s := strings.Builder{}
	if len(i.Instance) > 0 && len(i.Concierge) > 0 {
		for j := 0; j < len(i.Instance); j++ {
			if j > len(i.Instance) {
				break
			}
			iij := i.Instance[j]
			siij := strconv.Itoa(int(iij))
			if j+1 < len(i.Instance) {
				siij += `.`
			}
			s.Write([]byte(siij))
		}
		s.Write([]byte(UUIDSeparator))
		for k := 0; k < len(i.Concierge); k++ {
			if k > len(i.Concierge) {
				break
			}
			ivpk := i.Concierge[k]
			sivpk := strconv.Itoa(int(ivpk))
			if k+1 < len(i.Concierge) {
				sivpk += `.`
			}
			s.Write([]byte(sivpk))
		}
		s.Write([]byte(UUIDSeparator))
	}
	if len(i.Table) > 0 {
		for r := 0; r < len(i.Table); r++ {
			if r > len(i.Table) {
				break
			}
			itr := i.Table[r]
			sitr := strconv.Itoa(int(itr))
			if r+1 < len(i.Table) {
				sitr += `.`
			}
			s.Write([]byte(sitr))
		}
		s.Write([]byte(UUIDSeparator))
	}
	if i.Year > 0 {
		syear := strconv.Itoa(int(i.Year))
		s.Write([]byte(syear))
		s.Write([]byte(UUIDSeparator))
	}
	if len(i.Fragment) > 0 {
		for q := 0; q < len(i.Fragment); q++ {
			if q > len(i.Fragment) {
				break
			}
			ifq := i.Fragment[q]
			sifq := strconv.Itoa(int(ifq))
			if q+1 < len(i.Fragment) {
				sifq += `.`
			}
			bw, wErr := s.Write([]byte(sifq))
			if wErr != nil {
				log.Printf("s.Write([]byte(sifq) received wErr %v", wErr)
				break
			}
			if bw != len(sifq) {
				log.Printf("s.Write([]byte(sifq) wrote %d bytes of %d bytes", bw, len(sifq))
			}
		}
		s.Write([]byte(UUIDSeparator))
	}
	if len(i.Version.String()) > 0 {
		iv := []rune(i.Version.String())
		for v := 0; v < len(iv); v++ {
			if v > len(iv) {
				break
			}
			ivv := iv[v]
			var ivvb []byte
			if !unicode.IsDigit(ivv) {
				ivvb = []byte(string(ivv))
				s.Write(ivvb)
				continue
			}

			sifq := strconv.Itoa(int(ivv))
			if v+1 < len(iv) {
				sifq += `.`
			}
			bw, wErr := s.Write([]byte(sifq))
			if wErr != nil {
				log.Printf("s.Write([]byte(sifq) received wErr %v", wErr)
				break
			}
			if bw != len(sifq) {
				log.Printf("s.Write([]byte(sifq) wrote %d bytes of %d bytes", bw, len(sifq))
			}
		}
	}
	return s.String()
}

func (i *Identifier) Error() error {
	return i.e
}

func (i *Identifier) ErrorAt() time.Time {
	return i.eat
}

type Err error

var ErrNotRemote Err = errors.New("not remote")

func (i *Identifier) IsRemote() bool {
	if errors.Is(i.e, ErrNotRemote) {
		return false
	}
	_, parseErr := url.Parse(string(i.Instance))
	if parseErr != nil {
		i.e = errors.Join(ErrNotRemote, parseErr)
		i.eat = time.Now().UTC()
		return false
	}
	if len(i.Concierge) == 0 {
		i.e = errors.Join(ErrNotRemote, errors.New("i.Concierge is unset and is required"))
		i.eat = time.Now().UTC()
		return false
	}
	return true
}

func (i *Identifier) LocalExists() bool {
	iInfo, statErr := os.Stat(i.Path())
	if statErr != nil || iInfo.Size() == 0 {
		i.e = statErr
		i.eat = time.Now().UTC()
		return false
	}
	return true
}

const URLSeparator = `/`

// sPs returns the URLSeparator which is a / used for combining components of the Identifier into a path
func (i *Identifier) sPs() string {
	return URLSeparator
}

func (i *Identifier) URL() string {
	return i.sPs() + i.sPs() + string(i.Instance) + i.sPs() + string(i.Table) + i.sPs() + fmt.Sprintf("%04d", i.Year) + i.sPs() + string(i.Fragment)
}

// ParseIdentifierURL takes a string from Dot() and returns the Identifier
// 	ParseIdentifierURL("idoread.com/valet/documents/2024/1")
//
//	returns
//
//	  Identifier{
//	    Instance: "idoread.com",
//      Concierge: "valet",
//	    Table: "documents",
//	    Year: 2024,
//	    Fragment: 1,
//	  }
func ParseIdentifierURL(iURL string) *Identifier {
	parts := strings.Split(iURL, URLSeparator)
	lp := len(parts)
	id := &Identifier{}

	if lp == 6 { // idoread.com/valet/documents/2024/1/v0.0.1
		id.Instance = []rune(parts[0])
		id.Concierge = []rune(parts[1])
		id.Table = []rune(parts[2])
		year, intErr := strconv.Atoi(parts[3])
		if intErr == nil {
			id.Year = int16(year)
		} else {
			id.e = errors.Join(id.e, intErr)
			id.eat = time.Now().UTC()
		}
		id.Fragment = Fragment(parts[4])
		id.Version = ParseVersion(parts[5])
		return id
	}

	if lp == 5 { // idoread.com/valet/documents/2024/1
		id.Instance = []rune(parts[0])
		id.Concierge = []rune(parts[1])
		id.Table = []rune(parts[2])
		year, intErr := strconv.Atoi(parts[3])
		if intErr == nil {
			id.Year = int16(year)
		} else {
			id.e = errors.Join(id.e, intErr)
			id.eat = time.Now().UTC()
		}
		id.Fragment = Fragment(parts[4])
	}

	if lp == 4 { // documents/2024/1/v0.0.1
		id.Table = []rune(parts[0])
		year, intErr := strconv.Atoi(parts[1])
		if intErr == nil {
			id.Year = int16(year)
		} else {
			id.e = errors.Join(id.e, intErr)
			id.eat = time.Now().UTC()
		}
		id.Fragment = Fragment(parts[2])
		id.Version = ParseVersion(parts[3])
	}

	if lp == 3 { // documents/2024/1
		id.Table = []rune(parts[0])
		year, intErr := strconv.Atoi(parts[1])
		if intErr == nil {
			id.Year = int16(year)
		} else {
			id.e = errors.Join(id.e, intErr)
			id.eat = time.Now().UTC()
		}
		id.Fragment = Fragment(parts[2])
	}

	if lp == 2 { // documents/1
		id.Table = []rune(parts[0])
		id.Fragment = Fragment(parts[1])
	}

	if lp < 2 {
		id.e = errors.Join(id.e, errors.New("missing fragment"))
		id.eat = time.Now().UTC()
	}

	return id
}

func (i *Identifier) Path() string {
	return strings.ReplaceAll(i.URL(), i.sPs(), string(os.PathSeparator))
}

func (i *Identifier) String() string {
	return fmt.Sprintf("%04d%s", i.Year, strings.ToUpper(string(i.Fragment)))
}
