package store

type Handle struct {
	Handle     string `json:"handle"`
	Idx        int    `json:"idx"`
	Type       string `json:"type"`
	Data       string `json:"data"`
	TtlType    int    `json:"ttl_type"`
	Ttl        int    `json:"ttl"`
	Timestamp  int    `json:"timestamp"`
	AdminRead  bool   `json:"admin_read"`
	AdminWrite bool   `json:"admin_write"`
	PubRead    bool   `json:"pub_read"`
	PubWrite   bool   `json:"pub_write"`
}

/*
  `handle` varchar(255) NOT NULL,
  `idx` int(11) NOT NULL,
  `type` blob,
  `data` blob,
  `ttl_type` smallint(6) DEFAULT NULL,
  `ttl` int(11) DEFAULT NULL,
  `timestamp` int(11) DEFAULT NULL,
  `refs` blob,
  `admin_read` tinyint(1) DEFAULT NULL,
  `admin_write` tinyint(1) DEFAULT NULL,
  `pub_read` tinyint(1) DEFAULT NULL,
  `pub_write` tinyint(1) DEFAULT NULL,
*/
