package mocks

import "io/fs"

type DirEntryMock struct {
	NameFunc  func() string
	IsDirFunc func() bool
	TypeFunc  func() fs.FileMode
	InfoFunc  func() (fs.FileInfo, error)
}

func (m *DirEntryMock) Name() string {
	return m.NameFunc()
}
func (m *DirEntryMock) IsDir() bool {
	return m.IsDirFunc()
}
func (m *DirEntryMock) Type() fs.FileMode {
	return m.TypeFunc()
}
func (m *DirEntryMock) Info() (fs.FileInfo, error) {
	return m.InfoFunc()
}
