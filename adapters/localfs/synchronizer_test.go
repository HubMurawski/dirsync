//go:build linux

package localfs

import (
	"dirsync/app/logger"
	"dirsync/mocks"
	"os"
	"path/filepath"
	"testing"
)

func TestSynchronizer_purgeMissing(t *testing.T) {
	tests := []struct {
		name          string
		s             *Synchronizer
		initDir       func(t *testing.T, srcpath, dstPath string)
		initMapping   func(dst string, s *Synchronizer)
		expectedFiles []expectedFile
	}{
		{
			name: "basic delete",
			s:    NewSynchronizer("", "", true, logger.New()),
			initMapping: func(dst string, s *Synchronizer) {
				s.dst = dst
				s.dstFS["target_only.txt"] = &mocks.DirEntryMock{
					IsDirFunc: func() bool { return false },
				}
			},
			expectedFiles: []expectedFile{
				{
					path:        "target_only.txt",
					shouldExist: false,
				},
			},
		},
		{
			name: "dir delete",
			s:    NewSynchronizer("", "", true, logger.New()),
			initMapping: func(dst string, s *Synchronizer) {
				s.dst = dst
				s.dstFS["target_only.txt"] = &mocks.DirEntryMock{
					IsDirFunc: func() bool { return false },
				}
				s.dstFS["target_only_dir"] = &mocks.DirEntryMock{
					IsDirFunc: func() bool { return true },
				}
				s.dstFS[filepath.Join("target_only_dir", "target_only2.txt")] = &mocks.DirEntryMock{
					IsDirFunc: func() bool { return false },
				}
			},
			initDir: func(t *testing.T, srcpath, dstPath string) {
				if err := os.MkdirAll(filepath.Join(dstPath, "target_only_dir"), 0755); err != nil {
					t.Fatalf("failed to create target directory: %s", err)
				}
				f, err := os.Create(filepath.Join(dstPath, "target_only_dir", "target_only2.txt"))
				if err != nil {
					t.Fatalf("failed to create target file: %s", err)
				}
				f.Close()
			},
			expectedFiles: []expectedFile{
				{
					path:        "target_only.txt",
					shouldExist: false,
				},
				{
					path:        "target_only_dir",
					shouldExist: false,
				},
				{
					path:        filepath.Join("target_only_dir", "target_only2.txt"),
					shouldExist: false,
				},
			},
		},
		{
			name: "empty dir delete",
			s:    NewSynchronizer("", "", true, logger.New()),
			initMapping: func(dst string, s *Synchronizer) {
				s.dst = dst
				s.dstFS["target_only.txt"] = &mocks.DirEntryMock{
					IsDirFunc: func() bool { return false },
				}
				s.dstFS["target_only_dir"] = &mocks.DirEntryMock{
					IsDirFunc: func() bool { return true },
				}
			},
			initDir: func(t *testing.T, srcpath, dstPath string) {
				if err := os.MkdirAll(filepath.Join(dstPath, "target_only_dir"), 0755); err != nil {
					t.Fatalf("failed to create target directory: %s", err)
				}
			},
			expectedFiles: []expectedFile{
				{
					path:        "target_only.txt",
					shouldExist: false,
				},
				{
					path:        "target_only_dir",
					shouldExist: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srdDir, dstDir := setupTestDirectories(t)
			defer cleanupTestDirectories(srdDir, dstDir)

			tt.initMapping(dstDir, tt.s)
			if tt.initDir != nil {
				tt.initDir(t, srdDir, dstDir)
			}

			tt.s.purgeMissing()

			for i, f := range tt.expectedFiles {
				tt.expectedFiles[i].path = filepath.Join(dstDir, f.path)
			}
			assertPresence(t, tt.expectedFiles)
		})
	}
}

// setupTestDirectories creates test directories with sample files
func setupTestDirectories(t *testing.T) (string, string) {
	// Create temporary directories
	sourceDir, err := os.MkdirTemp("", "dirsync-src-*")
	if err != nil {
		t.Fatalf("failed to create source directory: %v", err)
	}

	targetDir, err := os.MkdirTemp("", "dirsync-dst-*")
	if err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}

	// Create sample files in source directory
	createSampleFile(t, filepath.Join(sourceDir, "file1.txt"), "content1")
	createSampleFile(t, filepath.Join(sourceDir, "file2.txt"), "content2")

	// Create subdirectory with a file
	subDir := filepath.Join(sourceDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	createSampleFile(t, filepath.Join(subDir, "file3.txt"), "content3")

	// Create a file in target that doesn't exist in source (for deletion testing)
	createSampleFile(t, filepath.Join(targetDir, "target_only.txt"), "target content")

	return sourceDir, targetDir
}

// createSampleFile creates a sample file with given content
func createSampleFile(t *testing.T, path, content string) {
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create sample file %s: %v", path, err)
	}
}

// cleanupTestDirectories removes test directories
func cleanupTestDirectories(src, dst string) {
	os.RemoveAll(src)
	os.RemoveAll(dst)
}

func assertPresence(t *testing.T, e []expectedFile) {
	for _, file := range e {
		if _, err := os.Stat(file.path); os.IsNotExist(err) == file.shouldExist {
			s := "should"
			if !file.shouldExist {
				s += " not"
			}
			t.Errorf("File %s exist: %s", s, file.path)
		}
	}
}

type expectedFile struct {
	path        string
	shouldExist bool
}

func TestSynchronizer_Run(t *testing.T) {
	tests := []struct {
		name          string
		initDir       func(t *testing.T, srcpath, dstPath string)
		expectedFiles []expectedFile
		wantErr       bool
	}{
		{
			name: "basic sync",
			expectedFiles: []expectedFile{
				{
					path:        "file1.txt",
					shouldExist: true,
				},
				{
					path:        "file2.txt",
					shouldExist: true,
				},
				{
					path:        "subdir",
					shouldExist: true,
				},
				{
					path:        filepath.Join("subdir", "file3.txt"),
					shouldExist: true,
				},
			},
		},
		{
			name: "empty subdir",
			expectedFiles: []expectedFile{
				{
					path:        "file1.txt",
					shouldExist: true,
				},
				{
					path:        "file2.txt",
					shouldExist: true,
				},
				{
					path:        "subdir",
					shouldExist: true,
				},
				{
					path:        filepath.Join("subdir", "file3.txt"),
					shouldExist: false,
				},
			},
			initDir: func(t *testing.T, srcPath, dstPath string) {
				err := os.Remove(filepath.Join(srcPath, "subdir", "file3.txt"))
				if err != nil {
					t.Fatalf("Failed to delete file: %s", err)
				}
			},
		},
		{
			name: "update",
			expectedFiles: []expectedFile{
				{
					path:        "file1.txt",
					shouldExist: true,
				},
				{
					path:        "file2.txt",
					shouldExist: true,
				},
				{
					path:        "subdir",
					shouldExist: true,
				},
				{
					path:        filepath.Join("subdir", "file3.txt"),
					shouldExist: false,
				},
			},
			initDir: func(t *testing.T, srcpath, dstPath string) {
				err := os.Remove(filepath.Join(srcpath, "subdir", "file3.txt"))
				if err != nil {
					t.Fatalf("Failed to delete file: %s", err)
				}
				createSampleFile(t, filepath.Join(srcpath, "target_only.txt"), "aezakmi")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srdDir, dstDir := setupTestDirectories(t)
			defer cleanupTestDirectories(srdDir, dstDir)
			if tt.initDir != nil {
				tt.initDir(t, srdDir, dstDir)
			}

			s := NewSynchronizer(srdDir, dstDir, false, logger.New())
			if err := s.Run(); (err != nil) != tt.wantErr {
				t.Errorf("Synchronizer.Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			for i, f := range tt.expectedFiles {
				tt.expectedFiles[i].path = filepath.Join(dstDir, f.path)
			}
			assertPresence(t, tt.expectedFiles)
		})
	}
}
