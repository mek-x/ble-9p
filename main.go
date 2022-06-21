package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/knusbaum/go9p"
	"github.com/knusbaum/go9p/fs"
	"github.com/knusbaum/go9p/proto"
)

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
	bs := make([]byte, 0, count)

	for k, e := range devices {
		line := []byte(fmt.Sprintf("%s (%ddBm): %s - %d entries\n", k, e.rssi, e.lastUpdate.Format(time.StampMilli), len(e.data)))
		if len(line) < int(count) {
			bs = append(bs, line...)
		} else {
			break
		}
		count = count - uint64(len(line))

	}
	return bs[offset:], nil
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
			d := make([]byte, 0)
			d = append(d, fmt.Sprintf("%s: r = %d\n", devices[name].lastUpdate.Format(time.StampMilli), devices[name].rssi)...)

			for k, v := range devices[name].data {
				d = append(d, []byte(fmt.Sprintf("%s: %s\n", k, hex.EncodeToString(v)))...)
			}
			return d
		},
	))
	if e != nil {
		log.Printf("can't add child: %s\n", e)
	}

	return uint32(len(data)), nil
}

func main() {

	go9p.Verbose = true

	myFs, root = fs.NewFS("glenda", "glenda", 0777,
		fs.WithCreateFile(fs.CreateStaticFile),
		fs.WithCreateDir(fs.CreateStaticDir),
		fs.WithRemoveFile(fs.RMFile),
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
