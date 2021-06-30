package dbdata

func Start() {
	initDb()
	initData()

}

func Stop() error {
	return x.Close()
}
