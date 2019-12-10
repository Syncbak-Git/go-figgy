package figgy

import (
	"crypto/md5"
	"encoding/base64"
	"reflect"
	"time"
	"unsafe"

	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// Updated is the callback function required for the Watcher interace.
type Updated func()

// Watcher is the interface that wraps the Watch method.
// It will poll for parameter changes at the specified frequency,
// calling updated when changed. Upon change, call Load() or LoadWithParamters()
// and use the new Watcher returned.
type Watcher interface {
	Watch(frequency time.Duration, updated Updated) error
}

// defaultWatcher is the default implementation of Watch.
type defaultWatcher struct {
	ssm  ssmiface.SSMAPI
	v    interface{}
	data interface{}
	hash string
}

// Watch creates a go routine which polls for changes in SSM
func (dw defaultWatcher) Watch(frequency time.Duration, updated Updated) error {
	go func(w defaultWatcher) {
		ticker := time.NewTicker(frequency)
		for {
			<-ticker.C
			LoadWithParameters(w.ssm, w.v, w.data)
			if hash(w.data) != w.hash {
				updated()
				return
			}
			//TODO: check hash
		}
	}(dw)
	return nil
}

// hash returns md5 hash taken from any object
func hash(i interface{}) string {
	v := reflect.ValueOf(i)

	size := unsafe.Sizeof(v.Interface())
	b := (*[1 << 10]uint8)(unsafe.Pointer(v.Pointer()))[:size:size]

	h := md5.New()
	return base64.StdEncoding.EncodeToString(h.Sum(b))
}
