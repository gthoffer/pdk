package leveldb

import (
	"bytes"
	"io/ioutil"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"testing"

	"github.com/pilosa/pdk"
	"github.com/pilosa/pdk/test"
)

func TestTranslator(t *testing.T) {
	levelDir := tempDirName(t)
	bt, err := NewTranslator(levelDir, "f1", "f2")
	if err != nil {
		t.Fatalf("couldn't get level translator: %v", err)
	}
	id1, err := bt.GetID("f1", []byte("hello"))
	if err != nil {
		t.Fatalf("couldn't get id for hello f1: %v", err)
	}
	id2, err := bt.GetID("f2", []byte("hello"))
	if err != nil {
		t.Fatalf("couldn't get id for hello in f2: %v", err)
	}
	id3, err := bt.GetID("fnew", []byte("hello"))
	if err != nil {
		t.Fatalf("couldn't get id for hello in fnew: %v", err)
	}

	val, err := bt.Get("f1", id1)
	test.ErrNil(t, err, "Get(f1, id1)")
	if !bytes.Equal([]byte(val.(pdk.S)), []byte("hello")) {
		t.Fatalf("unexpected value for hello id in f1: %s", val)
	}

	val, err = bt.Get("f2", id2)
	test.ErrNil(t, err, `Get("f2", id2)`)
	if !bytes.Equal([]byte(val.(pdk.S)), []byte("hello")) {
		t.Fatalf("unexpected value for hello id in f2: %s", val)
	}

	val, err = bt.Get("fnew", id3)
	test.ErrNil(t, err, `Get("fnew", id3)`)
	if !bytes.Equal([]byte(val.(pdk.S)), []byte("hello")) {
		t.Fatalf("unexpected value for hello id in fnew: %s", val)
	}

	err = bt.Close()
	if err != nil {
		t.Fatalf("closing level translator: %v", err)
	}

	bt, err = NewTranslator(levelDir, "f1", "f2")
	if err != nil {
		t.Fatalf("couldn't get level translator after closing: %v", err)
	}
	val, err = bt.Get("f1", id1)
	test.ErrNil(t, err, `Get("f1", id1)`)
	if !bytes.Equal([]byte(val.(pdk.S)), []byte("hello")) {
		t.Fatalf("after reopen, unexpected value for hello id in f1: %s", val)
	}

	val, err = bt.Get("f2", id2)
	test.ErrNil(t, err, `Get("f2", id2)`)
	if !bytes.Equal([]byte(val.(pdk.S)), []byte("hello")) {
		t.Fatalf("after reopen, unexpected value for hello id in f2: %s", val)
	}

	val, err = bt.Get("fnew", id3)
	test.ErrNil(t, err, `Get("fnew", id3)`)
	if !bytes.Equal([]byte(val.(pdk.S)), []byte("hello")) {
		t.Fatalf("after reopen, unexpected value for hello id in fnew: %s", val)
	}

	id1again, err := bt.GetID("f1", []byte("hello"))
	if err != nil {
		t.Fatalf("couldn't get id again for hello f1: %v", err)
	}
	id2again, err := bt.GetID("f2", []byte("hello"))
	if err != nil {
		t.Fatalf("couldn't get id again for hello in f2: %v", err)
	}

	id3again, err := bt.GetID("fnew", []byte("hello"))
	if err != nil {
		t.Fatalf("couldn't get id again for hello in fnew: %v", err)
	}

	if id1again != id1 || id2again != id2 || id3 != id3again {
		t.Fatalf("didn't get same ids for same values id1: %v, 1again: %v, 2: %v, 2again: %v, 3: %v 3again: %v", id1, id1again, id2, id2again, id3, id3again)
	}
}

func TestConcTranslator(t *testing.T) {
	levelDir := tempDirName(t)
	bt, err := NewTranslator(levelDir, "f1", "f2")
	if err != nil {
		t.Fatalf("couldn't get level translator: %v", err)
	}

	wg := &sync.WaitGroup{}
	rets := make([][]uint64, 8)
	for i := 0; i < 8; i++ {
		rets[i] = make([]uint64, 1000)
		wg.Add(1)
		go func(ret []uint64) {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				id, err := bt.GetID("f1", []byte(strconv.Itoa(j)))
				if err != nil {
					t.Fatalf("error getting id: %v", err)
				}
				ret[j] = id
			}
		}(rets[i])
	}

	wg.Wait()
	for i, ret := range rets {
		if i != 0 {
			if !reflect.DeepEqual(ret, rets[i-1]) {
				t.Fatalf("returned ids different in different threads: %v, %v", ret, rets[i-1])
			}
		}
		sort.Sort(test.Uint64Slice(ret))
		for j := 0; j < 1000; j++ {
			if ret[j] != uint64(j) {
				t.Fatalf("returned ids are not monotonic, pos: %v, val: %v, arr: %v", j, ret[j], ret)
			}
		}
	}
}

func BenchmarkTranslatorGetID(b *testing.B) {
	levelDir := tempDirName(b)
	bt, err := NewTranslator(levelDir, "f1", "f2")
	if err != nil {
		b.Fatalf("couldn't get level translator: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bt.GetID("f1", []byte(strconv.Itoa(i)))
	}
}

func tempDirName(t testing.TB) string {
	tf, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("couldn't get temp file: %v", err)
	}
	return tf
}