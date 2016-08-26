package navigation

import (
	"fmt"
	"testing"
)

func TestNavi(t *testing.T) {
	ret := CreateNavitation(1, "C:/home/flynet_rpc/assets/navigation", "srv_test.navmesh", NAVIGATION_NAVMESH)
	fmt.Println("create result:", ret)
	ret = CreateNavitation(2, "C:/home/flynet_rpc/assets/navigation", "srv_test.navmesh", NAVIGATION_NAVMESH)
	fmt.Println("create result:", ret)
	ret = CreateNavitation(3, "C:/home/flynet_rpc/server/src/vendor/navigation", "example.tmx", NAVIGATION_TILE)
	fmt.Println("create result:", ret)
	paths := FindPath(1, 0, -20.52549, 36.5, -0.12, 5.23, 26.009, -0.12)
	fmt.Println(paths)
	paths = FindPath(2, 0, -20.52549, 36.5, -0.12, 5.23, 26.009, -0.12)
	fmt.Println(paths)
	pos := Raycast(1, 0, -20.52549, 11.13, -0.12, -1, 11.13, -15.5)
	fmt.Println(pos)
	paths = FindPath(3, 0, 1, 1, 1, 64, 32, 1)
	fmt.Println(paths)
	Cleanup()
}
