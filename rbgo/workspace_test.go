package rbgo

import "testing"

func TestPackageRootFinder_Find(t *testing.T) {
	w, _ := NewWorkspace(".")
	if a, e := w.PackageRoot.Find("golang.org/x/crypto/ssh"), "golang.org/x/crypto"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := w.PackageRoot.Find("golang.org/x/crypto"), "golang.org/x/crypto"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := w.PackageRoot.Find("github.com/golang/crypto/ssh"), "github.com/golang/crypto"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := w.PackageRoot.Find("github.com/golang/crypto"), "github.com/golang/crypto"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := w.PackageRoot.Find("piyo"), ""; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
}

func TestExcludeDirs_Contains(t *testing.T) {
	w, _ := NewWorkspace(".")
	if a, e := w.ExcludeDirs.Contains("foo/bar/.git/config"), true; a != e {
		err := "err"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := w.ExcludeDirs.Contains("foo/bar/.git"), true; a != e {
		err := "err"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := w.ExcludeDirs.Contains(".git/config"), true; a != e {
		err := "err"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	if a, e := w.ExcludeDirs.Contains("foo/bar/git"), false; a != e {
		err := "err"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
}

func TestNewWorkDir(t *testing.T) {
	w, _ := NewWorkspace(".")
	if a, e := w.sourceEntry, "."; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	//if a, e := w.objectPath, "pkg"; a != e {
	//	err := "mismatch"
	//	t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	//}
}

func TestNewWorkDir_NotFound(t *testing.T) {
	_, err := NewWorkspace("piyo")
	if err == nil {
		t.Error("no error")
	}
}

func TestNewWorkDir_StandardLayout(t *testing.T) {
	w, _ := NewWorkspace("../example")
	if a, e := w.sourceEntry, "../example/src"; a != e {
		err := "mismatch"
		t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	}
	//if a, e := w.objectPath, "../example/pkg"; a != e {
	//	err := "mismatch"
	//	t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
	//}
}
//
//func TestWorkDir_Init(t *testing.T) {
//	w, _ := NewWorkDir("../example")
//	w.Init()
//	pkgs := make(map[string]*Package)
//	for _, pkg := range w.Package.All() {
//		pkgs[pkg.FullName] = pkg
//	}
//	if pkg, found := pkgs["example"]; !found {
//		t.Error("pkg not found")
//	} else {
//		if a, e := pkg.FullName, "example"; a != e {
//			err := "mismatch"
//			t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
//		}
//		if a, e := pkg.Name, "example"; a != e {
//			err := "mismatch"
//			t.Errorf("%s\nactual: %v\nexpect: %v", err, a, e)
//		}
//	}
//}
