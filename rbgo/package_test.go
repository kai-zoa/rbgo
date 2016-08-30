package rbgo

import (
	"testing"
	"regexp"
	"reflect"
	"time"
	"path/filepath"
)

func TestPackage_Fresh(t *testing.T) {
	finder := PackageRootFinder([]*regexp.Regexp{})
	finder = append(finder, regexp.MustCompile("github.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+"))
	sourceRoot := "../example/src"
	packagePath := "../example/src/hoge/piyo"
	pkg := NewPackage(sourceRoot, packagePath)
	pkg.Scan(finder)
	if a, e := pkg.Name, "piyo"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.FullName, "hoge/piyo"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.InVendor, false; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.SourceCount, 2; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.ProjectName, ""; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.WatchPath, packagePath; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.SourcePath, "../example/src/hoge/piyo"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.ObjectPath, "../example/pkg/hoge/piyo.a"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	wd, _ := filepath.Abs("../example")
	if a, e := pkg.WorkDir, wd; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.ModTime, (time.Time{}); a.Before(e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.Imports, []string{"github.com/kai-zoa/geeyoko"}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.Referrers, []*Package{}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.MissingImports, []string{}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
}

func TestPackage_Fresh_Vendor_NoDep(t *testing.T) {
	finder := PackageRootFinder([]*regexp.Regexp{})
	finder = append(finder, regexp.MustCompile("github.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+"))
	sourceRoot := "../example/src"
	packagePath := "../example/src/vendor/github.com/kai-zoa/yokohama/piyo"
	pkg := NewPackage(sourceRoot, packagePath)
	pkg.Scan(finder)
	if a, e := pkg.Name, "piyo"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.FullName, "github.com/kai-zoa/yokohama/piyo"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.InVendor, true; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.ProjectName, "github.com/kai-zoa/yokohama"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.SourcePath, "../example/src/vendor/github.com/kai-zoa/yokohama"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.ObjectPath, "../example/pkg/vendor/github.com/kai-zoa/yokohama.a"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	wd, _ := filepath.Abs("../example")
	if a, e := pkg.WorkDir, wd; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.Imports, []string{}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
}

func TestPackage_Fresh_Vendor_HasDep(t *testing.T) {
	finder := PackageRootFinder([]*regexp.Regexp{})
	finder = append(finder, regexp.MustCompile("github.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+"))
	sourceRoot := "../example/src"
	packagePath := "../example/src/vendor/github.com/kai-zoa/geeyoko"
	pkg := NewPackage(sourceRoot, packagePath)
	pkg.Scan(finder)
	if a, e := pkg.Name, "geeyoko"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.FullName, "github.com/kai-zoa/geeyoko"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.InVendor, true; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.ProjectName, "github.com/kai-zoa/geeyoko"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg.Imports, []string{"github.com/kai-zoa/yokohama"}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
}

func TestPackageRepository_Fresh(t *testing.T) {
	finder := PackageRootFinder([]*regexp.Regexp{})
	finder = append(finder, regexp.MustCompile("github.com/[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+"))
	sourceRoot := "../example/src"
	repo := new(PackageRepository).Init()
	pkg1 := NewPackage(sourceRoot, "../example/src/vendor/github.com/kai-zoa/geeyoko")
	pkg1.Scan(finder)
	repo.Put(pkg1)
	pkg2 := NewPackage(sourceRoot, "../example/src/hoge/piyo")
	pkg2.Scan(finder)
	repo.Put(pkg2)
	repo.UpdateDepends()
	if a, e := pkg1.Imports, []string{"github.com/kai-zoa/yokohama"}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg1.MissingImports, []string{"github.com/kai-zoa/yokohama"}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg1.Referrers, []*Package{pkg2}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg2.Imports, []string{"github.com/kai-zoa/geeyoko"}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg2.MissingImports, []string{}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := pkg2.Referrers, []*Package{}; !reflect.DeepEqual(a, e) {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
}
