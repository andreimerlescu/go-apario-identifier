package go_apario_identifier

import (
	`context`
	`encoding/json`
	`os`
	`path/filepath`
	`testing`
	`time`

	`github.com/andreimerlescu/go-apario-identifier/success`
)

type Example struct{}

func TestExample(t *testing.T) {
	// This is going to be a working example showing how this package is intended to be used within an application.
	// If you are looking at possibly using this package, this is what you can do with it. The README.md doesn't
	// clearly explain how it works without actually using it first to see how it works. The documentation about
	// the functions and such will need to be added in a future commit, but this working example will give you a
	// documentation go-to when looking at how something is implemented and used. The tests contained within this
	// file will show you the power of the package when it runs within your application.

	dbDir, dbDirOk := success. // cleaner way of interacting with test errors
		AnySucceeded( // a catcher that receives any, error
			os.MkdirTemp("", "users.db"), // actual task returns string, error
		). // a catcher that receives any, error
		SetM("os.MkdirTemp()"). // a helpful message
		SetT(t). // give the testing suite success
		Evaluate() // evaluate the result and return any, bool ; any must be typecast to string in this case
	// a few notes about the success package: the intended purpose of this is to feed an any, error response type into
	// a constructor of this data structure and allow me to assign to the any, error combination of data, a message
	// and the pointer to the testing suite. If the testing suite is provided, then a t.Errorf will trigger is error
	// is not nil; otherwise .Evaluate() will return a bool that will not interrupt or distract the example from
	// conventions used by the go programming language.

	var userDB string
	userDB, dbDirOk = dbDir.(string) // ensure that any is a string since os.MkdirTemp first response is a string type
	// if the DB was successfully created we don't need to escape and we can use the userDB string now too
	if !dbDirOk {
		return
	}

	// now with this directory, we can make it into a database. For users, I recommend using a countable database
	// so each user gets an incremental unique identifier for their ID. However lets create a normal database first.
	// then we can create a countable DB. but before we progress any further, lets create a context with a timeout
	// to ensure that if the test is running and a problem happens with the semaphores or mutexes we have plenty of
	// time to test before things bail on us.

	// This context will run for 3 minutes before canceling.
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	// THE VALET

	// The Valet is a special helper set of functions that allow you to safely interact with the database that was
	// just created. Why is this kind of database viable right now for this kind of application? Well, Project Apario
	// is a document based system and the system has the documents that we need. We need a safe way to concurrently
	// access those documents for reading/writing tasks. Traditional databases are too complex for what is needed
	// and to use a traditional database that is not purely written in Go, then you're kind of left on your own when
	// the open source project gets discontinued, bought out, starts charging money for it, or updates the package
	// that completely breaks your application. This database is not designed for high scale, but it is designed to
	// get the starting work necessary for the Project Apario decentralized and distributed appliance project.

	// Creating a new Valet only requires a string argument that is the path of the directory that will be used for
	// a database.
	valet := NewValet(userDB)

	// A regular Valet can also be created with a context so any of the for/select statements used in the package have
	// a safe way of exiting cleanly when a SIGINT or something similar is received by a watchdog and the context gets
	// canceled and the app is expected to clean up nicely. It is HIGHLY recommended that you use a context. In this
	// example, we'll be using the context.
	valet = NewValetWithContext(ctx, userDB)

	// one low cost functionality provided to the valet is the SafetyCheck. This ensures that the Semaphores and Mutexes
	// map are safely established so we don't get any write to nil map errors. SafetyCheck is called throughout the
	// package itself since its low cost and if you need to call it, you can too; its exported.
	valet.SafetyCheck()

	// A countable database is nothing more than a directory that contains a text file that is called .lastid that
	// can either be blank or contain digits only for an int64 number. Any invalid data or other filename, the
	// concept is ignored and the database will use a random ID generator to assign NewID.
	if !success. // lets get some success
		Succeeded( // the error catcher
			valet.NewCountableDatabase(userDB), // the actual action
		).
		SetM("NewCountableDatabase(userDB)"). // a debug message in case the error is not helpful
		SetT(t). // connect the testing suite handler
		Result() { // handle the error and interact with the testing suite if necessary to report the t.Errorf
		return
	}

	// now we have a countable database! lets start using it!!
	var uid1 *Identifier
	maybeId, idOk := success.AnySucceeded(valet.NextID(userDB)).SetT(t).SetM("valet.NextID(userDB)").Evaluate()
	uid1, idOk = maybeId.(*Identifier)
	if !idOk {
		return
	}

	// now a new user has an ID of 1 (base 36)
	// now lets say that we have some structure of data that represents a user
	type sampleUser struct {
		Email string      `json:"email"`
		Id    *Identifier `json:"id"`
	}

	// what we want to do now is save this record to disk as a JSON using the valet service.
	user1 := sampleUser{
		Email: uid1.String(),
		Id:    uid1,
	}

	// save sample user1 to disk now
	var cache *Cache
	maybeCache, cacheOk := success.AnySucceeded(valet.GetCache(userDB)).SetM("valet.GetCache(userDB)").SetT(t).Evaluate()
	cache, cacheOk = maybeCache.(*Cache)
	if !cacheOk {
		return
	}

	accountFile := filepath.Join(cache.Path, uid1.Path(), "account.json")

	// now we need to get the bytes of the payload
	maybeBytes, marshalOk := success.AnySucceeded(json.Marshal(user1)).SetM("json.Marshal(user1)").SetT(t).Evaluate()
	var accountBytes []byte
	accountBytes, marshalOk = maybeBytes.([]byte)
	maybeBytes = nil
	if !marshalOk || len(accountBytes) == 0 {
		return
	}

	// now that we have the bytes we can then write them to the database entry at that identifier path
	// this is where accountFile and accountBytes are leveraging the semaphores and wait groups in the control structure
	// of the access.
	if !success.Succeeded(cache.SafeWriteBytes(accountFile, accountBytes)).SetM("cache.SafeWriteBytes(accountFile)").SetT(t).Result() {
		return
	}

	// for example, lets say that we have a scenario where we want to knowingly throw an error. this can be useful for
	// testing purposes. consider the following.
	cache.Mutex(accountFile).Lock()

	// when I try to use the cache to access a safe lock on the file, I see this:
	hadNoErr := success.Succeeded(cache.LockIdentifier(uid1.String())).SetM("cache.LockIdentifier(uid1)").SetT(t).ExpectNoError()
	if !hadNoErr {
		t.Errorf("expected cache.LockIdentifier to not error out ; expected = true ; got = %v", hadNoErr)
		return
	}

	// however we can release the lock
	cache.Mutex(accountFile).Unlock()

	// with that released we can re-try locked
	receivedErr := cache.LockIdentifier(uid1.String())
	if receivedErr == nil {
		t.Errorf("failed cache.LockIdentifier(uid1.String()) ; expected = error ; got = %v", receivedErr)
		return
	}

	// perfect, now lets load the data back in, lets assign this user to the name of admin
	var admin sampleUser
	var adminBytes []byte
	var adminErr error
	var adminOk bool
	adminBytes, adminErr = cache.SafeLoadBytes(accountFile)
	if adminErr != nil {
		t.Errorf("not using success.Succeeded here... but cache.SafeLoadBytes(accountFile) failed due to err %v", adminErr)
		return
	}
	adminOk = success.Succeeded(json.Unmarshal(adminBytes, &admin)).SetM("json.Unmarshal(adminBytes, &admin)").SetT(t).Result()
	if !adminOk {
		t.Errorf("expected adminOk to be true ; got %v [adminErr = %v [should be nil]]", adminOk, adminErr)
		return
	}

	// now we can work with the admin user
	admin.Email = "admin@my.projectapario.com"

	maybeBytes, adminOk = success.AnySucceeded(json.Marshal(admin)).SetM("json.Marshal(admin)").SetT(t).Evaluate()
	adminBytes, adminOk = maybeBytes.([]byte)
	maybeBytes = nil
	if !adminOk {
		t.Errorf("expected adminOk to be true ; got = %v", adminOk)
		return
	}

	// now lets save the admin back, but this time using only the value inside of admin (say from inside a func)
	if !success.Succeeded(cache.SafeWriteBytes(filepath.Join(cache.Path, admin.Id.Path(), "account.json"), adminBytes)).SetM("cacheSafeWriteBytes(filepath.Join())").SetT(t).Result() {
		return
	}

	// now once we have the bytes written back to the disk we can manually lock the identifier to prevent future writes
	lockAt := time.Now().UTC()
	lockFile := filepath.Join(cache.Path, admin.Id.Path(), ".locked")
	err := cache.writeTimestampFile(admin.Id.String(), ".locked", lockAt)
	if err != nil {
		t.Errorf("cache.writeTimestampFile() received err %v", err)
		return
	}

	// verify the actual lock file at the OS level now, outside of the package just to verify
	lockInfo, infoErr := os.Stat(lockFile)
	if infoErr != nil {
		t.Errorf("os.Stat(lockFile) received err %v", infoErr)
		return
	}

	// Verify that its a normal file
	if lockInfo.IsDir() {
		t.Errorf("lockInfo.IsDir() returned true ; expected false")
		return
	}

	// verify that the size is good too
	if lockInfo.Size() < 1 {
		t.Errorf("lockInfo.Size() < 1 returned true ; expected false")
		return
	}

	// PERFECT! The lock file is present. Now let's read it and analyze it. This os.Stat is performed inside this
	// next function and is only in this example for demonstration purposes.

	// now lets verify what was written to the file
	lockedAt, readErr := cache.readTimestampFile(admin.Id.Path(), ".locked")
	if readErr != nil {
		t.Errorf("cache.readTimestampFile(admin.Id.Path(), '.locked') failed with err %v", readErr)
		return
	}

	// verify that the lockAt time.Time is equal to the lockedAt time.Time
	if lockAt.Unix() != lockedAt.Unix() {
		t.Errorf("expected lockAt.Equal(lockedAt) to be true got false")
		return
	}

	// perfect now we've locked the directory, now its time for us to try to do something to that account.json file again
	failure := cache.LockIdentifier(admin.Id.String())
	if failure == nil {
		t.Errorf("we expected an error with cache.LockIdentifier(admin.id.String()) ; got = %v", failure)
		return
	}

	// In the real world, when the cache.LockIdentifier fails it means that something else has locked the file
	// and the file cannot be written to safely. If the file is locked, it also shouldn't be read because it
	// ultimately means that the record is actively being modified.

	// now that we have the identifier in a locked state, we need to unlock it.
	cache.UnlockIdentifier(admin.Id.String())

	// now you should be able to lock your identifier
	failure = cache.LockIdentifier(admin.Id.String())
	if failure != nil {
		t.Errorf("we expected no error with cache.LockIdentifier(admin.id.String()) ; got = %v", failure)
		return
	}

	cache.UnlockIdentifier(admin.Id.String())

	// Now we are going to use the magic of Assign
	var response any
	newAdminPath := filepath.Join(cache.Path, admin.Id.Path(), "account.json")
	response = valet.AssignUnmarshalTargetType(&sampleUser{}).GetPathBytes(newAdminPath).Unmarshal()
	if response == nil {
		t.Errorf("")
	}
	newAdmin, isNewAdmin := response.(sampleUser)
	if newAdmin.Email != admin.Email {
		t.Errorf("expected newAdmin.Email == admin.Email since they are the same record loaded in two different ways ; isNewAdmin = %v", isNewAdmin)
		return
	}

}
