package rbgo

import (
	"testing"
	"regexp"
	"os"
)

func TestTask_Build(t *testing.T) {
	remove := func(path string) {
		err := os.Remove(path)
		if err != nil {
			t.Fatal(err)
		}
	}
	finder := PackageRootFinder([]*regexp.Regexp{})
	finder = append(finder, regexp.MustCompile("github.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+"))
	sourceRoot := "../example/src"
	repo := new(PackageRepository).Init()
	pkg1 := NewPackage(sourceRoot, "../example/src/hoge/piyo")
	pkg2 := NewPackage(sourceRoot, "../example/src/vendor/github.com/kai-zoa/geeyoko")
	pkg3 := NewPackage(sourceRoot, "../example/src/vendor/github.com/kai-zoa/yokohama")
	pkgAll := []*Package{pkg1, pkg2, pkg3}
	for _, pkg := range pkgAll {
		pkg.Scan(finder)
		repo.Put(pkg)
	}
	repo.UpdateDepends()
	factory := TaskFactory{Package: repo}
	task, err := factory.New(pkg1.WatchPath)
	if err != nil {
		t.Fatal(err)
	}
	// step 1
	dep, err := task.FindDepends()
	if err != nil {
		t.Fatal(err)
	}
	if dep == nil {
		t.Fatal("nodep")
	}
	if a, e := dep.PackageName, "github.com/kai-zoa/yokohama"; a != e {
		err := "mismatch"
		t.Fatalf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if err := dep.Build(); err != nil {
		t.Fatal(err)
	}
	defer remove(dep.ObjectPath)
	if _, err := os.Stat(dep.ObjectPath); err != nil {
		t.Fatal(err)
	}
	// step 2
	dep, err = task.FindDepends()
	if err != nil {
		t.Fatal(err)
	}
	if dep == nil {
		t.Fatal("nodep")
	}
	if a, e := dep.PackageName, "github.com/kai-zoa/geeyoko"; a != e {
		err := "mismatch"
		t.Fatalf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if err := dep.Build(); err != nil {
		t.Fatal(err)
	}
	defer remove(dep.ObjectPath)
	if _, err := os.Stat(dep.ObjectPath); err != nil {
		t.Fatal(err)
	}
	// step 3
	dep, err = task.FindDepends()
	if err != nil {
		t.Fatal(err)
	}
	if dep != nil {
		t.Fatalf("Dependency remaining: `%s`", dep.ObjectPath)
	}
	if err := task.Build(); err != nil {
		t.Fatal(err)
	}
	defer remove(task.ObjectPath)
	if _, err := os.Stat(task.ObjectPath); err != nil {
		t.Fatal(err)
	}
}
