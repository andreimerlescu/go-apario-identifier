package go_apario_identifier

import (
	`context`
	`encoding/json`
	`errors`
	`log`
	`net/url`
	`reflect`
	`strings`
	`time`
)

type Assign struct {
	A any
	V *Valet
	C *Cache
	B []byte
	E error
	S int64
}

func (v *Valet) AssignUnmarshalTargetType(a any) *Assign {
	rv := reflect.ValueOf(a)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &Assign{
			E: errors.New("argument must be a pointer"),
		}
	}
	return &Assign{
		A: a,
		V: v,
	}
}

func (a *Assign) MaybeAssignment() any {
	return a.A
}

func (a *Assign) Unmarshal() any {
	if a.E != nil {
		return nil
	}
	if a.A == nil {
		return nil
	}
	if len(a.B) == 0 {
		return nil
	}

	tempAnyType := reflect.TypeOf(a.A)
	if tempAnyType.Kind() == reflect.Ptr {
		tempAnyType = tempAnyType.Elem()
	}
	tempAny := reflect.New(tempAnyType).Interface()

	err := json.Unmarshal(a.B, &tempAny)
	if err != nil {
		a.E = err
		return nil
	}

	reflect.ValueOf(&a.A).Elem().Set(reflect.ValueOf(tempAny).Elem())
	return a.A
}

func (a *Assign) GetAnyError() (any, error) {
	if a.A == nil {
		return nil, errors.New(".A is nil")
	}

	if len(a.B) == 0 {
		return nil, errors.New("len(.B) is 0")
	}

	tempAnyType := reflect.TypeOf(a.A)
	if tempAnyType.Kind() == reflect.Ptr {
		tempAnyType = tempAnyType.Elem()
	}
	tempAny := reflect.New(tempAnyType).Interface()

	err := json.Unmarshal(a.B, &tempAny)
	if err != nil {
		return nil, err
	}

	reflect.ValueOf(&a.A).Elem().Set(reflect.ValueOf(tempAny).Elem())

	return a.A, nil
}

func (a *Assign) GetPathBytes(idPath string) *Assign {
	if a.E != nil {
		return a
	}
	identifierUrl, urlErr := url.Parse(idPath)
	if urlErr != nil {
		a.E = urlErr
		log.Printf("url.Parse(a.V.InitialPath + identifier.Path()) suppressed an err %v ; will check next type of combination", urlErr)
	} else {
		// currently yes this is a url, however with no return, this function will return for us if necessary
		if strings.Contains(identifierUrl.Path, "/valet") {
			// this is the remote valet service
			ctx, cancel := context.WithTimeout(context.Background(), a.V.RemoteTimeoutSeconds*time.Second)
			a.B = a.V.GetRemotePathFileBytes(idPath, ctx, cancel)
			a.S = int64(len(a.B))
			return a
		}
	}

	// this is a valid local path, process it as a filesystem request
	cache, cacheErr := a.V.GetCache(a.V.InitialPath)
	if cacheErr != nil {
		a.E = cacheErr
		log.Printf("GetPathBytes() -> a.V.GetCache(a.V.InitialPath) received err %v", cacheErr)
	}

	fileBytes, readErr := cache.SafeLoadBytes(idPath)
	if readErr != nil {
		a.E = readErr
		log.Printf("GetPathBytes() -> cache.SafeLoadBytes(a.V.InitialPath + idPath) received err %v", readErr)
	}

	a.E = nil
	a.B = fileBytes
	fileBytes = nil
	a.S = int64(len(a.B))

	return a
}
