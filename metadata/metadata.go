package metadata

import "fmt"

type ManagerInterface interface {
	Sample()
}

type Manager struct{}

func (m *Manager) Sample() {
	fmt.Println("Good!!!")
}

func GetManager() *Manager {
	return &Manager{}
}
