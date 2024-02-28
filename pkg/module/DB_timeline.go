package module

// /	#############################################################################################	///
type _historyFall_dbTimeline struct {
	globalObj *localSQLiteObj
}

/*	Получение точки истории по ID	*/
func (obj *_historyFall_dbTimeline) Get(id uint32) (_historyFall_dbTimeline, bool) {
	retObj := _historyFall_dbTimeline{}
	status := true

	return retObj, status
}
