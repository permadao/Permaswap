package router

func (r *Router) runProcess() {
	for {
		select {

		// user operations
		case id := <-r.userUnregister:
			r.userUnregisterProc(id)

		case msg := <-r.userQuery:
			r.userQueryProc(msg)

		case msg := <-r.userSubmit:
			r.userSubmitProc(msg)

		// lp node operations
		case id := <-r.lpInit:
			// before register lp session send salt
			r.lpInitProc(id)

		case msg := <-r.lpRegister:
			r.lpRegisterProc(msg)

		case id := <-r.lpUnregister:
			r.lpUnregisterProc(id)

		case msg := <-r.lpAdd:
			r.lpAddProc(msg)

		case msg := <-r.lpRemove:
			r.lpRemoveProc(msg)

		case msg := <-r.lpSign:
			r.lpSignProc(msg)

		case msg := <-r.lpReject:
			r.lpRejectProc(msg)

		// order
		case order := <-r.orderStatus:
			r.orderStatusProc(order)

		// api
		case accid := <-r.apiGetLpsByAccidReq:
			r.apiGetLpsRes <- r.getLpsByAccidProc(accid)

		case poolID := <-r.apiGetLpsByPoolidReq:
			r.apiGetLpsRes <- r.getLpsByPoolidProc(poolID)

		case poolID := <-r.apiGetPoolReq:
			r.apiGetPoolRes <- r.getPoolProc(poolID)

		//stats
		case <-r.getAllLpsReq:
			r.apiGetLpsRes <- r.getAllLpsProc()

		// nft
		case msg := <-r.NFTOwnerChange:
			r.nftOwnerChangeProc(msg)

		// close
		case <-r.close:
			log.Info("process closed")
			close(r.closed)
			return
		}
	}
}
