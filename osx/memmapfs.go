package osx

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/liamvdv/sharedHome/osx/mem"
)

const chmodBits = os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky // Only a subset of bits are allowed to be changed. Documented under os.Chmod()

type MemMapFs struct {
	mu   sync.RWMutex
	data map[string]*mem.FileData
	init sync.Once
}

func NewMemMapFs() Fs {
	return &MemMapFs{}
}

func (m *MemMapFs) getData() map[string]*mem.FileData {
	m.init.Do(func() {
		m.data = make(map[string]*mem.FileData)
		// Root should always exist, right?
		// TODO: what about windows?
		root := mem.CreateDir(HostPathSeparator)
		mem.SetMode(root, os.ModeDir|0755)
		m.data[HostPathSeparator] = root
	})
	return m.data
}

func (*MemMapFs) Name() string { return "MemMapFS" }

func (m *MemMapFs) Create(name string) (File, error) {
	name = normalizePath(name)
	m.mu.Lock()
	file := mem.CreateFile(name)
	m.getData()[name] = file
	m.registerWithParent(file, 0)
	m.mu.Unlock()
	return mem.NewFileHandle(file), nil
}

func (m *MemMapFs) unRegisterWithParent(fileName string) error {
	f, err := m.lockfreeOpen(fileName)
	if err != nil {
		return err
	}
	parent := m.findParent(f)
	if parent == nil {
		log.Panic("parent of ", f.Name(), " is nil")
	}

	parent.Lock()
	mem.RemoveFromMemDir(parent, f)
	parent.Unlock()
	return nil
}

func (m *MemMapFs) findParent(f *mem.FileData) *mem.FileData {
	pdir, _ := filepath.Split(f.Name())
	pdir = filepath.Clean(pdir)
	pfile, err := m.lockfreeOpen(pdir)
	if err != nil {
		return nil
	}
	return pfile
}

func (m *MemMapFs) registerWithParent(f *mem.FileData, perm os.FileMode) {
	if f == nil {
		return
	}
	parent := m.findParent(f)
	if parent == nil {
		pdir := filepath.Dir(filepath.Clean(f.Name()))
		err := m.lockfreeMkdir(pdir, perm)
		if err != nil {
			//log.Println("Mkdir error:", err)
			return
		}
		parent, err = m.lockfreeOpen(pdir)
		if err != nil {
			//log.Println("Open after Mkdir error:", err)
			return
		}
	}

	parent.Lock()
	mem.InitializeDir(parent)
	mem.AddToMemDir(parent, f)
	parent.Unlock()
}

func (m *MemMapFs) lockfreeMkdir(name string, perm os.FileMode) error {
	name = normalizePath(name)
	x, ok := m.getData()[name]
	if ok {
		// Only return ErrFileExists if it's a file, not a directory.
		i := mem.FileInfo{FileData: x}
		if !i.IsDir() {
			return ErrFileExists
		}
	} else {
		item := mem.CreateDir(name)
		mem.SetMode(item, os.ModeDir|perm)
		m.getData()[name] = item
		m.registerWithParent(item, perm)
	}
	return nil
}

func (m *MemMapFs) Mkdir(name string, perm os.FileMode) error {
	perm &= chmodBits
	name = normalizePath(name)

	m.mu.RLock()
	_, ok := m.getData()[name]
	m.mu.RUnlock()
	if ok {
		return &os.PathError{Op: "mkdir", Path: name, Err: ErrFileExists}
	}

	m.mu.Lock()
	item := mem.CreateDir(name)
	mem.SetMode(item, os.ModeDir|perm)
	m.getData()[name] = item
	m.registerWithParent(item, perm)
	m.mu.Unlock()

	return m.setFileMode(name, perm|os.ModeDir)
}

func (m *MemMapFs) MkdirAll(path string, perm os.FileMode) error {
	err := m.Mkdir(path, perm)
	if err != nil {
		if err.(*os.PathError).Err == ErrFileExists {
			return nil
		}
		return err
	}
	return nil
}

// Handle some relative paths
func normalizePath(path string) string {
	path = filepath.Clean(path)

	switch path {
	case ".":
		return HostPathSeparator
	case "..":
		return HostPathSeparator
	default:
		return path
	}
}

func (m *MemMapFs) Open(name string) (File, error) {
	f, err := m.open(name)
	if f != nil {
		return mem.NewReadOnlyFileHandle(f), err
	}
	return nil, err
}

func (m *MemMapFs) openWrite(name string) (File, error) {
	f, err := m.open(name)
	if f != nil {
		return mem.NewFileHandle(f), err
	}
	return nil, err
}

func (m *MemMapFs) open(name string) (*mem.FileData, error) {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return nil, &os.PathError{Op: "open", Path: name, Err: ErrFileNotFound}
	}
	return f, nil
}

func (m *MemMapFs) lockfreeOpen(name string) (*mem.FileData, error) {
	name = normalizePath(name)
	f, ok := m.getData()[name]
	if ok {
		return f, nil
	} else {
		return nil, ErrFileNotFound
	}
}

func (m *MemMapFs) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
	perm &= chmodBits
	chmod := false
	file, err := m.openWrite(name)
	if err == nil && (flag&os.O_EXCL > 0) {
		return nil, &os.PathError{Op: "open", Path: name, Err: ErrFileExists}
	}
	if os.IsNotExist(err) && (flag&os.O_CREATE > 0) {
		file, err = m.Create(name)
		chmod = true
	}
	if err != nil {
		return nil, err
	}
	if flag == os.O_RDONLY {
		file = mem.NewReadOnlyFileHandle(file.(*mem.File).Data())
	}
	if flag&os.O_APPEND > 0 {
		_, err = file.Seek(0, os.SEEK_END)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	if flag&os.O_TRUNC > 0 && flag&(os.O_RDWR|os.O_WRONLY) > 0 {
		err = file.Truncate(0)
		if err != nil {
			file.Close()
			return nil, err
		}
	}
	if chmod {
		return file, m.setFileMode(name, perm)
	}
	return file, nil
}

func (m *MemMapFs) Remove(name string) error {
	name = normalizePath(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.getData()[name]; ok {
		err := m.unRegisterWithParent(name)
		if err != nil {
			return &os.PathError{Op: "remove", Path: name, Err: err}
		}
		delete(m.getData(), name)
	} else {
		return &os.PathError{Op: "remove", Path: name, Err: os.ErrNotExist}
	}
	return nil
}

func (m *MemMapFs) RemoveAll(path string) error {
	path = normalizePath(path)
	m.mu.Lock()
	m.unRegisterWithParent(path)
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()

	for p := range m.getData() {
		if strings.HasPrefix(p, path) {
			m.mu.RUnlock()
			m.mu.Lock()
			delete(m.getData(), p)
			m.mu.Unlock()
			m.mu.RLock()
		}
	}
	return nil
}

func (m *MemMapFs) Rename(oldname, newname string) error {
	oldname = normalizePath(oldname)
	newname = normalizePath(newname)

	if oldname == newname {
		return nil
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.getData()[oldname]; ok {
		m.mu.RUnlock()
		m.mu.Lock()
		m.unRegisterWithParent(oldname)
		fileData := m.getData()[oldname]
		delete(m.getData(), oldname)
		mem.ChangeFileName(fileData, newname)
		m.getData()[newname] = fileData
		m.registerWithParent(fileData, 0)
		m.mu.Unlock()
		m.mu.RLock()
	} else {
		return &os.PathError{Op: "rename", Path: oldname, Err: ErrFileNotFound}
	}
	return nil
}

func (m *MemMapFs) LstatIfPossible(name string) (os.FileInfo, bool, error) {
	fileInfo, err := m.Stat(name)
	return fileInfo, false, err
}

func (m *MemMapFs) Stat(name string) (os.FileInfo, error) {
	f, err := m.Open(name)
	if err != nil {
		return nil, err
	}
	fi := mem.GetFileInfo(f.(*mem.File).Data())
	return fi, nil
}

func (m *MemMapFs) Chmod(name string, mode os.FileMode) error {
	mode &= chmodBits

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chmod", Path: name, Err: ErrFileNotFound}
	}
	prevOtherBits := mem.GetFileInfo(f).Mode() & ^chmodBits

	mode = prevOtherBits | mode
	return m.setFileMode(name, mode)
}

func (m *MemMapFs) setFileMode(name string, mode os.FileMode) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chmod", Path: name, Err: ErrFileNotFound}
	}

	m.mu.Lock()
	mem.SetMode(f, mode)
	m.mu.Unlock()

	return nil
}

func (m *MemMapFs) Chown(name string, uid, gid int) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chown", Path: name, Err: ErrFileNotFound}
	}

	mem.SetUID(f, uid)
	mem.SetGID(f, gid)

	return nil
}

func (m *MemMapFs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	name = normalizePath(name)

	m.mu.RLock()
	f, ok := m.getData()[name]
	m.mu.RUnlock()
	if !ok {
		return &os.PathError{Op: "chtimes", Path: name, Err: ErrFileNotFound}
	}

	m.mu.Lock()
	mem.SetModTime(f, mtime)
	m.mu.Unlock()

	return nil
}

// inserted from afero/ioutil.go and afero/iofs.go - slightly modified

// dirEntry provides adapter from os.FileInfo to fs.DirEntry
type dirEntry struct {
	fs.FileInfo
}

var _ fs.DirEntry = dirEntry{}

func (d dirEntry) Type() fs.FileMode { return d.FileInfo.Mode().Type() }

func (d dirEntry) Info() (fs.FileInfo, error) { return d.FileInfo, nil }

// readDirFile provides adapter from afero.File to fs.ReadDirFile needed for correct Open
type readDirFile struct {
	File
}

var _ fs.ReadDirFile = readDirFile{}

func (r readDirFile) ReadDir(n int) ([]fs.DirEntry, error) {
	items, err := r.File.Readdir(n)
	if err != nil {
		return nil, err
	}

	ret := make([]fs.DirEntry, len(items))
	for i := range items {
		ret[i] = dirEntry{items[i]}
	}

	return ret, nil
}

// byName implements sort.Interface.
type byName []os.FileInfo

func (f byName) Len() int           { return len(f) }
func (f byName) Less(i, j int) bool { return f[i].Name() < f[j].Name() }
func (f byName) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

func (m *MemMapFs) ReadDir(name string) ([]os.DirEntry, error) {
	f, err := m.Open(name)
	if err != nil {
		return nil, err
	}
	return readDirFile{f}.ReadDir(-1)
}

func (m *MemMapFs) ReadFile(name string) ([]byte, error) {
	f, err := m.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// It's a good but not certain bet that FileInfo will tell us exactly how much to
	// read, so let's try it but be prepared for the answer to be wrong.
	var n int64

	if fi, err := f.Stat(); err == nil {
		// Don't preallocate a huge buffer, just in case.
		if size := fi.Size(); size < 1e9 {
			n = size
		}
	}
	// As initial capacity for readAll, use n + a little extra in case Size is zero,
	// and to avoid another allocation after Read has filled the buffer.  The readAll
	// call will read into its allocated internal buffer cheaply.  If the size was
	// wrong, we'll either waste some space off the end or reallocate as needed, but
	// in the overwhelmingly common case we'll get it just right.
	return readAll(f, n+bytes.MinRead)
}

// readAll reads from r until an error or EOF and returns the data it read
// from the internal buffer allocated with a specified capacity.
func readAll(r io.Reader, capacity int64) (b []byte, err error) {
	buf := bytes.NewBuffer(make([]byte, 0, capacity))
	// If the buffer overflows, we will get bytes.ErrTooLarge.
	// Return that as an error. Any other panic remains.
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if panicErr, ok := e.(error); ok && panicErr == bytes.ErrTooLarge {
			err = panicErr
		} else {
			panic(e)
		}
	}()
	_, err = buf.ReadFrom(r)
	return buf.Bytes(), err
}

// ReadAll reads from r until an error or EOF and returns the data it read.
// A successful call returns err == nil, not err == EOF. Because ReadAll is
// defined to read from src until EOF, it does not treat an EOF from Read
// as an error to be reported.
func ReadAll(r io.Reader) ([]byte, error) {
	return readAll(r, bytes.MinRead)
}

func (m *MemMapFs) WriteFile(name string, data []byte, perm os.FileMode) error {
	f, err := m.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}

// Random number state.
// We generate random temporary file names so that there's a good
// chance the file doesn't exist yet - keeps the number of tries in
// TempFile to a minimum.
var rand uint32
var randmu sync.Mutex

func reseed() uint32 {
	return uint32(time.Now().UnixNano() + int64(os.Getpid()))
}

func nextRandom() string {
	randmu.Lock()
	r := rand
	if r == 0 {
		r = reseed()
	}
	r = r*1664525 + 1013904223 // constants from Numerical Recipes
	rand = r
	randmu.Unlock()
	return strconv.Itoa(int(1e9 + r%1e9))[1:]
}

// CreateTemp creates a new temporary file in the directory dir,
// opens the file for reading and writing, and returns the resulting *os.File.
// The filename is generated by taking pattern and adding a random
// string to the end. If pattern includes a "*", the random string
// replaces the last "*".
// If dir is the empty string, TempFile uses the default directory
// for temporary files (see os.TempDir).
// Multiple programs calling TempFile simultaneously
// will not choose the same file. The caller can use f.Name()
// to find the pathname of the file. It is the caller's responsibility
// to remove the file when no longer needed.
func (m *MemMapFs) CreateTemp(dir, pattern string) (f File, err error) {
	if dir == "" {
		dir = os.TempDir()
	}

	var prefix, suffix string
	if pos := strings.LastIndex(pattern, "*"); pos != -1 {
		prefix, suffix = pattern[:pos], pattern[pos+1:]
	} else {
		prefix = pattern
	}

	nconflict := 0
	for i := 0; i < 10000; i++ {
		name := filepath.Join(dir, prefix+nextRandom()+suffix)
		f, err = m.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
		if os.IsExist(err) {
			if nconflict++; nconflict > 10 {
				randmu.Lock()
				rand = reseed()
				randmu.Unlock()
			}
			continue
		}
		break
	}
	return
}

// end

func (m *MemMapFs) List() {
	for _, x := range m.data {
		y := mem.FileInfo{FileData: x}
		fmt.Println(x.Name(), y.Size())
	}
}
