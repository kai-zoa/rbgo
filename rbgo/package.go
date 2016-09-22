package rbgo

import (
	"time"
	"io/ioutil"
	"path/filepath"
	"strings"
	"go/token"
	"go/parser"
	"fmt"
	"go/ast"
	"regexp"
	"os"
	"runtime"
	"errors"
)

var (
	SourceNotFound = errors.New("SourceNotFound")
	PackageDirName = ""
)

func init() {
	goOS, goArch := runtime.GOOS, runtime.GOARCH
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if pair[0] == "GOOS" {
			goOS = pair[1]
		}
		if pair[1] == "GOARCH" {
			goArch = pair[1]
		}
	}
	PackageDirName = filepath.Join("pkg", fmt.Sprintf("%s_%s", goOS, goArch))
}

func IsGoSource(path string) bool {
	return strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go")
}

type Source struct {
	packageName string
	modTime     time.Time
	imports     []string
}

func ScanSources(path string) ([]Source, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	name := ""
	sources := make([]Source, 0, len(files))
	for _, fi := range files {
		fpath := filepath.Join(path, fi.Name())
		if !IsGoSource(fpath) {
			continue
		}
		fs := token.NewFileSet()
		astFile, err := parser.ParseFile(fs, fpath, nil, parser.ImportsOnly)
		if err != nil {
			return nil, err
		}
		if astFile.Name.Name == "main" {
			continue
		}
		if name != "" && name != astFile.Name.Name {
			return nil, fmt.Errorf("found multiple packages %s, %s ...", name, astFile.Name.Name)
		}
		src := Source{
			packageName: name,
			modTime: fi.ModTime(),
			imports: []string{},
		}
		sources = append(sources, src)
		name = astFile.Name.Name
		for _, decl := range astFile.Decls {
			if gd, ok := decl.(*ast.GenDecl); ok {
				if gd.Tok == token.IMPORT {
					for _, sp := range gd.Specs {
						if s, ok := sp.(*ast.ImportSpec); ok {
							src.imports = append(src.imports, strings.Trim(s.Path.Value, "\""))
						}
					}
				}
			}
		}
	}
	return sources, nil
}

func NewPackage(sourceRoot, watchPath string) *Package {
	return &Package{
		sourceRoot: sourceRoot,
		WatchPath: watchPath,
		Imports: []string{},
		Referrers: []*Package{},
		MissingImports: []string{},
	}
}

type Package struct {
	Name           string
	FullName       string
	InVendor       bool
	sourceRoot     string
	SourceCount    int
	WatchPath      string
	SourcePath     string
	ObjectPath     string
	WorkDir        string
	ProjectName    string
	ModTime        time.Time
	Imports        []string
	Referrers      []*Package
	MissingImports []string
}

func (p *Package) Scan(f PackageRootFinder) error {
	if p.sourceRoot == "" {
		return errors.New("Root source empty")
	}
	p.Name = ""
	p.FullName = ""
	p.InVendor = false
	p.SourcePath = ""
	p.ObjectPath = ""
	p.WorkDir = ""
	p.ProjectName = ""
	p.SourceCount = 0
	p.ModTime = time.Time{}
	name := ""
	// Scan Sources
	sources, err := ScanSources(p.WatchPath)
	if err != nil {
		return err
	}
	p.SourceCount = len(sources)
	if p.SourceCount == 0 {
		return SourceNotFound
	}
	imports := make([]string, 0, len(p.Imports))
	for _, s := range sources {
		name = s.packageName
		t := s.modTime
		if p.ModTime.Before(t) {
			p.ModTime = t
		}
		imports = append(imports, s.imports...)
	}
	//
	p.Name = name
	p.FullName = name
	p.Imports = imports
	vendorEntry := filepath.Join(p.sourceRoot, "vendor")
	absSourceRoot, _ := filepath.Abs(p.sourceRoot)
	absWatchPath, _ := filepath.Abs(p.WatchPath)
	absVendorPath, _ := filepath.Abs(vendorEntry)
	if strings.HasPrefix(p.WatchPath, vendorEntry) {

		p.FullName = absWatchPath[len(absVendorPath) + 1:]
		p.ProjectName, err = f.VendorProjectName(vendorEntry, p.FullName)
		if err != nil {
			return err
		}
		p.SourcePath = filepath.Join(vendorEntry, filepath.Join(strings.Split(p.ProjectName, "/")...))
		p.InVendor = true

	} else {
		p.FullName = absWatchPath[len(absSourceRoot) + 1:]
		p.SourcePath = p.WatchPath
	}
	pkgDir := PackageDirName
	objectEntry := filepath.Join(p.sourceRoot, pkgDir)
	wd := p.sourceRoot
	if strings.HasSuffix(p.sourceRoot, "src") {
		i := strings.LastIndex(p.sourceRoot, "/src")
		if i != -1 {
			wd = p.sourceRoot[:i]
			objectEntry = filepath.Join(wd, pkgDir)
		} else {
			wd = "."
			objectEntry = pkgDir
		}
	}
	p.WorkDir, _ = filepath.Abs(wd)
	if p.InVendor {
		objectEntry = filepath.Join(objectEntry, "vendor")
		p.ObjectPath = filepath.Join(objectEntry, filepath.Join(strings.Split(p.ProjectName, "/")...))
	} else {
		p.ObjectPath = filepath.Join(objectEntry, filepath.Join(strings.Split(p.FullName, "/")...))
	}
	p.ObjectPath = fmt.Sprintf("%s.a", p.ObjectPath)
	return nil
}

// PackageRepository
type PackageRepository struct {
	nameToPkg map[string]*Package
	pathToPkg map[string]*Package
	dirToPkgs map[string][]*Package
	extPrj    map[string]*PackageRepository
}

func (r *PackageRepository) Init() *PackageRepository {
	r.nameToPkg = make(map[string]*Package)
	r.pathToPkg = make(map[string]*Package)
	r.dirToPkgs = make(map[string][]*Package)
	r.extPrj = map[string]*PackageRepository{}
	return r
}

func (r *PackageRepository) All() []*Package {
	all := make([]*Package, 0, len(r.pathToPkg))
	for _, pkg := range r.pathToPkg {
		all = append(all, pkg)
	}
	return all
}

func (r *PackageRepository) FindByPath(path string) *Package {
	pkg, found := r.pathToPkg[path]
	if !found {
		return nil
	}
	return pkg
}

func (r *PackageRepository) FindByImportName(imp string) *Package {
	pkg, found := r.nameToPkg[imp]
	if !found {
		return nil
	}
	return pkg
}

func (r *PackageRepository) FindByDir(dir string) []*Package {
	p, found := r.dirToPkgs[dir]
	if found {
		return p
	}
	return []*Package{}
}

func (r *PackageRepository) Put(pkg *Package) {
	dir := filepath.Dir(pkg.WatchPath)
	pkgs, found := r.dirToPkgs[dir]
	if found {
		pkgs = append(pkgs, pkg)
	} else {
		pkgs = []*Package{pkg}
	}
	r.dirToPkgs[dir] = pkgs
	r.pathToPkg[pkg.WatchPath] = pkg
	r.nameToPkg[pkg.FullName] = pkg
}

func (r *PackageRepository) Delete(pkg *Package) {
	dir := filepath.Dir(pkg.WatchPath)
	pkgs, found := r.dirToPkgs[dir]
	if found {
		for i, p := range pkgs {
			if p == pkg {
				r.dirToPkgs[dir] = append(pkgs[:i], pkgs[i + 1:]...)
				break
			}
		}
	}
	delete(r.pathToPkg, pkg.WatchPath)
	delete(r.nameToPkg, pkg.FullName)
}

func (r *PackageRepository) ProjectReferrers(pn string) []*Package {
	pkgs := []*Package{}
	repo, found := r.extPrj[pn]
	if !found {
		return pkgs
	}
	for _, pkg := range repo.All() {
		pkgs = append(pkgs, pkg.Referrers...)
	}
	return pkgs
}

func (r *PackageRepository) UpdateDepends() {
	goPath := []string{filepath.Join(runtime.GOROOT(), "src")}
	//fmt.Printf("%v\n", goPath)
	r.extPrj = make(map[string]*PackageRepository, len(r.extPrj))
	for _, pkg := range r.All() {
		pkg.MissingImports = make([]string, 0, len(pkg.MissingImports))
		for _, imp := range pkg.Imports {
			if imp == "C" || imp == "appengine/cloudsql" {
				continue
			}
			ref := r.FindByImportName(imp)
			if ref != nil {
				ref.Referrers = append(ref.Referrers, pkg)
				continue
			}
			found := false
			for _, path := range goPath {
				list := []string{path}
				list = append(list, strings.Split(imp, "/")...)
				impPath := filepath.Join(list...)
				if _, err := os.Stat(impPath); err == nil {
					found = true
					break
				}
			}
			if !found {
				pkg.MissingImports = append(pkg.MissingImports, imp)
			}
		}
		if pkg.InVendor {
			repo, found := r.extPrj[pkg.ProjectName]
			if !found {
				repo = new(PackageRepository).Init()
			}
			repo.Put(pkg)
			r.extPrj[pkg.ProjectName] = repo
		}
	}
}

// PackageRootFinder
type PackageRootFinder []*regexp.Regexp

func (f PackageRootFinder) Find(packageName string) string {
	for _, rex := range []*regexp.Regexp(f) {
		if rex.MatchString(packageName) {
			return rex.FindAllString(packageName, 1)[0]
		}
	}
	return ""
}

func (f PackageRootFinder) VendorProjectName(vendorEntry, packageName string) (string, error) {
	projectName := f.Find(packageName)
	if projectName != packageName {
		packageRoot := filepath.Join(vendorEntry, projectName)
		sources, err := ScanSources(packageRoot)
		if err != nil {
			return "", err
		}
		if len(sources) == 0 {
			sourceDir :=  filepath.Join(vendorEntry, packageName)
			for {
				nextDir :=  filepath.Dir(sourceDir)
				if nextDir == packageRoot {
					break
				}
				sources, err := ScanSources(nextDir)
				if err != nil {
					return "", err
				}
				if len(sources) == 0 {
					break
				}
				sourceDir = nextDir
			}
			projectName = sourceDir[len(vendorEntry) + 1:]
		}
	}
	return projectName, nil
}
