package navigation

/*
#cgo CPPFLAGS:-I./Navigation/Include/ -I./DebugUtils/Include/ -I./Detour/Include/ -I./DetourCrowd/Include/ -I./DetourTileCache/Include/ -I./Recast/Include/ -I./tmxparser/base64/ -I./tmxparser/tinyxml/ -I./tmxparser/zlib/ -I./tmxparser/
#cgo LDFLAGS:-LC:/home/flynet_rpc/server/src/vendor/navigation/lib  -lnavi
#include "navigation.h"
*/
import "C"

import (
	"fmt"
	"reflect"
	"unsafe"
)

const (
	NAVIGATION_TILE    = 1
	NAVIGATION_NAVMESH = 2
)

func init() {
	rt := C.InitNavigation()
	fmt.Println("init navigation, ret: ", rt)
}

func Cleanup() {
	C.DestroyNavigation()
	fmt.Println("navigation cleanup")
}

func CreateNavitation(mapid int, path string, file string, maptyp int) int {
	cpath := C.CString(path)
	cfile := C.CString(file)
	defer C.Free(unsafe.Pointer(cpath))
	defer C.Free(unsafe.Pointer(cfile))
	res := C.CreateNavigation(C.int(mapid), cpath, cfile, C.int(maptyp))
	return int(res)
}

func FindPath(mapid, layer int, start_x, start_y, start_z, end_x, end_y, end_z float32) []float32 {
	cpaths := C.FindStraightPath(C.int(mapid), C.int(layer), C.float(start_x), C.float(start_y), C.float(start_z), C.float(end_x), C.float(end_y), C.float(end_z))
	//cpaths的第一位是数组的长度
	defer C.FreePaths(cpaths)
	length := int(C.GetPathArrSize(cpaths))
	//获取的长度为数据长度，整个数组长度为length+1
	if length < 3 {
		return nil
	}

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cpaths)),
		Len:  length + 1,
		Cap:  length + 1,
	}
	tmppaths := *(*[]C.float)(unsafe.Pointer(&hdr))
	gopaths := make([]float32, length)
	for k, v := range tmppaths[1:] {
		gopaths[k] = float32(v)
	}

	return gopaths
}

func Raycast(mapid, layer int, start_x, start_y, start_z, end_x, end_y, end_z float32) []float32 {
	cpaths := C.Raycast(C.int(mapid), C.int(layer), C.float(start_x), C.float(start_y), C.float(start_z), C.float(end_x), C.float(end_y), C.float(end_z))
	defer C.FreePaths(cpaths)
	if cpaths == nil {
		return nil
	}

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cpaths)),
		Len:  3,
		Cap:  3,
	}
	tmppaths := *(*[]C.float)(unsafe.Pointer(&hdr))

	pos := make([]float32, 3)
	pos[0] = float32(tmppaths[0])
	pos[1] = float32(tmppaths[1])
	pos[2] = float32(tmppaths[2])
	return pos
}

func GetHeight(mapid, layer int, x, y, z float32) float32 {
	h := float32(C.GetHeight(C.int(mapid), C.int(layer), C.float(x), C.float(y), C.float(z)))
	return h
}
