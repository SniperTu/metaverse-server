package utils

import (
	"fmt"
	"math/rand"
	"metaverse/logger"
	"net/http"
	"time"
)

// go运行状态查询端口,返回使用的端口号
func PProf() (rp int) {
	pch := make(chan int)
	pprofPort := 20000
	portBase := pprofPort
	rand.New(rand.NewSource(time.Now().UnixNano()))
	go func() {
		for i := 0; i < 3; i++ {
			fmt.Printf("pprof trying to listen on port:%d\n", pprofPort)
			logger.Infof("pprof trying to listen on port:%d", pprofPort)
			pch <- pprofPort
			if err := http.ListenAndServe(fmt.Sprintf(":%d", pprofPort), nil); err != nil {
				fmt.Printf("pprof start failed. %v\n", err)
				logger.Errorf("pprof start failed. %v", err)
				pprofPort = portBase + 1 + int(rand.Int31n(5000))
				continue
			}
		}
	}()
ploop:
	for {
		select {
		case rp = <-pch:
		case <-time.After(2 * time.Second):
			break ploop
		}
	}
	return
}
