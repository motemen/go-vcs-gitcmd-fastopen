package gitcmdfastopen

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/tools/godoc/vfs"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/go-vcs/vcs/gitcmd"
)

func init() {
	vcs.RegisterOpener("git", Open)
}

type repository struct {
	gitcmd.Repository
}

type clonedRepo struct {
	root string
}

func Open(dir string) (vcs.Repository, error) {
	return &repository{
		Repository: gitcmd.Repository{Dir: dir},
	}, nil
}

func (repo *repository) FileSystem(at vcs.CommitID) (vfs.FileSystem, error) {
	r, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}

	err = exec.Command("git", "clone", "--share", "--no-checkout", repo.GitRootDir(), r).Run()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "checkout", string(at))
	cmd.Dir = r
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	return clonedRepo{r}, nil
}

func (r clonedRepo) Open(name string) (vfs.ReadSeekCloser, error) {
	return os.Open(filepath.Join(r.root, name))
}

func (r clonedRepo) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(filepath.Join(r.root, path))
}

func (r clonedRepo) Stat(path string) (os.FileInfo, error) {
	return os.Stat(filepath.Join(r.root, path))
}

func (r clonedRepo) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(filepath.Join(r.root, path))
}

func (r clonedRepo) String() string {
	return r.root
}
