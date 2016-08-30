package rbgo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"runtime"
	"io"
	"io/ioutil"
	"errors"
)

type TaskFactory struct {
	Package *PackageRepository
}

func (f *TaskFactory) New(dirName string) (*Task, error) {
	pkg := f.Package.FindByPath(dirName)
	if pkg == nil {
		return nil, fmt.Errorf("Package not found: `%s`", dirName)
	}
	return newJob(pkg, f.Package), nil
}

func newJob(pkg *Package, r *PackageRepository) *Task {
	return &Task{
		PackageName: pkg.FullName,
		SourcePath: pkg.SourcePath,
		ObjectPath: pkg.ObjectPath,
		Package: pkg,
		repo: r,
	}
}

type Task struct {
	PackageName string
	SourcePath  string
	ObjectPath  string
	Package     *Package
	repo        *PackageRepository
}

func (t *Task) Build() error {

	arguments := []string{"build"}
	arguments = append(arguments, ([]string{"-o", t.ObjectPath, t.SourcePath})...)
	command := exec.Command("go", arguments...)
	command.Dir = t.Package.WorkDir
	// Set GOPATH
	goPath := ""
	env := make([]string, 0, len(os.Environ()))
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOPATH=") {
			goPath = strings.Split(e, "=")[1]
		} else {
			env = append(env, e)
		}
	}
	sep := ":"
	if runtime.GOOS == "windows" {
		sep = ";"
	}
	goPath = strings.Join([]string{t.Package.WorkDir, goPath}, sep)
	env = append(env, fmt.Sprintf("%s=%s", "GOPATH", goPath))
	command.Env = env

	//
	stderr, err := command.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	if err := command.Start(); err != nil {
		return err
	}

	io.Copy(os.Stdout, stdout)
	errBuf, _ := ioutil.ReadAll(stderr)

	if err := command.Wait(); err != nil {
		return errors.New(string(errBuf))
	}

	return nil
}

func (t *Task) FindDepends() (*Task, error) {
	dep, err := t.findDepends(t.Package)
	if err != nil {
		return nil, err
	}
	if t.Package.ObjectPath != dep.ObjectPath {
		return newJob(dep, t.repo), nil
	}
	return nil, nil
}

func (t *Task) findDepends(pkg *Package) (*Package, error) {
	if len(pkg.MissingImports) > 0 {
		return nil, fmt.Errorf("Package not found '%s'", pkg.MissingImports[0])
	}
	for _, name := range pkg.Imports {
		imp := t.repo.FindByImportName(name)
		if imp == nil {
			continue
		}
		dep, err := t.findDepends(imp)
		if err != nil {
			return nil, err
		}
		if dep != nil {
			return dep, nil
		}
	}
	s, err := os.Stat(pkg.ObjectPath)
	if err != nil {
		return pkg, nil
	}
	if s.ModTime().Before(pkg.ModTime) {
		return pkg, nil
	}
	return nil, nil
}
