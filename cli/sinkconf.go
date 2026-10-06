package main

type Config struct {
	Listen    []ConfigListener `json:"listen"`
	SqlConns  []ConfigSqlConn  `json:"sql-conns"`
	Resources []ConfigResource `json:"resources"`
}

type ConfigListener struct {
	Network string `json:"network"`
	Addr    string `json:"addr"`
}

func (conf *Config) Load(lstnr func(ConfigListener), rsrc func(ConfigResource), sqlconn func(ConfigSqlConn)) {
	if lstnr != nil {
		for li := range conf.Listen {
			if conf.Listen[li].Network == "" || conf.Listen[li].Addr == "" {
				continue
			}
			lstnr(conf.Listen[li])
		}
	}
	if sqlconn != nil {
		for si := range conf.SqlConns {
			if conf.SqlConns[si].Name == "" || conf.SqlConns[si].Datasource == "" {
				continue
			}
			sqlconn(conf.SqlConns[si])
		}
	}

	if rsrc != nil {
		for ri := range conf.Resources {
			if conf.Resources[ri].Path == "" {
				continue
			}
			rsrc(conf.Resources[ri])
		}
	}
}

type ConfigSqlConn struct {
	Name       string `json:"name"`
	Datasource string `json:"datasrc"`
}

type ConfigResource struct {
	Path      string `json:"path"`
	LocalRoot string `json:"localroot"`
}
