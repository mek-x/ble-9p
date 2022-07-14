package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

type DevicesEntryJson struct {
	LastUpdate  int64 `json:"last"`
	Rssi        int   `json:"rssi"`
	DataEntries int   `json:"avail"`
}

type DeviceFileJson struct {
	Addr       string            `json:"addr"`
	LastUpdate int64             `json:"last"`
	Rssi       int               `json:"rssi"`
	Data       map[string]string `json:"data"`
}

type device struct {
	lastUpdate time.Time
	rssi       int
	data       map[string][]byte
}

var devices map[string]device

var myFs *fs.FS
var dir *fs.StaticDir
var root *fs.StaticDir

type devicesFile struct {
	fs.BaseFile
}

func (f *devicesFile) Read(fid uint64, offset uint64, count uint64) ([]byte, error) {
	devicesData := make(map[string]DevicesEntryJson, len(devices))

	for k, e := range devices {
		var d DevicesEntryJson

		d.DataEntries = len(e.data)
		d.LastUpdate = e.lastUpdate.Unix()
		d.Rssi = e.rssi

		devicesData[k] = d
	}

	out, err := json.MarshalIndent(&devicesData, "", "\t")
	var bs []byte
	if err == nil {
		var length uint64
		if offset > uint64(len(out)) {
			length = 0
		} else {
			length = uint64(len(out)) - offset
		}

		if length < count {
			count = length
		}
		bs = make([]byte, count)

		copy(bs, out[offset:])
	}

	return bs, err
}

func (f *devicesFile) Stat() proto.Stat {
	s := f.BaseFile.Stat()
	s.Length = uint64(len(devices))
	return s
}

func (f *devicesFile) Write(fid uint64, offset uint64, data []byte) (uint32, error) {
	name := string(data)

	e := dir.AddChild(fs.NewDynamicFile(
		myFs.NewStat(name, "glenda", "glenda", 0666),
		func() []byte {
			var devs DeviceFileJson

			devs.Addr = name
			devs.LastUpdate = devices[name].lastUpdate.Unix()
			devs.Rssi = devices[name].rssi
			devs.Data = make(map[string]string)

			for k, v := range devices[name].data {
				devs.Data[k] = hex.EncodeToString(v)
			}
			d, err := json.MarshalIndent(&devs, "", "\t")

			if err != nil {
				return nil
			}
			return d
		},
	))
	if e != nil {
		log.Printf("can't add child: %s\n", e)
	}

	return uint32(len(data)), nil
}

func MyRMFile(Fs *fs.FS, f fs.FSNode) error {
	parent, ok := f.Parent().(fs.ModDir)
	if !ok {
		return fmt.Errorf("%s does not support modification.", fs.FullPath(f.Parent()))
	}
	if f.Parent() == Fs.Root {
		return fmt.Errorf("can't be removed")
	}
	return parent.DeleteChild(f.Stat().Name)
}

func main() {

	go9p.Verbose = true

	myFs, root = fs.NewFS("glenda", "glenda", 0777,
		fs.WithCreateFile(fs.CreateStaticFile),
		fs.WithCreateDir(fs.CreateStaticDir),
		fs.WithRemoveFile(MyRMFile),
		fs.IgnorePermissions(),
		//fs.WithAuth(fs.Plan9Auth),
		//fs.WithAuth(fs.PlainAuth(map[string]string{
		// 	"kyle": "foo",
		// 	"jake": "bar",
		//})),
	)

	root.AddChild(&devicesFile{
		*fs.NewBaseFile(myFs.NewStat("devices", "glenda", "glenda", 0666)),
	})

	dir = fs.NewStaticDir(myFs.NewStat("devs", "glenda", "glenda", 0777))
	root.AddChild(dir)

	devices = make(map[string]device)

	StartBleScan()

	log.Println(go9p.Serve("0.0.0.0:5640", myFs.Server()))
}

func UpdateDevice(t time.Time, addr string, rssi int, data map[string][]byte) {

	d := device{
		lastUpdate: t,
		rssi:       rssi,
		data:       data,
	}

	devices[addr] = d
}
