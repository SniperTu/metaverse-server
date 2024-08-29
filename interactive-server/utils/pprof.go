package utils

import (
	"fmt"
	"interactive-server/logger"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func PProf() (rp int) {

	/**
	    * @param:
	    * @return:
	    * @Description: go运行状态查询端口,返回使用的端口号
	    * @auth: daniel(2023/2/3)
	**/
	pch := make(chan int)
	pprofPort := 20000
	portBase := pprofPort
	rand.Seed(time.Now().UnixNano())
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
			//为端口试错等待2s
			break ploop
		}
	}
	return
}
