package modules

type RdbOptions struct {
	Addr     string // Redis 服务器地址
	Password string // Redis 服务器密码
	DB       int    // Redis 数据库编号
}

type Rdb struct {
	Base
}

func NewRdb(opts *RdbOptions) *Rdb {
	return &Rdb{}
}

func (r *Rdb) Name() string {
	return "rdb"
}

func (r *Rdb) Init() {
	// r.printInfo()
}

func (r *Rdb) Start() error {
	return nil
}

func (r *Rdb) Close() {}
