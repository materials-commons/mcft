package ft

import (
	"sync"

	"github.com/apex/log"
)

var mapMutex sync.Mutex
var mutexes = make(map[int]*sync.Mutex)

func acquireProjectMutex(projectID int) {
	mapMutex.Lock()
	defer mapMutex.Unlock()
	var p sync.Mutex
	projectMutex, ok := mutexes[projectID]
	if !ok {
		projectMutex = &p
		mutexes[projectID] = projectMutex
	}
	projectMutex.Lock()
}

func releaseProjectMutex(projectID int) {
	m, ok := mutexes[projectID]
	if !ok {
		log.Errorf("releaseProjectMutex called on project (%d) with no mutex", projectID)
		return
	}

	m.Unlock()
}
