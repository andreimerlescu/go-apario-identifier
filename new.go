package go_apario_identifier

import (
	`context`
	`sync`

	sema `github.com/andreimerlescu/go-sema`
)

func NewValetWithContext(ctx context.Context, databasePath string) *Valet {
	return &Valet{
		ctx:         ctx,
		InitialPath: databasePath,
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
		ctx:         ctx,
		InitialPath: databasePath,
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
