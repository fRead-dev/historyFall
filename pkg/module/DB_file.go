package module

// /	#############################################################################################	///
type _historyFall_dbFile struct {
	globalObj *localSQLiteObj

	buf map[string]uint32 //	Буфер для словаря активных файлов
}

// searchKey	Поиск ключа по буферу
func (obj *_historyFall_dbFile) searchKey(key *string) (uint32, bool) {
	if obj.buf == nil {
		return 0, false
	}

	value, status := obj.buf[*key]
	if !status {
		return 0, false
	}

	return value, true
}
