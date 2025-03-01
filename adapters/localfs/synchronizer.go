//go:build linux

package localfs

import (
	"dirsync/app/logger"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"time"
)

// Synchronizer handles the file synchronization logic
type Synchronizer struct {
	src, dst      string
	deleteMissing bool
	srcFS         map[string]fs.DirEntry
	dstFS         map[string]fs.DirEntry
	l             *logger.Logger
}

func NewSynchronizer(src, dst string, deleteMissing bool, l *logger.Logger) *Synchronizer {
	return &Synchronizer{
		src:           src,
		dst:           dst,
		deleteMissing: deleteMissing,
		srcFS:         make(map[string]fs.DirEntry),
		dstFS:         make(map[string]fs.DirEntry),
		l:             l,
	}
}

// Run starts the synchronization process
func (s *Synchronizer) Run() error {
	if err := s.mapDirs(); err != nil {
		return err
	}
	s.processSync()

	return nil
}

// mapDirs scans directories and populates synchronizer mapping with files rel path and file metadata
func (s *Synchronizer) mapDirs() error {
	absSrc, err := filepath.Abs(s.src)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path (%s): %w", s.src, err)
	}
	if err := filepath.WalkDir(absSrc, s.walkFn(absSrc, s.srcFS)); err != nil {
		return fmt.Errorf("failed to walk source dir: %w", err)
	}

	if _, err := os.Stat(s.dst); os.IsNotExist(err) {
		info, err := os.Lstat(absSrc)
		if err != nil {
			s.l.Error("failed to read file info: %s\n", err)
			return err
		}
		s.l.Info("creating destination directory: %s\n", s.dst)
		if err := os.MkdirAll(s.dst, info.Mode()); err != nil {
			s.l.Error("failed to create destination directory: %s\n", err)
			return err
		}
	}

	absDst, err := filepath.Abs(s.dst)
	if err != nil {
		return fmt.Errorf("failed to resolve absolute path (%s): %w", s.dst, err)
	}
	if err := filepath.WalkDir(absDst, s.walkFn(absDst, s.dstFS)); err != nil {
		return fmt.Errorf("failed to walk destination dir: %w", err)
	}

	return nil
}

func (s *Synchronizer) walkFn(root string, m map[string]fs.DirEntry) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.l.Error("%s error: %s", path, err)
			return nil
		}
		if path == root {
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			s.l.Error("failed to resolve rel path (%s): %s", path, err)
			return nil
		}
		m[relPath] = d
		return nil
	}
}

// processSync checks differences in mapping and upserts files from source to destination directory
func (s *Synchronizer) processSync() {
	//Get and sort path first so its easier to build directiories with the same permissions
	paths := make([]string, 0, len(s.srcFS))
	for relPath := range s.srcFS {
		paths = append(paths, relPath)
	}
	slices.Sort(paths)

	for _, relPath := range paths {
		if dstEntry, exists := s.dstFS[relPath]; exists {
			if s.needsUpdate(s.srcFS[relPath], dstEntry) {
				s.l.Info("overwriting  %s", relPath)
				if err := s.syncFile(relPath); err != nil {
					s.l.Error("failed to overwrite: %s", err)
				}
			}
			delete(s.dstFS, relPath)
			continue
		}
		s.l.Info("copying  %s", relPath)
		if err := s.syncFile(relPath); err != nil {
			s.l.Error("failed to copy: %s", err)
		}
		delete(s.dstFS, relPath)
	}

	if s.deleteMissing && len(s.dstFS) > 0 {
		s.purgeMissing()
	}
}

// needsUpdate compares file metadata and returns true if there is a mismatch
func (s *Synchronizer) needsUpdate(src, dst fs.DirEntry) bool {
	if src.IsDir() != dst.IsDir() || src.Type() != dst.Type() {
		return true
	}
	srcInfo, err := src.Info()
	if err != nil {
		s.l.Error("failed to read source file info: %s", err)
	}
	dstInfo, err := dst.Info()
	if err != nil {
		s.l.Error("failed to read destination file info: %s", err)
	}

	return srcInfo.Size() != dstInfo.Size() || srcInfo.ModTime().After(dstInfo.ModTime())
}

// syncFile copies/overwrites a file from source to destination
func (s *Synchronizer) syncFile(relPath string) error {
	srcPath := filepath.Join(s.src, relPath)
	dstPath := filepath.Join(s.dst, relPath)
	srcInfo, err := s.srcFS[relPath].Info()
	if err != nil {
		return fmt.Errorf("failed to read file info: %w", err)
	}

	if srcInfo.IsDir() {
		if err := os.MkdirAll(dstPath, srcInfo.Mode()); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		return nil
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := os.Chtimes(dstPath, time.Now(), srcInfo.ModTime()); err != nil {
		return fmt.Errorf("failed to set modification time: %w", err)
	}

	return nil
}

// purgeMissing deletes files and directories from the destination directory if they do not exist in the source directory
func (s *Synchronizer) purgeMissing() {
	for relPath, entry := range s.dstFS {
		if entry.IsDir() {
			continue
		}
		dstPath := filepath.Join(s.dst, relPath)
		s.l.Info("deleting  %s", dstPath)
		if err := os.Remove(dstPath); err != nil {
			s.l.Error("failed to delete file: %s", err)
		}
		delete(s.dstFS, relPath)
	}

	dirsToDel := make([]string, 0, len(s.dstFS))
	for k := range s.dstFS {
		delete(s.dstFS, k)
		dirsToDel = append(dirsToDel, k)
	}

	slices.Sort(dirsToDel)
	slices.Reverse(dirsToDel)

	for _, dir := range dirsToDel {
		dstPath := filepath.Join(s.dst, dir)
		s.l.Info("deleting  %s", dir)
		if err := os.Remove(dstPath); err != nil {
			s.l.Error("failed to delete directory: %s", err)
		}
	}
}
