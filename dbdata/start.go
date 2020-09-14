package dbdata

func Start() {
	initDb()
}

func Stop() error {
	return db.Close()
}
