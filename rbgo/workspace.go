package rbgo

import (
	"path/filepath"
	"os"
	"fmt"
	"strings"
	"regexp"
)

// Workspace
type Workspace struct {
	sourceEntry string
	objectPath  string
	ExcludeDirs ExcludeDirs
	PackageRoot PackageRootFinder
	Package     *PackageRepository
}

func NewWorkspace(path string) (*Workspace, error) {
	w := &Workspace{
		ExcludeDirs: ExcludeDirs([]string{".git", ".idea"}),
		PackageRoot: PackageRootFinder([]*regexp.Regexp{}),
		Package: new(PackageRepository).Init(),
	}
	_, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	w.sourceEntry = path
	//w.objectPath = filepath.Join(path, "pkg")
	s := filepath.Join(path, "src")
	if _, err = os.Stat(s); err == nil {
		w.sourceEntry = s
	}
	w.AddPackageRoot("golang.org/x/[a-zA-Z0-9_-]+")
	w.AddPackageRoot("github.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+")
	w.AddPackageRoot("bitbucket.org/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+")
	return w, err
}

func (w *Workspace) AddPackageRoot(str string) {
	w.PackageRoot = append(w.PackageRoot, regexp.MustCompile(str))
}

func (w *Workspace) Walk(f func(string) error) error {
	return filepath.Walk(w.sourceEntry, func(path string, fi os.FileInfo, err error) error {
		if !fi.IsDir() {
			return nil
		}
		if w.ExcludeDirs.Contains(path) {
			return nil
		}
		return f(path)
	})
}

func (w *Workspace) NewPackage(path string) *Package {
	return NewPackage(w.sourceEntry, path)
}

func (w *Workspace) Init() error {
	err := w.Walk(func(path string) error {
		pkg := w.Package.FindByPath(path)
		if pkg == nil {
			pkg = w.NewPackage(path)
		}
		err := pkg.Scan(w.PackageRoot)
		if err == nil {
			w.Package.Put(pkg)
			fmt.Printf("import: %s\n", pkg.FullName)
		} else if err != SourceNotFound {
			fmt.Printf("fresh: %s, %v\n", pkg.WatchPath, err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	w.Package.UpdateDepends()
	return nil
}

// ExcludeDirs
type ExcludeDirs []string

func (dirs ExcludeDirs) Contains(path string) bool {
	pathList := strings.Split(path, "/")
	for i := len(pathList) - 1; i > -1; i -= 1 {
		for _, d := range dirs {
			if pathList[i] == d {
				return true
			}
		}
	}
	return false
}
