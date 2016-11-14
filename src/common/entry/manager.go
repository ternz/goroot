package entry

import (
	"sort"
	"sync"
)

type RWLock struct {
	mutex *sync.RWMutex
}

func (r *RWLock) Lock() {
	if r.mutex != nil {
		r.mutex.Lock()
	}
}

func (r *RWLock) Unlock() {
	if r.mutex != nil {
		r.mutex.Unlock()
	}
}

func (r *RWLock) RLock() {
	if r.mutex != nil {
		r.mutex.RLock()
	}
}

func (r *RWLock) RUnlock() {
	if r.mutex != nil {
		r.mutex.RUnlock()
	}
}

func NewRWLock(lock bool) *RWLock {
	if lock {
		return &RWLock{
			mutex: &sync.RWMutex{},
		}
	} else {
		return nil //&RWLock{}
	}
}

type CallBack func(entry EntryInterface) bool
type EntryMapId map[uint64]EntryInterface
type EntryMapName map[string]EntryInterface

// Entry的操作接口
// EntryManagerId只有固定Id标示时用
// EntryManagerName只有固定Name标示时用
// EntryManager有固定Id和Name标示时用
// EntryManagerTempid没有固定标示时用
type EntryManagerId struct {
	rwlock *RWLock
	m      EntryMapId
}

func NewEntryManagerId(lock bool) *EntryManagerId {
	return &EntryManagerId{
		m:      EntryMapId{},
		rwlock: NewRWLock(lock),
	}
}

func (e *EntryManagerId) AddEntry(entry EntryInterface) bool {
	if e.rwlock != nil {
		e.rwlock.Lock()
		defer e.rwlock.Unlock()
	}
	_, exists := e.m[entry.GetId()]
	if !exists {
		e.m[entry.GetId()] = entry
	}

	return !exists
}

func (e *EntryManagerId) RemoveEntry(entry EntryInterface) {
	if entry == nil {
		return
	}
	if e.rwlock != nil {
		e.rwlock.Lock()
		defer e.rwlock.Unlock()
	}
	delete(e.m, entry.GetId())
}

func (e *EntryManagerId) RemoveEntryById(id uint64) {
	if e.rwlock != nil {
		e.rwlock.Lock()
		defer e.rwlock.Unlock()
	}
	delete(e.m, id)
}

func (e *EntryManagerId) RemoveEntryAll() {
	if e.rwlock != nil {
		e.rwlock.Lock()
		defer e.rwlock.Unlock()
	}
	e.m = EntryMapId{}
}

func (e *EntryManagerId) GetEntryById(id uint64) EntryInterface {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	entry := e.m[id]
	return entry
}

func (e *EntryManagerId) ExecEverySorted(callback CallBack) (int, bool) {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	ret := 0
	ok := true
	intList := make([]int, len(e.m))

	for k, _ := range e.m {
		intList[ret] = int(k)
		ret++
	}
	sort.Ints(intList)

	ret = 0
	for i := 0; i < len(intList); i++ {
		ret++
		if !callback(e.m[uint64(intList[i])]) {
			ok = false
			break
		}
	}

	return ret, ok
}

func (e *EntryManagerId) ExecEvery(callback CallBack) (int, bool) {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	ret := 0
	ok := true
	for _, v := range e.m {
		ret++
		if !callback(v) {
			ok = false
			break
		}
	}
	return ret, ok
}

func (e *EntryManagerId) GetSize() int {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	n := len(e.m)
	return n
}

func (e *EntryManagerId) ListAll() int {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	ret := 0
	for _, v := range e.m {
		ret++
		v.Debug("ListAll:%d", ret)
	}
	return ret
}

type EntryManagerName struct {
	rwlock *RWLock
	m      EntryMapName
}

func NewEntryManagerName(lock bool) *EntryManagerName {
	return &EntryManagerName{
		m:      EntryMapName{},
		rwlock: NewRWLock(lock),
	}
}

func (e *EntryManagerName) AddEntry(entry EntryInterface) bool {
	if e.rwlock != nil {
		e.rwlock.Lock()
		defer e.rwlock.Unlock()
	}
	_, exists := e.m[entry.GetName()]
	if !exists {
		e.m[entry.GetName()] = entry
	}
	return !exists
}

func (e *EntryManagerName) RemoveEntry(entry EntryInterface) {
	if entry == nil {
		return
	}
	if e.rwlock != nil {
		e.rwlock.Lock()
		defer e.rwlock.Unlock()
	}
	delete(e.m, entry.GetName())
}

func (e *EntryManagerName) RemoveEntryByName(name string) {
	e.rwlock.Lock()
	delete(e.m, name)
	e.rwlock.Unlock()
}

func (e *EntryManagerName) RemoveEntryAll() {
	if e.rwlock != nil {
		e.rwlock.Lock()
		defer e.rwlock.Unlock()
	}
	e.m = EntryMapName{}
}

func (e *EntryManagerName) GetEntryByName(name string) EntryInterface {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	entry := e.m[name]
	return entry
}

func (e *EntryManagerName) ExecEvery(callback CallBack) (int, bool) {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	ret := 0
	ok := true
	for _, v := range e.m {
		ret++
		if !callback(v) {
			ok = false
			break
		}
	}
	return ret, ok
}

func (e *EntryManagerName) GetSize() int {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	n := len(e.m)
	return n
}

func (e *EntryManagerName) ListAll() int {
	if e.rwlock != nil {
		e.rwlock.RLock()
		defer e.rwlock.RUnlock()
	}
	ret := 0
	for _, v := range e.m {
		ret++
		v.Debug("ListAll:%d", ret)
	}
	return ret
}

type EntryManager struct {
	*EntryManagerId
	*EntryManagerName
	rwlock *RWLock
}

func NewEntryManager(lock bool) *EntryManager {
	return &EntryManager{
		EntryManagerId:   NewEntryManagerId(false),
		EntryManagerName: NewEntryManagerName(false),
		rwlock:           NewRWLock(lock),
	}
}

func (em *EntryManager) AddEntry(entry EntryInterface) bool {
	ok := true
	if em.rwlock != nil {
		em.rwlock.Lock()
		defer em.rwlock.Unlock()
	}
	if em.GetEntryById(entry.GetId()) != nil || em.GetEntryByName(entry.GetName()) != nil {
		ok = false
	} else {
		em.EntryManagerId.AddEntry(entry)
		em.EntryManagerName.AddEntry(entry)
	}
	return ok
}

func (em *EntryManager) RemoveEntry(entry EntryInterface) {
	if em.rwlock != nil {
		em.rwlock.Lock()
		defer em.rwlock.Unlock()
	}
	em.EntryManagerId.RemoveEntry(entry)
	em.EntryManagerName.RemoveEntry(entry)
}

func (em *EntryManager) RemoveEntryById(id uint64) {
	if em.rwlock != nil {
		em.rwlock.Lock()
		defer em.rwlock.Unlock()
	}
	entry := em.EntryManagerId.GetEntryById(id)
	if entry != nil {
		em.EntryManagerId.RemoveEntry(entry)
		em.EntryManagerName.RemoveEntry(entry)
	}
}

func (em *EntryManager) RemoveEntryByName(name string) {
	if em.rwlock != nil {
		em.rwlock.Lock()
		defer em.rwlock.Unlock()
	}
	entry := em.EntryManagerName.GetEntryByName(name)
	if entry != nil {
		em.EntryManagerId.RemoveEntry(entry)
		em.EntryManagerName.RemoveEntry(entry)
	}
}

func (em *EntryManager) RemoveEntryAll() {
	if em.rwlock != nil {
		em.rwlock.Lock()
		em.rwlock.Unlock()
	}
	em.EntryManagerId.RemoveEntryAll()
	em.EntryManagerName.RemoveEntryAll()
}

type EntryManagerTempid struct {
	rwlock *RWLock
	m      EntryMapId
	tempid uint64
}

func NewEntryManagerTempid(lock bool) *EntryManagerTempid {
	return &EntryManagerTempid{
		m:      EntryMapId{},
		rwlock: NewRWLock(lock),
		tempid: 100000,
	}
}

func (et *EntryManagerTempid) AddEntry(entry EntryInterface) bool {
	if et.rwlock != nil {
		et.rwlock.Lock()
		defer et.rwlock.Unlock()
	}
	entry.SetId(et.tempid)
	et.tempid++
	_, exists := et.m[entry.GetId()]
	if !exists {
		et.m[entry.GetId()] = entry
	}
	return !exists
}

func (et *EntryManagerTempid) RemoveEntry(entry EntryInterface) {
	if entry == nil {
		return
	}
	if et.rwlock != nil {
		et.rwlock.Lock()
		defer et.rwlock.Unlock()
	}
	delete(et.m, entry.GetId())
}

func (et *EntryManagerTempid) RemoveEntryById(id uint64) {
	if et.rwlock != nil {
		et.rwlock.Lock()
		defer et.rwlock.Unlock()
	}
	delete(et.m, id)
}

func (et *EntryManagerTempid) RemoveEntryAll() {
	if et.rwlock != nil {
		et.rwlock.Lock()
		defer et.rwlock.Unlock()
	}
	et.m = EntryMapId{}
}

func (et *EntryManagerTempid) GetEntryById(id uint64) EntryInterface {
	if et.rwlock != nil {
		et.rwlock.RLock()
		defer et.rwlock.RUnlock()
	}
	entry := et.m[id]
	return entry
}

func (et *EntryManagerTempid) ExecEveryNolock(callback CallBack) (int, bool) {
	ret := 0
	ok := true
	for _, v := range et.m {
		ret++
		if !callback(v) {
			ok = false
			break
		}
	}
	return ret, ok
}

func (et *EntryManagerTempid) ExecEvery(callback CallBack) (int, bool) {
	if et.rwlock != nil {
		et.rwlock.RLock()
		defer et.rwlock.RUnlock()
	}
	ret := 0
	ok := true
	for _, v := range et.m {
		ret++
		if !callback(v) {
			ok = false
			break
		}
	}
	return ret, ok
}

func (et *EntryManagerTempid) GetSize() int {
	if et.rwlock != nil {
		et.rwlock.RLock()
		defer et.rwlock.RUnlock()
	}
	n := len(et.m)
	return n

}

func (et *EntryManagerTempid) ListAll() int {
	if et.rwlock != nil {
		et.rwlock.RLock()
		defer et.rwlock.RUnlock()
	}
	ret := 0
	for _, v := range et.m {
		ret++
		v.Debug("ListAll:%d", ret)
	}
	return ret
}
